# Library Management REST API

A REST API for library management built with Go and PostgreSQL.

## Features
- List and create books
- Get a book by ID
- Search books by title or ISBN
- Manage authors and categories
- Associate books with authors and categories
- List library branches
- Check API and database health

## Requirements
- Go 1.21 or later
- PostgreSQL

## Configuration
Set DB_PASSWORD in your local environment. Never commit passwords to GitHub.

## Run
```bash
go mod download
go run .
```

## API Endpoints
- GET /health
- GET /books
- POST /books
- GET /books/details
- GET /books/{id}
- GET /books/search?q=Go
- GET /authors
- POST /authors
- GET /categories
- POST /categories
- GET /branches
- POST /books/authors
- POST /books/categories

## Database
Create the PostgreSQL database, user, and tables before running the application.
- GET /books
- POST /books
- GET /books/details
- GET /books/{id}
- GET /books/search?q=Go
- GET /authors
- POST /authors
- GET /categories
- POST /categories
- GET /branches
- POST /books/authors
- POST /books/categories

## Database Setup
Create the PostgreSQL database, user, and tables before running the application.
Configure database credentials locally using environment variables.
Never commit passwords or other secrets to GitHub.
