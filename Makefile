
# Define source and destination directories
MD_SRC_DIR := public
MD_DST_DIR := ../ewintr.nl/content/movies

run-api:
	go run ./cmd/api-service/service.go -apikey localOnly

run-tui-local:
	EMDB_BASE_URL=http://localhost:8085/ EMDB_API_KEY=hoi go run ./cmd/terminal-client/main.go

run-tui:
	go run ./terminal-client/main.go

run-md-export:
	go run ./cmd/markdown-export/main.go
	for dir in $(MD_SRC_DIR)/*; do \
   		if [ -n "$$(ls -A $$dir)" ]; then \
   			cp -r $$dir/* $(MD_DST_DIR)/`basename $$dir`; \
   		fi \
   	done

run-worker:
	go run ./cmd/worker/main.go


build-api:
	go build -o emdb-api ./cmd/api-service/service.go

deploy-api:
	ssh ewintr.nl /home/erik/bin/deploy-emdb-api.sh