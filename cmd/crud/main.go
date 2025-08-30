package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/tjarkko/go-demo/cmd/crud/db"

	_ "github.com/lib/pq"
)

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Author  string `json:"author"`
}

type UpdatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Author  string `json:"author"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Server struct {
	db  *db.Queries
	mux *http.ServeMux
}

func NewServer(dbConn *sql.DB) *Server {
	queries := db.New(dbConn)
	mux := http.NewServeMux()

	server := &Server{
		db:  queries,
		mux: mux,
	}

	// Register routes
	server.registerRoutes()

	return server
}

func (s *Server) registerRoutes() {
	// Health check
	s.mux.HandleFunc("GET /health", s.handleHealth)

	// CRUD endpoints
	s.mux.HandleFunc("GET /posts", s.handleListPosts)
	s.mux.HandleFunc("POST /posts", s.handleCreatePost)
	s.mux.HandleFunc("GET /posts/{id}", s.handleGetPost)
	s.mux.HandleFunc("PUT /posts/{id}", s.handleUpdatePost)
	s.mux.HandleFunc("DELETE /posts/{id}", s.handleDeletePost)
	s.mux.HandleFunc("GET /posts/author/{author}", s.handleGetPostsByAuthor)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleListPosts(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // default limit
	offset := 0 // default offset

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	posts, err := s.db.ListPosts(context.Background(), db.ListPostsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		s.writeError(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(posts)
}

func (s *Server) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.Content == "" || req.Author == "" {
		s.writeError(w, "Title, content, and author are required", http.StatusBadRequest)
		return
	}

	post, err := s.db.CreatePost(context.Background(), db.CreatePostParams{
		Title:   req.Title,
		Content: req.Content,
		Author:  req.Author,
	})
	if err != nil {
		s.writeError(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func (s *Server) handleGetPost(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		s.writeError(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := s.db.GetPost(context.Background(), int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			s.writeError(w, "Post not found", http.StatusNotFound)
			return
		}
		s.writeError(w, "Failed to fetch post", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(post)
}

func (s *Server) handleUpdatePost(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		s.writeError(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var req UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.Content == "" || req.Author == "" {
		s.writeError(w, "Title, content, and author are required", http.StatusBadRequest)
		return
	}

	post, err := s.db.UpdatePost(context.Background(), db.UpdatePostParams{
		ID:      int32(id),
		Title:   req.Title,
		Content: req.Content,
		Author:  req.Author,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			s.writeError(w, "Post not found", http.StatusNotFound)
			return
		}
		s.writeError(w, "Failed to update post", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(post)
}

func (s *Server) handleDeletePost(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		s.writeError(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	err = s.db.DeletePost(context.Background(), int32(id))
	if err != nil {
		s.writeError(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGetPostsByAuthor(w http.ResponseWriter, r *http.Request) {
	author := r.PathValue("author")
	if author == "" {
		s.writeError(w, "Author parameter is required", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // default limit
	offset := 0 // default offset

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	posts, err := s.db.GetPostsByAuthor(context.Background(), db.GetPostsByAuthorParams{
		Author: author,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		s.writeError(w, "Failed to fetch posts by author", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(posts)
}

func (s *Server) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func main() {
	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "blogdb")
	dbUser := getEnv("DB_USER", "bloguser")
	dbPassword := getEnv("DB_PASSWORD", "blogpass")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	dbConn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer dbConn.Close()

	// Test database connection
	if err := dbConn.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	server := NewServer(dbConn)

	port := getEnv("PORT", "8080")
	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, server))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
