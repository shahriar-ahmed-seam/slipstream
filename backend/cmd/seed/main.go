// Command seed generates synthetic event traffic against a running metrics
// service. It is a developer convenience for demoing the dashboard without
// wiring up a real producer. Run the backend first, then `make seed`.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type event struct {
	ID        string            `json:"id"`
	Source    string            `json:"source"`
	Metric    string            `json:"metric"`
	Value     float64           `json:"value"`
	Timestamp time.Time         `json:"timestamp"`
	IsError   bool              `json:"isError"`
	Tags      map[string]string `json:"tags,omitempty"`
}

type batch struct {
	Events []event `json:"events"`
}

func main() {
	target := flag.String("target", "http://localhost:8080", "metrics service base URL")
	rate := flag.Int("rate", 200, "approximate events per second")
	errPct := flag.Float64("errors", 0.07, "fraction of events flagged as errors (0..1)")
	flag.Parse()

	if *rate < 1 {
		*rate = 1
	}
	url := *target + "/api/events"

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	client := &http.Client{Timeout: 5 * time.Second}
	tick := 100 * time.Millisecond
	perTick := int(float64(*rate) * tick.Seconds())
	if perTick < 1 {
		perTick = 1
	}

	routes := []string{"/api/login", "/api/orders", "/api/search", "/api/checkout", "/api/profile"}
	metrics := []string{"http.request.duration_ms", "db.query.duration_ms", "cache.lookup.duration_ms"}

	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	var sent uint64
	fmt.Fprintf(os.Stdout, "seeding %s at ~%d ev/s (errors %.0f%%); ctrl-c to stop\n", url, *rate, *errPct*100)

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintf(os.Stdout, "\nstopped after %d events\n", sent)
			return
		case <-ticker.C:
			evs := make([]event, 0, perTick)
			now := time.Now().UTC()
			for i := 0; i < perTick; i++ {
				metric := metrics[rand.Intn(len(metrics))]
				route := routes[rand.Intn(len(routes))]
				// Log-normalish latency so the histogram and p95 look realistic.
				base := 20 + 60*math.Abs(rand.NormFloat64())
				isErr := rand.Float64() < *errPct
				status := "200"
				if isErr {
					base += 200 + 300*rand.Float64()
					status = "500"
				}
				evs = append(evs, event{
					ID:        fmt.Sprintf("evt-%d", time.Now().UnixNano()),
					Source:    "load-generator",
					Metric:    metric,
					Value:     base,
					Timestamp: now,
					IsError:   isErr,
					Tags:      map[string]string{"route": route, "status": status},
				})
			}
			if err := post(ctx, client, url, evs); err != nil {
				fmt.Fprintf(os.Stderr, "post failed: %v\n", err)
				continue
			}
			sent += uint64(len(evs))
		}
	}
}

func post(ctx context.Context, client *http.Client, url string, evs []event) error {
	payload, err := json.Marshal(batch{Events: evs})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}
