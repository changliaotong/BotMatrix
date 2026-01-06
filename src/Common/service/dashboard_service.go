package service

import (
	"BotMatrix/common/models"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"
)

// DashboardService handles the web dashboard and API
type DashboardService struct {
	DB *gorm.DB
}

// NewDashboardService creates a new DashboardService
func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{DB: db}
}

// StartServer starts the HTTP server
func (s *DashboardService) StartServer(port string) {
	mux := http.NewServeMux()

	// API Endpoints
	mux.HandleFunc("/api/stats", s.handleGetStats)
	mux.HandleFunc("/api/tasks", s.handleGetRecentTasks)
	mux.HandleFunc("/api/employees", s.handleGetEmployees)

	// Dashboard Page
	mux.HandleFunc("/", s.handleDashboard)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting Dashboard Server at http://localhost%s", addr)
	// Run server in a goroutine so it doesn't block the main thread if needed,
	// but here we might want it to block if it's the main function.
	// However, usually this is called from main alongside other things.
	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
}

func (s *DashboardService) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Try to locate the template file
	// Assuming running from project root or src/Common/cmd/...
	// We'll try a few paths
	possiblePaths := []string{
		"d:/projects/BotMatrix/src/Common/templates/dashboard.html",
		"../templates/dashboard.html",
		"../../templates/dashboard.html",
		"templates/dashboard.html",
	}

	var tmpl *template.Template
	var err error

	for _, path := range possiblePaths {
		tmpl, err = template.ParseFiles(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		http.Error(w, "Could not load dashboard template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure tmpl is not nil before executing
	if tmpl == nil {
		http.Error(w, "Template is nil after parsing", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)
}

// API Responses
type StatsResponse struct {
	ActiveEmployees int64 `json:"active_employees"`
	PendingTasks    int64 `json:"pending_tasks"`
	CompletedToday  int64 `json:"completed_today"`
	FailedTasks     int64 `json:"failed_tasks"`
}

func (s *DashboardService) handleGetStats(w http.ResponseWriter, r *http.Request) {
	var stats StatsResponse

	// Count Active Employees
	s.DB.Model(&models.DigitalEmployee{}).Where("status = ?", "active").Count(&stats.ActiveEmployees)

	// Count Pending Tasks
	s.DB.Model(&models.DigitalEmployeeTask{}).Where("status = ?", "pending").Count(&stats.PendingTasks)

	// Count Completed Tasks (Today)
	startOfDay := time.Now().Truncate(24 * time.Hour)
	s.DB.Model(&models.DigitalEmployeeTask{}).Where("status = ? AND updated_at >= ?", "completed", startOfDay).Count(&stats.CompletedToday)

	// Count Failed Tasks
	s.DB.Model(&models.DigitalEmployeeTask{}).Where("status = ?", "failed").Count(&stats.FailedTasks)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *DashboardService) handleGetRecentTasks(w http.ResponseWriter, r *http.Request) {
	var tasks []models.DigitalEmployeeTask
	// Fetch recent 20 tasks with Assignee preloaded
	result := s.DB.Preload("Assignee").Order("created_at desc").Limit(20).Find(&tasks)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (s *DashboardService) handleGetEmployees(w http.ResponseWriter, r *http.Request) {
	var employees []models.DigitalEmployee
	s.DB.Order("id asc").Find(&employees)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employees)
}
