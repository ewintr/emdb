
run-api:
	go run ./cmd/api-service/service.go -apikey localOnly

run-tui:
	go run ./cmd/terminal-client/main.go

build-api:
	go build -o emdb-api ./cmd/api-service/service.go

deploy-api:
	ssh ewintr.nl deploy/emdb-api.sh