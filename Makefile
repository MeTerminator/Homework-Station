.PHONY: install install-dev dev-backend dev-web build test format format-check

install:
	cd backend && go mod download
	cd web && npm ci

install-dev: install
	python3 -m venv .venv
	.venv/bin/python -m pip install -r requirements-dev.txt

dev-backend:
	cd backend && go run ./cmd/server

dev-web:
	cd web && npm run dev

build:
	cd web && npm run build
	cd backend && go build -o bin/homework-station ./cmd/server

test:
	cd backend && go test ./...
	python3 -m unittest discover -s tests -p 'test_*.py'
	cd web && npm test
	cd web && npm run build

format:
	python3 scripts/format.py

format-check:
	python3 scripts/format.py --check
