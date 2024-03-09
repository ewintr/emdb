.PHONY: tui, md-exprt, worker

# Define source and destination directories
MD_SRC_DIR := public
MD_DST_DIR := ../ewintr.nl/content/movies

tui:
	go run ./terminal-client/main.go

md-export:
	go run ./markdown-export/main.go

worker:
	go run ./worker-client/main.go

