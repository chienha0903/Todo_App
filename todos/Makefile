APP_NAME=todo-app 
 
.PHONY: run tidy fmt vet build

run:
	go run cmd/main.go

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

build:
	go build -o $(APP_NAME) cmd/main.go
