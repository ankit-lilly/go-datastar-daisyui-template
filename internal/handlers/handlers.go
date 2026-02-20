package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync/atomic"

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

// Demo serves the demo page
func (h *Handlers) Demo(w http.ResponseWriter, r *http.Request) {
	if err := views.DemoPage().Render(r.Context(), w); err != nil {
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

// StartJob starts a background job and returns its ID
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
				// Simulate work
				select {
				case <-j.Context().Done():
					return j.Context().Err()
				case <-r.Context().Done():
					return r.Context().Err()
				default:
				}
			}
		}
		return nil
	})

	// Submit job to hub
	h.jobHub.Submit(job)

	// Return job ID to client
	sse.PatchSignals([]byte(fmt.Sprintf(`{"jobId": "%s", "jobStatus": "running"}`, job.ID)))
	html := renderComponent(r.Context(), views.JobInfo(job.ID, "alert-info", "Job started: "+job.ID))
	sse.PatchElements(html)
}

// JobStatus streams job progress via SSE
func (h *Handlers) JobStatus(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	if jobID == "" {
		http.Error(w, "job id required", http.StatusBadRequest)
		return
	}

	job, ok := h.jobHub.Get(jobID)
	if !ok {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	sse := datastar.NewSSE(w, r)

	// Stream job progress
	for update := range job.Updates() {
		html := renderComponent(r.Context(), views.JobProgress(update.Progress))
		sse.PatchElements(html)

		if update.Done {
			status := "completed"
			alertClass := "alert-success"
			message := "Job completed"
			if update.Error != nil {
				status = "failed"
				alertClass = "alert-error"
				message = "Job failed: " + update.Error.Error()
			}
			infoHTML := renderComponent(r.Context(), views.JobInfo(jobID, alertClass, message))
			sse.PatchElements(infoHTML)
			sse.PatchSignals([]byte(`{"jobStatus": "` + status + `"}`))
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

// Unused - keeping for reference
func (h *Handlers) renderCounter(count int64) string {
	return fmt.Sprintf(`<span id="counter-value">%d</span>`, count)
}

func (h *Handlers) renderProgress(progress int) string {
	return fmt.Sprintf(`
		<div id="job-progress">
			<progress class="progress progress-primary w-full" value="%d" max="100"></progress>
			<span class="text-sm">%d%%</span>
		</div>
	`, progress, progress)
}

// Ensure io.Writer is imported (for templ.Component interface)
var _ io.Writer = (*bytes.Buffer)(nil)
