package handlers

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/a-h/templ"
	"github.com/ankit-lilly/go-datastar-daisyui-template/internal/jobs"
	"github.com/ankit-lilly/go-datastar-daisyui-template/internal/views"
	"github.com/starfederation/datastar-go/datastar"
)

type Handlers struct {
	logger  *slog.Logger
	jobHub  *jobs.Hub
	counter atomic.Int64
}

func New(logger *slog.Logger, jobHub *jobs.Hub) *Handlers {
	return &Handlers{
		logger: logger,
		jobHub: jobHub,
	}
}

// Index serves the main page
func (h *Handlers) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if err := views.IndexPage().Render(r.Context(), w); err != nil {
		h.logger.Error("template render error", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Counter handles SSE streaming for counter updates
func (h *Handlers) Counter(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	// Send initial counter value using templ component
	count := h.counter.Load()
	html := renderComponent(r.Context(), views.CounterValue(count))
	sse.PatchElements(html)
}

// Increment handles counter increment via POST
func (h *Handlers) Increment(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	count := h.counter.Add(1)
	html := renderComponent(r.Context(), views.CounterValue(count))
	sse.PatchElements(html)
}

// StartJob starts a background job and streams progress
func (h *Handlers) StartJob(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	// Create a new job
	job := h.jobHub.NewJob("demo-task", func(j *jobs.Job) error {
		// Simulate long-running work
		for i := 0; i <= 100; i += 10 {
			select {
			case <-j.Context().Done():
				return j.Context().Err()
			default:
				j.SetProgress(i)
				time.Sleep(500 * time.Millisecond)
			}
		}
		return nil
	})

	// Submit job to hub
	h.jobHub.Submit(job)

	// Send initial state
	sse.PatchSignals([]byte(fmt.Sprintf(`{"jobId": "%s", "jobStatus": "running", "jobProgress": 0}`, job.ID)))
	html := renderComponent(r.Context(), views.JobInfo(job.ID, "alert-info", "Job started"))
	sse.PatchElements(html)

	// Stream job progress
	for update := range job.Updates() {
		sse.PatchSignals([]byte(fmt.Sprintf(`{"jobProgress": %d}`, update.Progress)))

		if update.Done {
			status := "completed"
			alertClass := "alert-success"
			message := "Job completed!"
			if update.Error != nil {
				status = "failed"
				alertClass = "alert-error"
				message = "Job failed: " + update.Error.Error()
			}
			infoHTML := renderComponent(r.Context(), views.JobInfo(job.ID, alertClass, message))
			sse.PatchElements(infoHTML)
			sse.PatchSignals([]byte(fmt.Sprintf(`{"jobStatus": "%s"}`, status)))
			break
		}
	}
}

// renderComponent renders a templ component to a string
func renderComponent(ctx context.Context, component templ.Component) string {
	var buf bytes.Buffer
	if err := component.Render(ctx, &buf); err != nil {
		return ""
	}
	return buf.String()
}
