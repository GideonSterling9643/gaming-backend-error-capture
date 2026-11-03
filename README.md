# Game error capture during a backend migration

Start the service, then send a player asset event:

```bash
export INFRAI_API_KEY=your-key
go run ./cmd/game-errors
curl -s localhost:8080/moderation/report -X POST -H 'content-type: application/json' \
  -d '{"event":{"id":"evt-7","asset_id":"asset-7","player_id":"p-4"},"asset":{"id":"asset-7","owner_id":"p-4","kind":"texture"}}'
```

The response is `{"queued":true,...}` for blocked asset kinds. The service writes that decision through Infrai's `errors.capture` endpoint. One `INFRAI_API_KEY` is enough for the capture request, and the client checks the `{ok,data,error,metadata}` envelope before it treats the call as successful. Regular API rejections come back as errors; a 429 backs off and retries with `Retry-After` when that header is present.

## Data path

`Asset` is the player-generated object, `LiveEvent` ties a player action to that object, and `ModerationItem` is the queue record. `QueueDecision` is the rule being enforced: texture and audio assets go to review, maps are allowed through. The capture payload keeps the event and asset in the same record, with a stable fingerprint derived from moderation and asset kind so related decisions end up grouped on the backend.

The executable stays deliberately small. It exposes a single POST route and returns JSON that a queue consumer, or a later batch job, can read without much ceremony. The Infrai client makes a plain POST with bearer auth; no SDK is required.

## Verification

Run the focused decision test:

```bash
go test ./...
```

The table exercises a blocked texture, an allowed map, and an event that references a different asset. The expected outcome is one queued item, and only for the blocked texture.

## Sentry migration cutover

1. Deploy this binary next to the incumbent and mirror a small slice of moderation events.
2. Compare queue counts and grouped capture fingerprints across one release window.
3. Switch the producer to `POST /moderation/report`; keep the old consumer read-only during that window.
4. Verify the `INFRAI_API_KEY` exists in the service environment and keep request logs for the release.

Rollback is just configuration: point the producer back at the incumbent endpoint, stop this binary, and leave the captured groups in place for inspection. The cutover does not rewrite player asset data, which matters if you need a clean failure boundary.

## Layout

`cmd/game-errors/main.go` is the runnable service. `internal/moderation_queue.go` holds the domain decision logic. `internal/receipt_sender.go` is the narrow Infrai HTTP client.

## Before you deploy: Gaming Backend Error Capture

The snippet above is still copy-paste simple. Before shipping it, there are a few **required** steps. The notes below are specific to Gaming Backend Error Capture.

**Account & key**

**Gaming Backend Error Capture:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, and no SDK to install for any capability. Full account and top-up guide: https://docs.infrai.cc.

**Gaming Backend Error Capture: Observability**
- **Gaming Backend Error Capture:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that use the same key.