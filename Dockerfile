# Stage 1: Build the frontend using Bun
FROM oven/bun:1-alpine as frontend-builder
WORKDIR /app
COPY frontend/package.json frontend/bun.lock ./
RUN bun install --frozen-lockfile
COPY frontend/ ./
RUN bun run build

# Stage 2: Build the Go backend
FROM golang:1.26.3-alpine as backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Copy the built frontend into the Go build context
# (Required if your Go app uses go:embed to serve the frontend)
COPY --from=frontend-builder /app/dist ./frontend/dist

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o graph-auth-server .

# Stage 3: Final runtime image
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

# Copy the compiled binary from the backend builder
COPY --from=backend-builder /app/graph-auth-server .

# Copy the frontend dist directly to the filesystem just in case
# your Go app serves static files directly from disk instead of embedding
COPY --from=frontend-builder /app/dist ./frontend/dist

# Dokku automatically sets and routes the PORT environment variable.
ENV PORT=8080
EXPOSE 8080

CMD ["./graph-auth-server"]
