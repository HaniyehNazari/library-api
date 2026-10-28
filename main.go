package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	db *pgxpool.Pool
}

type Book struct {
	ID              int     `json:"book_id"`
	Title           string  `json:"title"`
	ISBN            *string `json:"isbn"`
	Price           float64 `json:"price"`
	PublicationYear *int    `json:"publication_year"`
	Description     *string `json:"description"`
}

type CreateBookRequest struct {
	Title           string  `json:"title"`
	ISBN            *string `json:"isbn"`
	Price           float64 `json:"price"`
	PublicationYear *int    `json:"publication_year"`
	Description     *string `json:"description"`
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := a.db.Ping(ctx); err != nil {
		respondJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "database unavailable"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok", "database": "connected"})
}

func (a *App) books(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := a.db.Query(r.Context(), `
			SELECT book_id, title, isbn, price::float8,
			       publication_year, description
			FROM books
			ORDER BY book_id
		`)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not retrieve books"})
			return
		}
		defer rows.Close()

		books := []Book{}
		for rows.Next() {
			var b Book
			if err := rows.Scan(
				&b.ID, &b.Title, &b.ISBN, &b.Price,
				&b.PublicationYear, &b.Description,
			); err != nil {
				respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not read books"})
				return
			}
			books = append(books, b)
		}
		if err := rows.Err(); err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "error reading books"})
			return
		}
		respondJSON(w, http.StatusOK, books)

	case http.MethodPost:
		var req CreateBookRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		if req.Title == "" {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
			return
		}
		if req.Price < 0 {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "price cannot be negative"})
			return
		}

		var b Book
		err := a.db.QueryRow(r.Context(), `
			INSERT INTO books (title, isbn, price, publication_year, description)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING book_id, title, isbn, price::float8,
			          publication_year, description
		`, req.Title, req.ISBN, req.Price, req.PublicationYear, req.Description).Scan(
			&b.ID, &b.Title, &b.ISBN, &b.Price,
			&b.PublicationYear, &b.Description,
		)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "could not create book; check ISBN and field values"})
			return
		}
		respondJSON(w, http.StatusCreated, b)

	default:
		w.Header().Set("Allow", "GET, POST")
		respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (a *App) authors(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := a.db.Query(r.Context(),
			"SELECT author_id, full_name FROM authors ORDER BY author_id")
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not retrieve authors"})
			return
		}
		defer rows.Close()

		result := []map[string]any{}
		for rows.Next() {
			var id int
			var name string
			if err := rows.Scan(&id, &name); err != nil {
				respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not read authors"})
				return
			}
			result = append(result, map[string]any{"author_id": id, "full_name": name})
		}
		respondJSON(w, http.StatusOK, result)

	case http.MethodPost:
		var req struct {
			FullName string `json:"full_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.FullName == "" {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "full_name is required"})
			return
		}

		var id int
		var name string
		err := a.db.QueryRow(r.Context(),
			"INSERT INTO authors (full_name) VALUES ($1) RETURNING author_id, full_name",
			req.FullName,
		).Scan(&id, &name)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "could not create author; name may already exist"})
			return
		}
		respondJSON(w, http.StatusCreated, map[string]any{"author_id": id, "full_name": name})

	default:
		w.Header().Set("Allow", "GET, POST")
		respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (a *App) categories(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := a.db.Query(r.Context(),
			"SELECT category_id, category_name FROM categories ORDER BY category_id")
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not retrieve categories"})
			return
		}
		defer rows.Close()

		result := []map[string]any{}
		for rows.Next() {
			var id int
			var name string
			if err := rows.Scan(&id, &name); err != nil {
				respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not read categories"})
				return
			}
			result = append(result, map[string]any{"category_id": id, "category_name": name})
		}
		respondJSON(w, http.StatusOK, result)

	case http.MethodPost:
		var req struct {
			CategoryName string `json:"category_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CategoryName == "" {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "category_name is required"})
			return
		}

		var id int
		var name string
		err := a.db.QueryRow(r.Context(),
			"INSERT INTO categories (category_name) VALUES ($1) RETURNING category_id, category_name",
			req.CategoryName,
		).Scan(&id, &name)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "could not create category; name may already exist"})
			return
		}
		respondJSON(w, http.StatusCreated, map[string]any{"category_id": id, "category_name": name})

	default:
		w.Header().Set("Allow", "GET, POST")
		respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (a *App) branches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	rows, err := a.db.Query(r.Context(), `
		SELECT branch_id, branch_name, address, phone, is_active
		FROM branches ORDER BY branch_id
	`)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not retrieve branches"})
		return
	}
	defer rows.Close()

	result := []map[string]any{}
	for rows.Next() {
		var id int
		var name, address string
		var phone *string
		var active bool
		if err := rows.Scan(&id, &name, &address, &phone, &active); err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not read branches"})
			return
		}
		result = append(result, map[string]any{
			"branch_id": id, "branch_name": name,
			"address": address, "phone": phone, "is_active": active,
		})
	}
	respondJSON(w, http.StatusOK, result)
}

