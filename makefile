PHONY fmt vet run

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

run: vet
	go run ./long-polling/longpolling_local/longpolling_server.go