.PHONY: help send-traces query-db

help:
	@echo "test-mo11y - Testing tools for mo11y"
	@echo ""
	@echo "Usage:"
	@echo "  make send-traces  Send test traces to mo11y (must be running)"
	@echo "  make query-db     Query mo11y database"

send-traces:
	@echo "Sending test traces to mo11y at localhost:4318..."
	@go run ./cmd/send-traces

query-db:
	@go run ./cmd/query-db -db ../mo11y/mo11y.duckdb
