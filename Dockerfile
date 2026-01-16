# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./

# Install dependencies
RUN npm ci --ignore-scripts || npm install --ignore-scripts

# Copy frontend source
COPY tsconfig.json index.html ./
COPY src/ ./src/
COPY public/ ./public/

# Build frontend
RUN npm run build

# Stage 2: Build backend
FROM golang:1.22-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY tile-backend/go.mod tile-backend/go.sum ./

# Download dependencies
RUN go mod download

# Copy backend source
COPY tile-backend/ ./

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /server ./cmd/server/main.go

# Stage 3: Final image
FROM alpine:3.19

# Install nginx and supervisor
RUN apk add --no-cache nginx supervisor

# Create necessary directories
RUN mkdir -p /var/log/supervisor /run/nginx /var/www/html

# Copy frontend build
COPY --from=frontend-builder /app/dist /var/www/html

# Copy backend binary
COPY --from=backend-builder /server /usr/local/bin/server

# Copy nginx configuration
COPY docker/nginx.conf /etc/nginx/http.d/default.conf

# Copy supervisor configuration
COPY docker/supervisord.conf /etc/supervisord.conf

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost/api/v1/health || exit 1

# Start supervisor
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]
