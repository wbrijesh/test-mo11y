package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/marcboeker/go-duckdb"
)

func main() {
	dbPath := flag.String("db", "../mo11y/mo11y.duckdb", "Path to mo11y database")
	flag.Parse()

	db, err := sql.Open("duckdb", *dbPath+"?access_mode=read_only")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Query spans
	fmt.Println("=== SPANS ===")
	rows, err := db.Query(`
		SELECT trace_id, span_id, name, duration_ns, resource_attrs 
		FROM spans 
		ORDER BY start_time
	`)
	if err != nil {
		log.Fatal(err)
	}

	spanCount := 0
	for rows.Next() {
		var traceID, spanID, name string
		var durationNs int64
		var resourceAttrs duckdb.Map
		if err := rows.Scan(&traceID, &spanID, &name, &durationNs, &resourceAttrs); err != nil {
			log.Fatal(err)
		}
		spanCount++
		fmt.Printf("  [%d] %s/%s: %s (%.2fms)\n", 
			spanCount, traceID[:8], spanID[:8], name, float64(durationNs)/1e6)
		if svc, ok := resourceAttrs["service.name"]; ok {
			fmt.Printf("      service.name=%v\n", svc)
		}
	}
	rows.Close()
	fmt.Printf("Total: %d spans\n\n", spanCount)

	// Query events
	fmt.Println("=== SPAN EVENTS ===")
	rows, err = db.Query(`
		SELECT trace_id, span_id, event_name 
		FROM span_events 
		ORDER BY event_time
	`)
	if err != nil {
		log.Fatal(err)
	}

	eventCount := 0
	for rows.Next() {
		var traceID, spanID, eventName string
		if err := rows.Scan(&traceID, &spanID, &eventName); err != nil {
			log.Fatal(err)
		}
		eventCount++
		fmt.Printf("  [%d] %s/%s: %s\n", eventCount, traceID[:8], spanID[:8], eventName)
	}
	rows.Close()
	fmt.Printf("Total: %d events\n\n", eventCount)

	// Query links
	fmt.Println("=== SPAN LINKS ===")
	rows, err = db.Query(`
		SELECT trace_id, span_id, linked_trace_id, linked_span_id 
		FROM span_links
	`)
	if err != nil {
		log.Fatal(err)
	}

	linkCount := 0
	for rows.Next() {
		var traceID, spanID, linkedTraceID, linkedSpanID string
		if err := rows.Scan(&traceID, &spanID, &linkedTraceID, &linkedSpanID); err != nil {
			log.Fatal(err)
		}
		linkCount++
		fmt.Printf("  [%d] %s/%s -> %s/%s\n", 
			linkCount, traceID[:8], spanID[:8], linkedTraceID[:8], linkedSpanID[:8])
	}
	rows.Close()
	fmt.Printf("Total: %d links\n", linkCount)

	// Query logs
	fmt.Println("\n=== LOGS ===")
	rows, err = db.Query(`
		SELECT timestamp, severity_text, body, resource_attrs
		FROM logs 
		ORDER BY timestamp
	`)
	if err != nil {
		log.Fatal(err)
	}

	logCount := 0
	for rows.Next() {
		var timestamp, severityText, body string
		var resourceAttrs duckdb.Map
		if err := rows.Scan(&timestamp, &severityText, &body, &resourceAttrs); err != nil {
			log.Fatal(err)
		}
		logCount++
		fmt.Printf("  [%d] %s %s: %s\n", logCount, timestamp[:19], severityText, body)
	}
	rows.Close()
	fmt.Printf("Total: %d logs\n", logCount)

	// Query metrics
	fmt.Println("\n=== METRICS ===")
	rows, err = db.Query(`
		SELECT name, type, value, histogram_json, resource_attrs
		FROM metrics 
		ORDER BY timestamp
	`)
	if err != nil {
		log.Fatal(err)
	}

	metricCount := 0
	typeNames := map[int8]string{1: "Gauge", 2: "Sum", 3: "Histogram"}
	for rows.Next() {
		var name string
		var metricType int8
		var value float64
		var histogramJSON sql.NullString
		var resourceAttrs duckdb.Map
		if err := rows.Scan(&name, &metricType, &value, &histogramJSON, &resourceAttrs); err != nil {
			log.Fatal(err)
		}
		metricCount++
		typeName := typeNames[metricType]
		if metricType == 3 {
			fmt.Printf("  [%d] %s (%s): %s\n", metricCount, name, typeName, histogramJSON.String[:50]+"...")
		} else {
			fmt.Printf("  [%d] %s (%s): %.2f\n", metricCount, name, typeName, value)
		}
	}
	rows.Close()
	fmt.Printf("Total: %d metrics\n", metricCount)

	if spanCount == 0 && eventCount == 0 && logCount == 0 && metricCount == 0 {
		fmt.Fprintf(os.Stderr, "\nNo data found. Run 'make send-traces' first.\n")
		os.Exit(1)
	}
}
