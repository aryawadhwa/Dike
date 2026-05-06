package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aryawadhwa/dike/pkg/audit"
)

// StartServer starts a minimal HTTP server for the Web Audit UI.
func StartServer(port int) error {
	http.HandleFunc("/api/audit", func(w http.ResponseWriter, r *http.Request) {
		logs, err := audit.GetAuditLogs()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(logs)
	})

	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		stats, err := audit.GetStats()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	// Serve static files from the frontend directory
	fs := http.FileServer(http.Dir("../frontend"))
	http.Handle("/", fs)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("🌐 Web Audit UI started at http://localhost%s\n", addr)
	return http.ListenAndServe(addr, nil)
}
