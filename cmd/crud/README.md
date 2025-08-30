# CRUD Demo Application

A simple blog CRUD API built with Go, PostgreSQL, and sqlc. This demonstrates a complete setup with database migrations, type-safe SQL queries, and a RESTful API.

## Features

- **RESTful API** with full CRUD operations for blog posts
- **PostgreSQL** database with Docker
- **Database migrations** using golang-migrate
- **Type-safe SQL** with sqlc code generation
- **Plain net/http** with Go 1.22+ pattern matching
- **Easy setup** with shell scripts

## Prerequisites

Before running this application, make sure you have the following installed:

- **Go 1.22+** (for pattern matching in ServeMux)
- **Docker**
- **golang-migrate** CLI tool
- **sqlc** CLI tool

### Installing Required Tools

```bash
# Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Install sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

## Quick Start

1. **Setup the database and run migrations:**

   ```bash
   cd cmd/crud
   ./scripts/db.sh setup
   ```

2. **Add PostgreSQL driver to go.mod:**

   ```bash
   go get github.com/lib/pq
   ```

3. **Run the application:**
   ```bash
   go run main.go
   ```

The API will be available at `http://localhost:8080`

## API Endpoints

### Health Check

- `GET /health` - Check if the service is running

### Posts

- `GET /posts` - List all posts (supports `limit` and `offset` query parameters)
- `POST /posts` - Create a new post
- `GET /posts/{id}` - Get a specific post by ID
- `PUT /posts/{id}` - Update a post by ID
- `DELETE /posts/{id}` - Delete a post by ID
- `GET /posts/author/{author}` - Get posts by author (supports `limit` and `offset` query parameters)

## Database Management

The `scripts/db.sh` script provides easy database management:

```bash
# Start the database
./scripts/db.sh start

# Stop the database
./scripts/db.sh stop

# Reset the database (delete all data)
./scripts/db.sh reset

# Run migrations
./scripts/db.sh migrate

# Generate sqlc code
./scripts/db.sh generate

# Complete setup (start db, run migrations, generate code)
./scripts/db.sh setup

# Check database status
./scripts/db.sh status

# View database logs
./scripts/db.sh logs
```

## Database Schema

The application uses a simple `posts` table:

```sql
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    author VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## Example API Usage

### Create a Post

```bash
curl -X POST http://localhost:8080/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My First Post",
    "content": "This is the content of my first blog post.",
    "author": "John Doe"
  }'
```

### Get All Posts

```bash
curl http://localhost:8080/posts
```

### Get Posts with Pagination

```bash
curl "http://localhost:8080/posts?limit=5&offset=0"
```

### Get a Specific Post

```bash
curl http://localhost:8080/posts/1
```

### Update a Post

```bash
curl -X PUT http://localhost:8080/posts/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Title",
    "content": "Updated content.",
    "author": "John Doe"
  }'
```

### Delete a Post

```bash
curl -X DELETE http://localhost:8080/posts/1
```

### Get Posts by Author

```bash
curl "http://localhost:8080/posts/author/John%20Doe?limit=10"
```

## Development

### Project Structure

```
cmd/crud/
├── main.go                 # Main application
├── sqlc.yaml              # sqlc configuration
├── Dockerfile.db          # Database container setup
├── scripts/
│   └── db.sh             # Database management script
├── db/
│   ├── migrations/       # Database migrations
│   ├── query/           # SQL queries for sqlc
│   └── generated/       # Generated Go code (after running sqlc)
└── README.md            # This file
```

### Adding New Migrations

1. Create new migration files:

   ```bash
   migrate create -ext sql -dir db/migrations -seq add_new_table
   ```

2. Edit the generated `.up.sql` and `.down.sql` files

3. Run the migration:
   ```bash
   ./scripts/db.sh migrate
   ```

### Adding New Queries

1. Add SQL queries to `db/query/` files
2. Generate Go code:
   ```bash
   ./scripts/db.sh generate
   ```

### Environment Variables

The application supports the following environment variables:

- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_NAME` - Database name (default: blogdb)
- `DB_USER` - Database user (default: bloguser)
- `DB_PASSWORD` - Database password (default: blogpass)
- `PORT` - Application port (default: 8080)

## Troubleshooting

### Database Connection Issues

- Make sure Docker is running
- Check if the database container is up: `./scripts/db.sh status`
- Verify the database is ready: `./scripts/db.sh logs`
- If issues persist, try: `./scripts/db.sh reset`

### Migration Issues

- Ensure golang-migrate is installed correctly
- Check migration files for syntax errors
- Use `./scripts/db.sh reset` to start fresh

### sqlc Issues

- Ensure sqlc is installed correctly
- Check `sqlc.yaml` configuration
- Verify SQL query syntax in `db/query/` files

## Testing

To test the API endpoints, you can use the curl examples above or any API testing tool like Postman or Insomnia.
