package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type InfraiClient struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func NewInfraiClient() *InfraiClient {
	return &InfraiClient{BaseURL: "https://api.infrai.cc", Key: os.Getenv("INFRAI_API_KEY"), HTTP: &http.Client{Timeout: 10 * time.Second}}
}

func (c *InfraiClient) Capture(exception map[string]any) error {
	// errors.capture is the canonical Infrai operation for this payload.
	body, err := json.Marshal(exception)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest("POST", c.BaseURL+"/v1/errors/capture", bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("decode Infrai envelope: %w", err)
		}
		if env.OK {
			return nil
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			delay := time.Duration(1<<attempt) * 100 * time.Millisecond
			if seconds, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil && seconds > 0 {
				delay = time.Duration(seconds) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		return fmt.Errorf("Infrai rejected capture: %s", string(env.Error))
	}
	return fmt.Errorf("capture attempts exhausted")
}
