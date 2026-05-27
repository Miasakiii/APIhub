# Build stage for frontend
FROM node:22-alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Build stage for backend
FROM golang:1.26-alpine AS backend-builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o apihub ./cmd/apihub

# Production stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /app/apihub .

# Copy frontend dist
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Create data directory
RUN mkdir -p /app/data

ENV APIHUB_DATA_DIR=/app/data
ENV APIHUB_PORT=8080
ENV APIHUB_FRONTEND_DIST=/app/frontend/dist
ENV APIHUB_CORS_ORIGIN=http://localhost:8080

EXPOSE 8080

CMD ["./apihub"]
