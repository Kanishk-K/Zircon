package router

import (
	"net/http"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/services"
)

// This allows us to call service methods in handlers.
type JobSchedulerRouter struct {
	service services.JobSchedulerServiceMethods
}

// Creates a new JobSchedulerRouter which is required to have
// the methods defined in JobSchedulerServiceMethods
// It is provided a fully initialized JobSchedulerServiceMethods
func NewJobSchedulerRouter(jss services.JobSchedulerServiceMethods) *JobSchedulerRouter {
	return &JobSchedulerRouter{
		service: jss,
	}
}

// Registers the routes for the JobSchedulerRouter
func (jsr *JobSchedulerRouter) RegisterRoutes() {
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			jsr.HandleTest(w, r)
		} else {
			http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		}
	})
}

// Handles the test route
func (jsr *JobSchedulerRouter) HandleTest(w http.ResponseWriter, r *http.Request) {
	// Queue the job
	err := jsr.service.QueueJob()
	if err != nil {
		http.Error(w, "Failed to queue job", http.StatusInternalServerError)
		return
	}
	// Respond with a success message
	if _, err := w.Write([]byte("Job queued successfully")); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}
