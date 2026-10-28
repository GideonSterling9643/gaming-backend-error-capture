package internal

import "fmt"

type Asset struct {
	ID      string
	OwnerID string `json:"owner_id"`
	Kind    string
}
type LiveEvent struct {
	ID       string
	AssetID  string `json:"asset_id"`
	PlayerID string `json:"player_id"`
}
type ModerationItem struct{ EventID, Reason string }

func QueueDecision(event LiveEvent, asset Asset, blockedKinds map[string]bool) (ModerationItem, bool) {
	if event.AssetID != asset.ID {
		return ModerationItem{}, false
	}
	if blockedKinds[asset.Kind] {
		return ModerationItem{EventID: event.ID, Reason: fmt.Sprintf("asset kind %s requires review", asset.Kind)}, true
	}
	return ModerationItem{}, false
}
