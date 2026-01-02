# Build Frontend
FROM node:24.12.0-alpine AS frontend-builder
WORKDIR /app/frontend

# Configure npm for better network resilience
RUN npm config set fetch-retries 5 && \
    npm config set fetch-retry-mintimeout 20000 && \
    npm config set fetch-retry-maxtimeout 120000 && \
    npm config set fetch-timeout 300000

COPY frontend/package*.json ./
# Use npm ci (faster and more reliable with package-lock.json)
RUN npm ci --prefer-offline --no-audit

COPY frontend/ ./
# Set VITE_API_URL to use relative path that works with BASE_PATH
ENV VITE_API_URL=/darts/api
RUN npm run build

# Build Backend
FROM golang:1.25.5-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
# CGO_ENABLED=1 is needed for go-sqlite3, requiring gcc
RUN apk add --no-cache gcc musl-dev
RUN CGO_ENABLED=1 GOOS=linux go build -o darts-server cmd/server/main.go

# Final Stage
FROM alpine:latest
WORKDIR /app
# Install sqlite libs if needed (often static, but good to have)
RUN apk add --no-cache sqlite-libs

COPY --from=backend-builder /app/darts-server .
COPY --from=frontend-builder /app/frontend/dist ./dist

# Create /data directory with proper permissions for SQLite
# SQLite needs write access to the directory for lock files and journal files
RUN mkdir -p /data && chmod 777 /data

# Environment variables will be set in K8s
ENV PORT=8080
ENV DB_PATH=/data/darts.db
ENV BASE_PATH=/darts

EXPOSE 8080
CMD ["./darts-server"]
