# --- Build stage ---
FROM golang:alpine AS builder

WORKDIR /app

# Copy go.mod/go.sum từ root (single module)
COPY go.mod go.sum ./
RUN go mod download

# Copy toàn bộ source
COPY . .

# Build binary
RUN go build -o /bin/todo-bff ./services/todo-bff/cmd/main.go

# --- Runtime stage ---
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /bin/todo-bff /bin/todo-bff

EXPOSE 8080

ENTRYPOINT ["/bin/todo-bff"]