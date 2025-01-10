package router

import (
	"encoding/json"
	"net/http"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/models"
	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/services"
	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/shared/authutil"
)

// This allows us to call service methods in handlers.
type JobSchedulerRouter struct {
	service   services.JobSchedulerServiceMethods
	jwtClient authutil.AuthClientMethods
}

// Creates a new JobSchedulerRouter which is required to have
// the methods defined in JobSchedulerServiceMethods
// It is provided a fully initialized JobSchedulerServiceMethods
func NewJobSchedulerRouter(jss services.JobSchedulerServiceMethods, jwtClient authutil.AuthClientMethods) *JobSchedulerRouter {
	return &JobSchedulerRouter{
		service:   jss,
		jwtClient: jwtClient,
	}
}

// Registers the routes for the JobSchedulerRouter
func (jsr *JobSchedulerRouter) RegisterRoutes() {
	http.HandleFunc("/process", jsr.HandleIncomingJob)
	http.HandleFunc("/status", jsr.StatusCheck)
	http.HandleFunc("/existing", jsr.HandleExistingJob)
}

// Handles the video download route
func (jsr *JobSchedulerRouter) HandleIncomingJob(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		claims, err := jsr.jwtClient.SecureRoute(w, r)
		if err != nil {
			return
		}
		requestBody := models.JobQueueRequest{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Failed to decode request body", http.StatusBadRequest)
			return
		}
		// Realistically this should not error as that would be caught by SecureRoute.
		requestBody.UserID = claims["sub"].(string)
		// TODO: Validate the contents of the request.
		err = jsr.service.ValidateQuery(&requestBody)
		if err != nil {
			http.Error(w, "Failed to validate request", http.StatusBadRequest)
			return
		}
		// Queue the job
		err = jsr.service.ScheduleJob(&requestBody)
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
	} else {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
	}
}

func (jsr *JobSchedulerRouter) StatusCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		_, err := jsr.jwtClient.SecureRoute(w, r)
		if err != nil {
			return
		}
		requestBody := models.JobStatusRequest{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Failed to decode request body", http.StatusBadRequest)
			return
		}
		response, err := jsr.service.CheckStatus(&requestBody)
		if err != nil {
			http.Error(w, "Failed to check status", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
	}
}

func (jsr *JobSchedulerRouter) HandleExistingJob(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		_, err := jsr.jwtClient.SecureRoute(w, r)
		if err != nil {
			return
		}
		requestBody := models.JobStatusRequest{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Failed to decode request body", http.StatusBadRequest)
			return
		}
		response, err := jsr.service.JobProcessedContent(&requestBody)
		if err != nil {
			http.Error(w, "Failed to check existing job", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
	}
}
