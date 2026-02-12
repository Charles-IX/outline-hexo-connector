package test

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func PrintWebhook(w http.ResponseWriter, r *http.Request) {
	log.Println("Received webhook request:")
	fmt.Printf("Method: %s\n", r.Method)

	for k, v := range r.Header {
		fmt.Printf("Header: %s = %v\n", k, v)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading webhook - %v", err)
		http.Error(w, "Error reading webhook", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	fmt.Printf("Payload: %s\n", string(body))

	// 必须返回 200 OK，否则 Outline 会认为推送失败并尝试重试
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Acknowledged"))
}
