package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("=== Starting Minimal Tencent Bot ===")
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Minimal Tencent Bot Running")
	})
	
	fmt.Println("Server starting on port 8080...")
	fmt.Println("Access at: http://localhost:8080")
	
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}