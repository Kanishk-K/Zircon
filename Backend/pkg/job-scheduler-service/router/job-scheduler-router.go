package router

import (
	"encoding/json"
	"net/http"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/models"
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
	http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			jsr.HandleIncomingJob(w, r)
		} else {
			http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		}
	})
}

// Handles the video download route
func (jsr *JobSchedulerRouter) HandleIncomingJob(w http.ResponseWriter, r *http.Request) {
	requestBody := models.JobInformation{}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	// TODO: Validate the contents of the request.
	// Queue the job
	err := jsr.service.ScheduleJob(&requestBody)
	if err != nil {
		http.Error(w, "Failed to queue job", http.StatusInternalServerError)
		return
	}
	// Respond with a success message in JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Job queued successfully"}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
