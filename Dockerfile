FROM node:24-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.25-bookworm AS backend
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend /app/backend/internal/frontend/dist ./internal/frontend/dist
RUN CGO_ENABLED=1 go build -trimpath -o /app/server ./cmd/server

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/* && useradd --uid 10001 --create-home app
WORKDIR /app
COPY --from=backend /app/server /app/server
RUN mkdir -p /app/data && chown -R app:app /app
USER app
ENV DATA_DIR=/app/data GIN_MODE=release PORT=8080
EXPOSE 8080
VOLUME ["/app/data"]
CMD ["/app/server"]
