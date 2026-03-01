.PHONY: help build up down logs clean test

help:
	@echo "Available commands:"
	@echo "  make build      - Build Docker images"
	@echo "  make up         - Start all services"
	@echo "  make down       - Stop all services"
	@echo "  make logs       - View application logs"
	@echo "  make db-logs    - View database logs"
	@echo "  make ps         - Show running containers"
	@echo "  make clean      - Remove containers and volumes"
	@echo "  make restart    - Restart all services"
	@echo "  make test       - Run tests"

build:
	@echo "Building Docker images..."
	docker compose build

up:
	@echo "Starting services..."
	docker compose up --build -d
	@echo "Services started! API is available at http://localhost:8080"
	@echo "Health check: http://localhost:8080/health"

down:
	@echo "Stopping services..."
	docker compose down

logs:
	@echo "Application logs:"
	docker logs -f practice4-api

db-logs:
	@echo "Database logs:"
	docker logs -f practice4-db

ps:
	@echo "Running containers:"
	docker ps -a

clean:
	@echo "Removing containers and volumes..."
	docker compose down -v
	docker system prune -f

restart:
	@echo "Restarting services..."
	docker compose restart

test:
	@echo "Testing API endpoints..."
	@echo "1. Health check:"
	curl -s http://localhost:8080/health | jq .
	@echo "\n2. Get all users:"
	curl -s http://localhost:8080/users | jq .
	@echo "\nTests completed!"