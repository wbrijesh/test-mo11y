.PHONY: help send-traces send-logs send-metrics query-db

help:
	@echo "test-mo11y - Testing tools for mo11y"
	@echo ""
	@echo "Usage:"
	@echo "  make send-traces   Send test traces to mo11y"
	@echo "  make send-logs     Send test logs to mo11y"
	@echo "  make send-metrics  Send test metrics to mo11y"
	@echo "  make query-db      Query mo11y database"

send-traces:
	@go run ./cmd/send-traces

send-logs:
	@go run ./cmd/send-logs

send-metrics:
	@go run ./cmd/send-metrics

query-db:
	@go run ./cmd/query-db -db ../mo11y/mo11y.duckdb
