package ui

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
)

// PipelineRun defines the structure for a single pipeline run history.
type PipelineRun struct {
	ID        int64     `json:"id"`
	GroupName string    `json:"group_name"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	EndTime   sql.NullTime `json:"end_time"`
}

// TaskLog defines the structure for a single log entry.
type TaskLog struct {
	ID        int64     `json:"id"`
	RunID     int64     `json:"run_id"`
	TaskName  string    `json:"task_name"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

// RunDetails defines the structure for the detailed view of a run.
type RunDetails struct {
	Run  PipelineRun `json:"run"`
	Logs []TaskLog   `json:"logs"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections for simplicity. In production, you'd want to restrict this.
		return true
	},
}

// StartServer initializes and starts the web server.
func StartServer(db *sql.DB) {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/runs", getRunsHandler(db))
		r.Get("/runs/{id}", getRunDetailsHandler(db))
	})

	// WebSocket route
	r.Get("/ws/runs/{id}", websocketHandler(db))

	// Serve the embedded frontend
	// content, err := fs.Sub(embeddedFiles, "ui/dist")
	// if err != nil {
	// 	log.Fatal("failed to get embedded fs subdirectory: ", err)
	// }
	// r.Handle("/*", http.FileServer(http.FS(content)))

	log.Println("Starting Sloth-Runner UI on http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
func getRunsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, group_name, status, start_time, end_time FROM pipeline_runs ORDER BY start_time DESC")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var runs []PipelineRun
		for rows.Next() {
			var run PipelineRun
			if err := rows.Scan(&run.ID, &run.GroupName, &run.Status, &run.StartTime, &run.EndTime); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			runs = append(runs, run)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(runs)
	}
}

func getRunDetailsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		// Fetch run details
		var run PipelineRun
		err := db.QueryRow("SELECT id, group_name, status, start_time, end_time FROM pipeline_runs WHERE id = ?", id).Scan(&run.ID, &run.GroupName, &run.Status, &run.StartTime, &run.EndTime)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// Fetch logs
		rows, err := db.Query("SELECT id, run_id, task_name, timestamp, message FROM task_logs WHERE run_id = ? ORDER BY timestamp ASC", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var logs []TaskLog
		for rows.Next() {
			var logEntry TaskLog
			if err := rows.Scan(&logEntry.ID, &logEntry.RunID, &logEntry.TaskName, &logEntry.Timestamp, &logEntry.Message); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			logs = append(logs, logEntry)
		}

		response := RunDetails{
			Run:  run,
			Logs: logs,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func websocketHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket upgrade error:", err)
			return
		}
		defer conn.Close()

		// TODO: Implement logic to stream logs for a specific run
		log.Println("WebSocket connection established")
	}
}
