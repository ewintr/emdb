
run-api:
	go run ./cmd/api-service/service.go -apikey localOnly

run-tui-local:
	EMDB_BASE_URL=http://localhost:8085/ EMDB_API_KEY=hoi go run ./cmd/terminal-client/main.go

run-tui:
	go run ./cmd/terminal-client/main.go

run-md-export:
	go run ./cmd/markdown-export/main.go


build-api:
	go build -o emdb-api ./cmd/api-service/service.go

deploy-api:
	ssh ewintr.nl /home/erik/bin/deploy-emdb-api.sh