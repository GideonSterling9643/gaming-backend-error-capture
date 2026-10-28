package main

import (
	"encoding/json"
	"log"
	"net/http"

	"gaming-error-capture/internal"
)

type report struct {
	Event internal.LiveEvent `json:"event"`
	Asset internal.Asset     `json:"asset"`
}

func main() {
	client := internal.NewInfraiClient()
	blocked := map[string]bool{"texture": true, "audio": true}
	http.HandleFunc("/moderation/report", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}
		var input report
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid report", http.StatusBadRequest)
			return
		}
		item, queued := internal.QueueDecision(input.Event, input.Asset, blocked)
		if queued {
			capture := map[string]any{"title": "moderation queue decision", "message": item.Reason, "level": "info", "fingerprint": []string{"moderation", input.Asset.Kind}, "exception": map[string]any{"event": input.Event, "asset": input.Asset}, "context": map[string]any{"queue_item": item}}
			if err := client.Capture(capture); err != nil {
				log.Printf("capture rejected: %v", err)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"queued": queued, "item": item})
	})
	log.Println("game moderation service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
