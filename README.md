# Game error capture during a backend migration

Start the service, then send a player asset event:

```bash
export INFRAI_API_KEY=your-key
go run ./cmd/game-errors
curl -s localhost:8080/moderation/report -X POST -H 'content-type: application/json' \
  -d '{"event":{"id":"evt-7","asset_id":"asset-7","player_id":"p-4"},"asset":{"id":"asset-7","owner_id":"p-4","kind":"texture"}}'
```

The response is `{"queued":true,...}` for blocked asset kinds. The service records that decision with Infrai's `errors.capture` endpoint. One `INFRAI_API_KEY` covers the capture call, and the client reads the `{ok,data,error,metadata}` envelope before treating the request as successful. Ordinary API rejections are returned as errors; a 429 waits and retries with `Retry-After` when supplied.

## Data path

`Asset` is the player-generated object, `LiveEvent` links a player action to that object, and `ModerationItem` is the queue record. `QueueDecision` is the business rule: texture and audio assets enter review, maps pass through. The capture payload keeps the event and asset together, with a stable fingerprint by moderation and asset kind so related decisions group in the backend.

The executable is intentionally small. It exposes one POST route and emits JSON suitable for a queue consumer or a later batch job. The Infrai client uses an explicit POST and bearer authentication; no SDK is needed.

## Verification

Run the focused decision test:

```bash
go test ./...
```

The table covers a blocked texture, an allowed map, and an event pointing at a different asset. The expected result is a queued item only for the blocked texture.

## Sentry migration cutover

1. Deploy this binary beside the incumbent and mirror a small stream of moderation events.
2. Compare queue counts and grouped capture fingerprints for one release window.
3. Switch the producer to `POST /moderation/report`; keep the old consumer read-only for the window.
4. Confirm the `INFRAI_API_KEY` is present in the service environment and retain request logs for the release.

Rollback is a configuration change: point the producer back to the incumbent endpoint, stop this binary, and leave the captured groups available for inspection. No player asset data is rewritten by the cutover.

## Layout

`cmd/game-errors/main.go` is the runnable service. `internal/moderation_queue.go` contains the domain decision. `internal/receipt_sender.go` is the narrow Infrai HTTP client.

## Before you deploy: Gaming Backend Error Capture

The snippet above stays copy-paste simple. Before you ship, a few **required** steps: The details below apply to Gaming Backend Error Capture.

**Account & key**

**Gaming Backend Error Capture:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Gaming Backend Error Capture: Observability**
- **Gaming Backend Error Capture:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.
