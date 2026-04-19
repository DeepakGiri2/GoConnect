package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("DEV_SERVER_PORT")
	if port == "" {
		port = "3000"
	}

	fs := http.FileServer(http.Dir("./temp"))
	http.Handle("/", corsMiddleware(fs))

	fmt.Printf("\n")
	fmt.Printf("╔═══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                                                               ║\n")
	fmt.Printf("║   ⚠️  DEVELOPMENT TEST SERVER - DO NOT USE IN PRODUCTION ⚠️     ║\n")
	fmt.Printf("║                                                               ║\n")
	fmt.Printf("╚═══════════════════════════════════════════════════════════════╝\n")
	fmt.Printf("\n")
	fmt.Printf("🚀 Serving static files from ./temp directory\n")
	fmt.Printf("🌐 Server running at: http://localhost:%s\n", port)
	fmt.Printf("📄 Test page: http://localhost:%s/index.html\n", port)
	fmt.Printf("\n")
	fmt.Printf("Press Ctrl+C to stop\n")
	fmt.Printf("\n")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
