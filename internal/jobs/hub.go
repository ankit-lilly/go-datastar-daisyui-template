package jobs

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/ankit-lilly/go-datastar-daisyui-template/internal/util"
)

// JobFunc is the function signature for job work
type JobFunc func(j *Job) error

// JobUpdate represents a job progress update
type JobUpdate struct {
	Progress int
	Done     bool
	Error    error
}

// Job represents a background job
type Job struct {
	ID        string
	Name      string
	Status    string
	Progress  int
	CreatedAt time.Time
	Error     error

	ctx     context.Context
	cancel  context.CancelFunc
	work    JobFunc
	updates chan JobUpdate
	mu      sync.RWMutex
}

// NewJob creates a new job
func newJob(name string, work JobFunc) *Job {
	ctx, cancel := context.WithCancel(context.Background())
	return &Job{
		ID:        util.GenerateID(),
		Name:      name,
		Status:    "pending",
		CreatedAt: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
		work:      work,
		updates:   make(chan JobUpdate, 100),
	}
}

// Context returns the job's context
func (j *Job) Context() context.Context {
	return j.ctx
}

// SetProgress updates the job's progress (0-100)
func (j *Job) SetProgress(p int) {
	j.mu.Lock()
	j.Progress = p
	j.mu.Unlock()

	select {
	case j.updates <- JobUpdate{Progress: p}:
	default:
		// Channel full, skip update
	}
}

// Updates returns a channel for receiving job updates
func (j *Job) Updates() <-chan JobUpdate {
	return j.updates
}

// Cancel cancels the job
func (j *Job) Cancel() {
	j.cancel()
}

// Hub manages background jobs
type Hub struct {
	jobs   map[string]*Job
	submit chan *Job
	done   chan struct{}
	logger *slog.Logger
	mu     sync.RWMutex
}

// NewHub creates a new job hub
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		jobs:   make(map[string]*Job),
		submit: make(chan *Job, 100),
		done:   make(chan struct{}),
		logger: logger,
	}
}

// Run starts the job hub processing loop
func (h *Hub) Run() {
	for {
		select {
		case job := <-h.submit:
			go h.execute(job)
		case <-h.done:
			return
		}
	}
}

// Stop stops the job hub
func (h *Hub) Stop() {
	close(h.done)

	// Cancel all running jobs
	h.mu.RLock()
	for _, job := range h.jobs {
		job.Cancel()
	}
	h.mu.RUnlock()
}

// NewJob creates a new job (factory method for external use)
func (h *Hub) NewJob(name string, work JobFunc) *Job {
	return newJob(name, work)
}

// Submit submits a job for execution
func (h *Hub) Submit(job *Job) {
	h.mu.Lock()
	h.jobs[job.ID] = job
	h.mu.Unlock()

	select {
	case h.submit <- job:
	default:
		h.logger.Warn("job queue full", "job_id", job.ID)
	}
}

// Get retrieves a job by ID
func (h *Hub) Get(id string) (*Job, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	job, ok := h.jobs[id]
	return job, ok
}

func (h *Hub) execute(job *Job) {
	job.mu.Lock()
	job.Status = "running"
	job.mu.Unlock()

	h.logger.Info("job started", "job_id", job.ID, "name", job.Name)

	err := job.work(job)

	job.mu.Lock()
	if err != nil {
		job.Status = "failed"
		job.Error = err
		h.logger.Error("job failed", "job_id", job.ID, "error", err)
	} else {
		job.Status = "completed"
		job.Progress = 100
		h.logger.Info("job completed", "job_id", job.ID)
	}
	job.mu.Unlock()

	// Send final update
	job.updates <- JobUpdate{
		Progress: job.Progress,
		Done:     true,
		Error:    err,
	}
	close(job.updates)
}
