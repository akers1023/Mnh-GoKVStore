package main

import (
  "fmt"
  "log"
  "net/http"
  "os"
)

func handler(w http.ResponseWriter, r *http.Request) {
  port := os.Getenv("PORT")
  if port == "" {
    port = "8081"
  }
  fmt.Fprintf(w, "Hello from Backend Server on port %s", port)
}

func main() {
  port := os.Getenv("PORT")
  if port == "" {
    port = "8081"
  }
  
  fmt.Printf("Starting backend server on port %s\n", port)
  http.HandleFunc("/", handler)
  log.Fatal(http.ListenAndServe(":"+port, nil))
}
