
run-server:
	go run ./cmd/api-service/service.go -apikey localOnly

run-cli:
	go run ./cmd/terminal-client/main.go