func main() {
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "library_db"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "library_app"
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		log.Fatal("DB_PASSWORD environment variable is required")
	}

	connURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(dbUser, dbPassword),
		Host:   dbHost + ":" + port,
		Path:   dbName,
	}
	q := connURL.Query()
	q.Set("sslmode", "disable")
	connURL.RawQuery = q.Encode()
	connString := connURL.String()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatal("Database configuration error: ", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("Database connection failed: ", err)
	}

	app := &App{db: pool}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", app.health)
	mux.HandleFunc("/books", app.books)
	mux.HandleFunc("/books/details", app.bookDetails)
	mux.HandleFunc("/books/authors", app.bookAuthors)
	mux.HandleFunc("/books/categories", app.bookCategories)
	mux.HandleFunc("/authors", app.authors)
	mux.HandleFunc("/categories", app.categories)
	mux.HandleFunc("/branches", app.branches)
	mux.HandleFunc("/books/search", app.searchBooks)
	mux.HandleFunc("/books/", app.bookByID)

	serverPort := os.Getenv("APP_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	log.Printf("Library API running at http://localhost:%s", serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, mux))

	_ = strconv.Itoa
}

func (a *App) bookAuthors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		BookID   int `json:"book_id"`
		AuthorID int `json:"author_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil ||
		req.BookID <= 0 || req.AuthorID <= 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "valid book_id and author_id are required",
		})
		return
	}

	_, err := a.db.Exec(r.Context(),
		"INSERT INTO book_authors (book_id, author_id) VALUES ($1, $2)",
		req.BookID, req.AuthorID,
	)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "could not link book and author; check IDs or existing link",
		})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"message":   "book linked to author",
		"book_id":   req.BookID,
		"author_id": req.AuthorID,
	})
}

func (a *App) bookCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		BookID     int `json:"book_id"`
		CategoryID int `json:"category_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil ||
		req.BookID <= 0 || req.CategoryID <= 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "valid book_id and category_id are required",
		})
		return
	}

	_, err := a.db.Exec(r.Context(),
		"INSERT INTO book_categories (book_id, category_id) VALUES ($1, $2)",
		req.BookID, req.CategoryID,
	)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "could not link book and category; check IDs or existing link",
		})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"message":     "book linked to category",
		"book_id":     req.BookID,
		"category_id": req.CategoryID,
	})
}

func (a *App) bookDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		respondJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	rows, err := a.db.Query(r.Context(), `
		SELECT
			b.book_id,
			b.title,
			b.isbn,
			b.price::float8,
			b.publication_year,
			b.description,
			COALESCE(
				(SELECT json_agg(a.full_name ORDER BY a.full_name)
				 FROM book_authors ba
				 JOIN authors a ON a.author_id = ba.author_id
				 WHERE ba.book_id = b.book_id),
				'[]'::json
			) AS authors,
			COALESCE(
				(SELECT json_agg(c.category_name ORDER BY c.category_name)
				 FROM book_categories bc
				 JOIN categories c ON c.category_id = bc.category_id
				 WHERE bc.book_id = b.book_id),
				'[]'::json
			) AS categories
		FROM books b
		ORDER BY b.book_id
	`)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "could not retrieve book details",
		})
		return
	}
	defer rows.Close()

	type BookDetails struct {
		Book
		Authors    json.RawMessage `json:"authors"`
		Categories json.RawMessage `json:"categories"`
	}

	result := []BookDetails{}
	for rows.Next() {
		var item BookDetails
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.ISBN,
			&item.Price,
			&item.PublicationYear,
			&item.Description,
			&item.Authors,
			&item.Categories,
		); err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "could not read book details",
			})
			return
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "error reading book details",
		})
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *App) searchBooks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		http.Error(w, "Query parameter q is required", http.StatusBadRequest)
		return
	}

	rows, err := a.db.Query(r.Context(), `
		SELECT book_id, title, isbn, price::float8,
		       publication_year, description
		FROM books
		WHERE title ILIKE $1 OR isbn ILIKE $1
		ORDER BY book_id
	`, "%"+q+"%")
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	books := []Book{}
	for rows.Next() {
		var b Book
		if err := rows.Scan(
			&b.ID,
			&b.Title,
			&b.ISBN,
			&b.Price,
			&b.PublicationYear,
			&b.Description,
		); err != nil {
			http.Error(w, "Failed to read books", http.StatusInternalServerError)
			return
		}
		books = append(books, b)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to read books", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func (a *App) bookByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idText := strings.TrimPrefix(r.URL.Path, "/books/")
	id, err := strconv.Atoi(idText)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	var b Book
	err = a.db.QueryRow(r.Context(), `
		SELECT book_id, title, isbn, price::float8,
		       publication_year, description
		FROM books
		WHERE book_id = $1
	`, id).Scan(
		&b.ID,
		&b.Title,
		&b.ISBN,
		&b.Price,
		&b.PublicationYear,
		&b.Description,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}
