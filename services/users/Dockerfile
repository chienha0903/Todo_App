# --- Build stage ---
FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/users ./services/users/cmd/main.go

# --- Runtime stage ---
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/users /bin/users

COPY --from=builder /app/services/users/internal/infra/datastore/migrations \
  ./services/users/internal/infra/datastore/migrations

EXPOSE 50052

ENTRYPOINT ["/bin/users"]