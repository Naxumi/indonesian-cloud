package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusCompleted TaskStatus = "completed"
)

type Task struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Status    TaskStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

func main() {
	cfgDB, err := Load()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	db, err := NewPostgreSQLDB(&cfgDB.DatabaseConfig)
	if err != nil {
		fmt.Println("Error connecting to DB:", err)
		return
	}

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://*", "https://*"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		MaxAge:         300,
	}))

	logFormat := httplog.SchemaECS.Concise(false)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: logFormat.ReplaceAttr,
	})).With(
		slog.String("app", "mini-task-management-system"),
		slog.String("version", "v1.0.0"),
		slog.String("env", "development"),
	)

	r.Use(httplog.RequestLogger(logger, &httplog.Options{
		Level:           slog.LevelDebug,
		Schema:          httplog.SchemaECS,
		LogRequestBody:  func(req *http.Request) bool { return true },
		LogResponseBody: func(req *http.Request) bool { return true },
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API Task Management Running!"))
	})

	r.Get("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		getAllTasks(w, r, db)
	})
	r.Post("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		createTask(w, r, db)
	})

	// Endpoint PATCH ini sekarang menangani update Status, Title, DAN Note
	r.Patch("/api/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		updateTask(w, r, db)
	})
	r.Delete("/api/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		deleteTask(w, r, db)
	})

	fmt.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", r)
}

func getAllTasks(w http.ResponseWriter, r *http.Request, db *DB) {
	// Filter berdasarkan status (all / pending / completed), kosong = all
	statusFilter := r.URL.Query().Get("status")
	var statusParam TaskStatus
	switch statusFilter {
	case "", "all":
		statusParam = ""
	case "pending":
		statusParam = TaskStatusPending
	case "completed":
		statusParam = TaskStatusCompleted
	default:
		http.Error(w, "Invalid status filter, allowed values are all, pending, completed", http.StatusBadRequest)
		return
	}

	query := "SELECT id, title, status, created_at FROM tasks"
	args := []any{}
	if statusParam != "" {
		query += " WHERE status = $1"
		args = append(args, statusParam)
	}
	query += " ORDER BY created_at DESC"

	rows, err := db.Query(r.Context(), query, args...)
	if err != nil {
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		// Scan juga wajib menyertakan note
		if err := rows.Scan(&t.ID, &t.Title, &t.Status, &t.CreatedAt); err != nil {
			http.Error(w, "Failed to scan row", http.StatusInternalServerError)
			log.Printf("Failed to scan row: %v", err)
			return
		}
		tasks = append(tasks, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func createTask(w http.ResponseWriter, r *http.Request, db *DB) {
	var t Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(t.Title) == "" {
		http.Error(w, "Task title can not be empty", http.StatusBadRequest)
		return
	}

	if len(t.Title) > 255 {
		http.Error(w, "Task title character cannot exceed 255 characters", http.StatusBadRequest)
		return
	}

	if t.Status == "" {
		t.Status = "pending"
	}

	if t.Status != TaskStatusPending && t.Status != TaskStatusCompleted {
		http.Error(w, "Only pending or completed status is allowed", http.StatusBadRequest)
		return
	}

	// Insert menyertakan note
	_, err := db.Exec(r.Context(), "INSERT INTO tasks (title, status) VALUES ($1, $2)",
		t.Title, t.Status)
	if err != nil {
		http.Error(w, "Failed to insert data", http.StatusInternalServerError)
		log.Printf("Failed to insert data: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

// UPDATE (PATCH) yang Dinamis
func updateTask(w http.ResponseWriter, r *http.Request, db *DB) {
	id := chi.URLParam(r, "id")

	_, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "requested ID is not a valid id", http.StatusBadRequest)
		return
	}

	// PATCH yang dinamis: hanya field yang dikirim yang akan di-update
	var payload struct {
		Title  *string `json:"title"`
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if payload.Title == nil && payload.Status == nil {
		http.Error(w, "Nothing to update", http.StatusBadRequest)
		return
	}

	if payload.Title != nil && strings.TrimSpace(*payload.Title) == "" {
		http.Error(w, "Task title can not be empty", http.StatusBadRequest)
		return
	}

	if payload.Status != nil && TaskStatus(*payload.Status) != TaskStatusCompleted && TaskStatus(*payload.Status) != TaskStatusPending {
		http.Error(w, "Only pending or completed status is allowed", http.StatusBadRequest)
		return
	}

	// COALESCE akan menyimpan nilai lama jika nilai baru yang dikirim adalah nil (tidak dikirim)
	_, err = db.Exec(r.Context(), `
		UPDATE tasks
		SET
			title  = COALESCE($1, title),
			status = COALESCE($2, status)
		WHERE id = $3`,
		payload.Title, payload.Status, id)

	if err != nil {
		http.Error(w, "Failed to update data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Task updated successfully"}`))
}

func deleteTask(w http.ResponseWriter, r *http.Request, db *DB) {
	id := chi.URLParam(r, "id")

	_, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "requested ID is not a valid id", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(r.Context(), "DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Failed to delete data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Task deleted successfully"}`))
}
