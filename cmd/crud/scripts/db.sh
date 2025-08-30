#!/bin/bash

# Database management script for CRUD demo

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Database connection details
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="blogdb"
DB_USER="bloguser"
DB_PASSWORD="blogpass"
MIGRATION_PATH="db/migrations"
CONTAINER_NAME="crud-postgres"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker first."
        exit 1
    fi
}

# Function to check if migrate is installed
check_migrate() {
    if ! command -v migrate > /dev/null 2>&1; then
        print_error "golang-migrate is not installed. Please install it first:"
        echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
        exit 1
    fi
}

# Function to check if sqlc is installed
check_sqlc() {
    if ! command -v sqlc > /dev/null 2>&1; then
        print_error "sqlc is not installed. Please install it first:"
        echo "  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"
        exit 1
    fi
}

# Start the database
start_db() {
    print_status "Starting PostgreSQL database..."
    check_docker
    
    # Check if container already exists
    if docker ps -a --format "table {{.Names}}" | grep -q "$CONTAINER_NAME"; then
        if docker ps --format "table {{.Names}}" | grep -q "$CONTAINER_NAME"; then
            print_status "Database is already running"
            return 0
        else
            print_status "Starting existing container..."
            docker start "$CONTAINER_NAME"
        fi
    else
        print_status "Creating new database container..."
        docker run -d \
            --name "$CONTAINER_NAME" \
            -e POSTGRES_DB="$DB_NAME" \
            -e POSTGRES_USER="$DB_USER" \
            -e POSTGRES_PASSWORD="$DB_PASSWORD" \
            -p "$DB_PORT:$DB_PORT" \
            postgres:15-alpine
    fi
    
    print_status "Waiting for database to be ready..."
    sleep 3
    
    # Wait for database to be ready
    for i in {1..30}; do
        if docker exec "$CONTAINER_NAME" pg_isready -U "$DB_USER" -d "$DB_NAME" > /dev/null 2>&1; then
            print_status "Database is ready!"
            return 0
        fi
        sleep 1
    done
    
    print_error "Database failed to start properly"
    exit 1
}

# Stop the database
stop_db() {
    print_status "Stopping PostgreSQL database..."
    docker stop "$CONTAINER_NAME" 2>/dev/null || true
    print_status "Database stopped"
}

# Reset the database (stop, remove volumes, start)
reset_db() {
    print_warning "This will delete all data in the database!"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_status "Resetting database..."
        docker stop "$CONTAINER_NAME" 2>/dev/null || true
        docker rm "$CONTAINER_NAME" 2>/dev/null || true
        start_db
        print_status "Database reset complete"
    else
        print_status "Reset cancelled"
    fi
}

# Run migrations
run_migrations() {
    print_status "Running database migrations..."
    check_migrate
    
    DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"
    
    migrate -path $MIGRATION_PATH -database "$DATABASE_URL" up
    
    print_status "Migrations completed"
}

# Generate sqlc code
generate_code() {
    print_status "Generating sqlc code..."
    check_sqlc
    
    sqlc generate
    
    print_status "Code generation completed"
}

# Show database status
status() {
    print_status "Database status:"
    if docker ps --format "table {{.Names}}" | grep -q "$CONTAINER_NAME"; then
        print_status "✓ PostgreSQL is running"
    else
        print_status "✗ PostgreSQL is not running"
    fi
    
    print_status "Connection details:"
    echo "  Host: $DB_HOST"
    echo "  Port: $DB_PORT"
    echo "  Database: $DB_NAME"
    echo "  User: $DB_USER"
    echo "  Password: $DB_PASSWORD"
}

# Show logs
logs() {
    docker logs "$CONTAINER_NAME"
}

# Main script logic
case "${1:-}" in
    "start")
        start_db
        ;;
    "stop")
        stop_db
        ;;
    "reset")
        reset_db
        ;;
    "migrate")
        run_migrations
        ;;
    "generate")
        generate_code
        ;;
    "setup")
        start_db
        run_migrations
        generate_code
        print_status "Setup complete! You can now run the application."
        ;;
    "status")
        status
        ;;
    "logs")
        logs
        ;;
    *)
        echo "Usage: $0 {start|stop|reset|migrate|generate|setup|status|logs}"
        echo ""
        echo "Commands:"
        echo "  start     - Start the PostgreSQL database"
        echo "  stop      - Stop the PostgreSQL database"
        echo "  reset     - Reset the database (delete all data)"
        echo "  migrate   - Run database migrations"
        echo "  generate  - Generate sqlc code"
        echo "  setup     - Complete setup (start db, run migrations, generate code)"
        echo "  status    - Show database status"
        echo "  logs      - Show database logs"
        exit 1
        ;;
esac
