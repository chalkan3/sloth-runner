package ui

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

// DBHandler is a slog.Handler that writes log records to an SQLite database.
type DBHandler struct {
	db    *sql.DB
	runID int64
}

// NewDBHandler creates a new DBHandler.
func NewDBHandler(db *sql.DB, runID int64) *DBHandler {
	return &DBHandler{db: db, runID: runID}
}

func (h *DBHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *DBHandler) Handle(ctx context.Context, r slog.Record) error {
	// For now, we only care about the message. We can add attrs later if needed.
	// We also try to extract the task name from the attributes.
	taskName := "pipeline" // Default if no task name is found
	r.Attrs(func(attr slog.Attr) bool {
		if attr.Key == "task" {
			taskName = attr.Value.String()
		}
		return true
	})

	_, err := h.db.Exec("INSERT INTO task_logs (run_id, task_name, timestamp, message) VALUES (?, ?, ?, ?)",
		h.runID, taskName, r.Time, r.Message)
	return err
}

func (h *DBHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Not implemented for this simple handler
	return h
}

func (h *DBHandler) WithGroup(name string) slog.Handler {
	// Not implemented for this simple handler
	return h
}

// CreateRunEntry creates a new record for a pipeline run and returns its ID.
func CreateRunEntry(db *sql.DB, groupName string) (int64, error) {
	res, err := db.Exec("INSERT INTO pipeline_runs (group_name, status, start_time) VALUES (?, ?, ?)",
		groupName, "running", time.Now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateRunStatus updates the status and end time of a pipeline run.
func UpdateRunStatus(db *sql.DB, runID int64, status string) error {
	_, err := db.Exec("UPDATE pipeline_runs SET status = ?, end_time = ? WHERE id = ?",
		status, time.Now(), runID)
	return err
}
