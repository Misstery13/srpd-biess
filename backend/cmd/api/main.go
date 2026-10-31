package main

import (
	"log"
	"net/http"
	"os"
)

// Punto de entrada. El cableado de dependencias (repository -> service -> handler)
// se hace aquí cuando existan las capas en internal/.
func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Printf("API escuchando en :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
