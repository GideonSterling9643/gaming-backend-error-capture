package internal

import (
	"encoding/json"
	"testing"
)

func TestDocumentedReportFields(t *testing.T) {
	var input struct {
		Event LiveEvent `json:"event"`
		Asset Asset     `json:"asset"`
	}
	data := []byte(`{"event":{"id":"evt-7","asset_id":"asset-7","player_id":"p-4"},"asset":{"id":"asset-7","owner_id":"p-4","kind":"texture"}}`)
	if err := json.Unmarshal(data, &input); err != nil {
		t.Fatal(err)
	}
	if input.Event.AssetID != "asset-7" || input.Event.PlayerID != "p-4" || input.Asset.OwnerID != "p-4" {
		t.Fatalf("documented fields not decoded: event=%+v asset=%+v", input.Event, input.Asset)
	}
	if _, queued := QueueDecision(input.Event, input.Asset, map[string]bool{"texture": true}); !queued {
		t.Fatal("documented blocked texture must be queued")
	}
}

func TestQueueDecision(t *testing.T) {
	tests := []struct {
		name    string
		event   LiveEvent
		asset   Asset
		blocked map[string]bool
		want    bool
		reason  string
	}{
		{"blocked player asset", LiveEvent{ID: "evt-1", AssetID: "asset-1"}, Asset{ID: "asset-1", Kind: "texture"}, map[string]bool{"texture": true}, true, "asset kind texture requires review"},
		{"allowed asset", LiveEvent{ID: "evt-2", AssetID: "asset-2"}, Asset{ID: "asset-2", Kind: "map"}, map[string]bool{"texture": true}, false, ""},
		{"different asset", LiveEvent{ID: "evt-3", AssetID: "asset-x"}, Asset{ID: "asset-3", Kind: "texture"}, map[string]bool{"texture": true}, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, queued := QueueDecision(tt.event, tt.asset, tt.blocked)
			if queued != tt.want || item.Reason != tt.reason {
				t.Fatalf("got (%v, %q), want (%v, %q)", queued, item.Reason, tt.want, tt.reason)
			}
		})
	}
}
