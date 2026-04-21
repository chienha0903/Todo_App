## Chạy từ thư mục root của monorepo: make -C services/todos <target>
## Hoặc cd services/todos && make <target>
ROOT = ../..

.PHONY: run build wire

run:
	go run $(ROOT)/services/todos/cmd/main.go

build:
	go build -o $(ROOT)/bin/todos $(ROOT)/services/todos/cmd/main.go

wire:
	wire gen $(ROOT)/services/todos/internal/di/
