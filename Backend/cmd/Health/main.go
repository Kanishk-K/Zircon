package main

import (
	"log"
	"os"
	"time"

	apiresponse "github.com/Kanishk-K/UniteDownloader/Backend/pkg/apiResponse"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/hibiken/asynq"
)

type HealthService struct {
	jobInspector *asynq.Inspector
}

type FilteredServerInfo struct {
	ID      string    `json:"id"`
	Started time.Time `json:"started"`
}

type QueueInfo struct {
	Processed int `json:"processed"`
}

type HealthResponse struct {
	ServerInfo []FilteredServerInfo `json:"serverInfo"`
	QueueInfo  map[string]QueueInfo `json:"queueInfo,omitempty"`
}

func NewHealthService(jobInspector *asynq.Inspector) *HealthService {
	return &HealthService{jobInspector: jobInspector}
}

func (hs *HealthService) handler() (events.APIGatewayProxyResponse, error) {
	resp := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type,Authorization",
		},
		IsBase64Encoded: false,
	}
	var healthResponse HealthResponse
	// Get the server info
	servers, err := hs.jobInspector.Servers()
	if err != nil {
		log.Printf("Error getting server info: %v", err)
		apiresponse.APIErrorResponse(500, "Internal Server Error", &resp)
		return resp, nil
	}
	for _, server := range servers {
		healthResponse.ServerInfo = append(healthResponse.ServerInfo, FilteredServerInfo{
			ID:      server.ID,
			Started: server.Started,
		})
	}
	// Get all queues
	queues, err := hs.jobInspector.Queues()
	if err != nil {
		log.Printf("Error getting queues: %v", err)
		apiresponse.APIErrorResponse(500, "Internal Server Error", &resp)
		return resp, nil
	}
	// Get the queue info
	if len(queues) > 0 {
		healthResponse.QueueInfo = make(map[string]QueueInfo)
		for _, queue := range queues {
			// Get the queue stats
			stats, err := hs.jobInspector.GetQueueInfo(queue)
			if err != nil {
				log.Printf("Error getting queue stats: %v", err)
				apiresponse.APIErrorResponse(500, "Internal Server Error", &resp)
				return resp, nil
			}
			// Get the queue info
			healthResponse.QueueInfo[queue] = QueueInfo{
				Processed: stats.ProcessedTotal,
			}
		}
	}
	apiresponse.APISuccessResponse(healthResponse, &resp)
	return resp, nil
}

func main() {
	// Initialize the service
	client := asynq.NewInspector(asynq.RedisClientOpt{Addr: os.Getenv("REDIS_URL")})
	if client == nil {
		log.Printf("Could not connect to Redis")
		return
	}
	defer client.Close()
	qs := HealthService{jobInspector: client}
	lambda.Start(qs.handler)
}
