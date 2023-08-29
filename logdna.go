package logdna

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
)

var ingestionKey = os.Getenv("INGESTION_KEY")

func init() {
	if ingestionKey == "" {
		panic("empty INGESTION_KEY")
	}

	functions.CloudEvent("LogDNAUpload", logDNAUpload)
}

// MessagePublishedData contains the full Pub/Sub message
// See the documentation for more details:
// https://cloud.google.com/eventarc/docs/cloudevents#pubsub
type MessagePublishedData struct {
	Message PubSubMessage
}

// PubSubMessage is the payload of a Pub/Sub event.
// See the documentation for more details:
// https://cloud.google.com/pubsub/docs/reference/rest/v1/PubsubMessage
type PubSubMessage struct {
	Data        []byte    `json:"data"`
	PublishTime time.Time `json:"publishTime"`
}

func logDNAUpload(ctx context.Context, e event.Event) error {
	var msg MessagePublishedData
	if err := e.DataAs(&msg); err != nil {
		return fmt.Errorf("event.DataAs: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(msg.Message.Data, &parsed); err != nil {
		return err
	}

	labels := getLabels(parsed)

	// https://docs.mezmo.com/log-analysis-api/ref#ingest
	values := url.Values{
		"hostname": []string{labels["project_id"]},
		"now":      []string{fmt.Sprintf("%d", time.Now().UnixMicro())},
	}
	url := "https://logs.logdna.com/logs/ingest?" + values.Encode()

	line, err := json.Marshal(parsed["jsonPayload"])
	if err != nil {
		return err
	}

	timestamp := msg.Message.PublishTime
	if rawTimestamp, ok := parsed["timestamp"].(string); ok {
		parsed, err := time.Parse(time.RFC3339Nano, rawTimestamp)
		if err == nil {
			timestamp = parsed
		}
	}

	app := labels["service_name"]
	if job, ok := labels["job_name"]; ok {
		app = job
	}

	var meta map[string]string
	if rev, ok := labels["revision_name"]; ok {
		meta = map[string]string{
			"revision": rev,
		}
	}

	body := map[string]any{
		"lines": []any{
			map[string]any{
				"timestamp": fmt.Sprintf("%d", timestamp.UnixMilli()),
				"app":       app,
				"line":      string(line),
				"meta":      meta,
			},
		},
	}
	marshaled, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, url, bytes.NewReader(marshaled),
	)
	if err != nil {
		return err
	}

	// https://docs.mezmo.com/log-analysis-api/ref#authentication
	req.Header.Set("apikey", ingestionKey)

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		bod, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bad status from LogDNA: %s: %s", res.Status, bod)
	}

	return nil
}

func getLabels(data map[string]any) map[string]string {
	resource := data["resource"].(map[string]any)

	labels := resource["labels"].(map[string]any)

	out := make(map[string]string)
	for key, value := range labels {
		out[key] = value.(string)
	}
	return out
}
