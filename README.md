# Library Management REST API

A REST API for library management built with Go and PostgreSQL.

This project provides endpoints to manage books, authors, categories, and library branches.

## Features

- Check API and database health
- List all books
- Create new books
- Retrieve a book by ID
- Search books by title or ISBN
- List and create authors
- List and create categories
- Associate books with authors
- Associate books with categories
- List library branches
- Return detailed book information, including authors and categories

## Technologies

- Go 1.21+
- PostgreSQL
- pgx PostgreSQL driver

## Requirements

Before running the application, make sure you have installed:

- Go 1.21 or later
- PostgreSQL 14 or later
- Git

## Configuration

The application uses environment variables for database configuration.

Set your database password in the terminal before starting the application:

```bash
export DB_PASSWORD="your_database_password"
```

Replace `your_database_password` with your own PostgreSQL password.

Never commit real passwords, tokens, or other secrets to GitHub.

The application uses these default settings unless configured otherwise:

| Variable | Default |
|---|---|
| `APP_PORT` | `8080` |
| `DB_HOST` | `localhost` |
| `DB_PORT` | `5432` |
| `DB_NAME` | `library_db` |
| `DB_USER` | `library_app` |
| `DB_PASSWORD` | Must be configured locally |

## Database Setup

Before starting the API, create the PostgreSQL database, configure the database user, and create the tables required by the application.

Make sure the database credentials match your local environment configuration.

The application requires a working PostgreSQL connection.

## Installation and Running

Clone the repository:

```bash
git clone https://github.com/HaniyehNazari/library-api.git
cd library-api
```

Download the Go dependencies:

```bash
go mod download
```

Configure your database password:

```bash
export DB_PASSWORD="your_database_password"
```

Start the API:

```bash
go run .
```

By default, the API runs at:

```text
http://localhost:8080
```

Keep the terminal running while using the API.

## API Endpoints

### Health

| Method | Endpoint | Description |
|---|---|---|
| GET | `/health` | Check API and database health |

### Books

| Method | Endpoint | Description |
|---|---|---|
| GET | `/books` | List books |
| POST | `/books` | Create a book |
| GET | `/books/details` | Get books with author and category details |
| GET | `/books/{id}` | Get a book by ID |
| GET | `/books/search?q=Go` | Search books by title or ISBN |

### Authors

| Method | Endpoint | Description |
|---|---|---|
| GET | `/authors` | List authors |
| POST | `/authors` | Create an author |

### Categories

| Method | Endpoint | Description |
|---|---|---|
| GET | `/categories` | List categories |
| POST | `/categories` | Create a category |

### Library Branches

| Method | Endpoint | Description |
|---|---|---|
| GET | `/branches` | List library branches |

### Book Relationships

| Method | Endpoint | Description |
|---|---|---|
| POST | `/books/authors` | Associate a book with an author |
| POST | `/books/categories` | Associate a book with a category |

## Testing the API

Check the health endpoint:

```bash
curl http://localhost:8080/health
```

List books:

```bash
curl http://localhost:8080/books
```

Get a book by ID:

```bash
curl
