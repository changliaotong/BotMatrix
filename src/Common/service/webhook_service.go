package service

import (
	"BotMatrix/common/ai/employee"
	"BotMatrix/common/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WebhookService handles external webhooks (e.g., GitLab, GitHub)
type WebhookService struct {
	DB          *gorm.DB
	TaskService employee.DigitalEmployeeTaskService
}

// NewWebhookService creates a new WebhookService
func NewWebhookService(db *gorm.DB, taskService employee.DigitalEmployeeTaskService) *WebhookService {
	return &WebhookService{
		DB:          db,
		TaskService: taskService,
	}
}

// StartServer starts the Webhook listener
func (s *WebhookService) StartServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/gitlab", s.handleGitLabWebhook)
	mux.HandleFunc("/webhook/github", s.handleGitHubWebhook)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting Webhook Service at http://localhost%s", addr)
	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatalf("Failed to start webhook server: %v", err)
		}
	}()
}

// Minimal struct for GitLab Pipeline/Job Event
type GitLabWebhookEvent struct {
	ObjectKind       string `json:"object_kind"` // "pipeline", "build"
	ObjectAttributes struct {
		ID             int    `json:"id"`
		Status         string `json:"status"` // "failed", "success"
		Ref            string `json:"ref"`
		SHA            string `json:"sha"`
		Duration       int    `json:"duration"`
		DetailedStatus string `json:"detailed_status"`
	} `json:"object_attributes"`
	Project struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		WebURL     string `json:"web_url"`
		GitHTTPURL string `json:"git_http_url"`
	} `json:"project"`
	Builds []struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"builds"`
}

func (s *WebhookService) handleGitLabWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var event GitLabWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Only care about failed pipelines or builds
	if event.ObjectAttributes.Status != "failed" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ignored non-failed event"))
		return
	}

	log.Printf("Received GitLab Failure Event: %s (ID: %d)", event.ObjectKind, event.ObjectAttributes.ID)

	// Construct Task Description
	description := fmt.Sprintf("Fix CI/CD failure for project %s.\nRepo: %s\nBranch: %s\nCommit: %s\n",
		event.Project.Name, event.Project.GitHTTPURL, event.ObjectAttributes.Ref, event.ObjectAttributes.SHA)

	if len(event.Builds) > 0 {
		description += "Failed Jobs:\n"
		for _, build := range event.Builds {
			if build.Status == "failed" {
				description += fmt.Sprintf("- Job %s (ID: %d) failed\n", build.Name, build.ID)
			}
		}
	}

	// Find a suitable assignee (Developer Role)
	var assignee models.DigitalEmployee
	// Try to find an employee with "Developer" or "Software Engineer" in their Title or Role Template Name
	// This is a simplified logic. In a real system, we might look up a specific team.
	err = s.DB.Joins("JOIN \"DigitalRoleTemplate\" ON \"DigitalEmployee\".\"RoleTemplateId\" = \"DigitalRoleTemplate\".\"Id\"").
		Where("\"DigitalEmployee\".\"Status\" = ? AND (\"DigitalRoleTemplate\".\"Name\" LIKE ? OR \"DigitalEmployee\".\"Title\" LIKE ?)", "active", "%Developer%", "%Engineer%").
		First(&assignee).Error

	assigneeID := uint(0)
	if err == nil {
		assigneeID = assignee.ID
		log.Printf("Assigning CI/CD fix task to: %s (ID: %d)", assignee.Name, assignee.ID)
	} else {
		log.Printf("No specific developer found, task will be unassigned. Error: %v", err)
	}

	// Create Digital Employee Task
	task := &models.DigitalEmployeeTask{
		ExecutionID: uuid.New().String(),
		Title:       fmt.Sprintf("Fix CI Build #%d for %s", event.ObjectAttributes.ID, event.Project.Name),
		Description: description,
		TaskType:    "ai", // Trigger AI Agent
		Status:      "pending",
		Priority:    "high",
		CreatorID:   "system_webhook",
		AssigneeID:  assigneeID,
		Context:     string(body), // Save full webhook payload as context
	}

	if err := s.TaskService.CreateTask(context.Background(), task); err != nil {
		log.Printf("Failed to create repair task: %v", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	// Trigger Execution immediately
	go func() {
		log.Printf("Triggering execution for task %s", task.ExecutionID)
		if err := s.TaskService.ExecuteTask(context.Background(), task.ExecutionID); err != nil {
			log.Printf("Failed to execute task %s: %v", task.ExecutionID, err)
		}
	}()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Task created: %s", task.ExecutionID)))
}

func (s *WebhookService) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	// Placeholder for GitHub
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("GitHub webhook received (not implemented)"))
}
