package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"google.golang.org/protobuf/proto"
	pb "ac.local/ingest/pb"
)

// --------- модели для ClickHouse JSONEachRow ----------
type rawRow struct {
	BatchID      string `json:"batch_id"`
	SessionID    string `json:"session_id"`
	PlayerIDHash string `json:"player_id_hash"`
	ReceivedAt   string `json:"received_at"`
	PayloadBytes string `json:"payload_bytes"` // base64
}

type evtRow struct {
	ReceivedAt      string  `json:"received_at"`
	BatchID         string  `json:"batch_id"`
	SessionID       string  `json:"session_id"`
	PlayerIDHash    string  `json:"player_id_hash"`
	FpsAvg          float32 `json:"fps_avg"`
	PingMsAvg       float32 `json:"ping_ms_avg"`
	SpeedAvgMps     float32 `json:"speed_avg_mps"`
	SpeedMaxMps     float32 `json:"speed_max_mps"`
	TeleportSuspect uint8   `json:"teleport_suspect"`
	FlightSuspect   uint8   `json:"flight_suspect"`
	NoclipSuspect   uint8   `json:"noclip_suspect"`
	AnomalyScore    float32 `json:"anomaly_score"`
}

type incidentRow struct {
	Ts           string  `json:"ts"`
	SessionID    string  `json:"session_id"`
	PlayerIDHash string  `json:"player_id_hash"`
	Rule         string  `json:"rule"`
	Value        float32 `json:"value"`
	Details      string  `json:"details"`
}

// ------------------------------------------------------

func main() {
	chHTTP := env("CH_HTTP", "http://clickhouse:8123")
	db := env("CH_DB", "anticheat")
	chUser := env("CH_USER", "app")
	chPass := env("CH_PASS", "app")

	http.HandleFunc("/api/v1/ingest", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}

		nowStr := time.Now().UTC().Truncate(time.Second).Format("2006-01-02 15:04:05")
		batchID := req.Header.Get("X-Batch-Id")
		sessionID := req.Header.Get("X-Session-Id")
		playerID := req.Header.Get("X-Player-Id")

		// 1) пишем сырые данные
		raw := rawRow{
			BatchID:      batchID,
			SessionID:    sessionID,
			PlayerIDHash: playerID,
			ReceivedAt:   nowStr,
			PayloadBytes: base64.StdEncoding.EncodeToString(body),
		}
		if err := insertJSONEachRow(chHTTP, db, chUser, chPass, "telemetry_raw", raw); err != nil {
			log.Println("raw insert:", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		// 2) пробуем распарсить protobuf; если не получилось — просто 204
		var batch pb.TelemetryBatch
		if err := proto.Unmarshal(body, &batch); err != nil {
			log.Println("proto unmarshal:", err)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// 3) нормализованные события + простые инциденты
		for _, ev := range batch.Events {
			// безопасные чтения без дженериков
			var fpsAvg, pingAvg float32
			if ev.Perf != nil {
				fpsAvg = ev.Perf.FpsAvg
				pingAvg = ev.Perf.PingMsAvg
			}

			var spAvg, spMax float32
			var tp, fl, nc uint8
			if ev.Move != nil {
				spAvg = ev.Move.SpeedAvgMps
				spMax = ev.Move.SpeedMaxMps
				if ev.Move.TeleportSuspect {
					tp = 1
				}
				if ev.Move.FlightSuspect {
					fl = 1
				}
				if ev.Move.NoclipSuspect {
					nc = 1
				}
			}

			var anomaly float32
			var honeypot bool
			if ev.Triggers != nil {
				anomaly = ev.Triggers.AnomalyScore
				honeypot = ev.Triggers.HoneypotHit
			}

			evt := evtRow{
				ReceivedAt:      nowStr,
				BatchID:         batchID,
				SessionID:       sessionID,
				PlayerIDHash:    playerID,
				FpsAvg:          fpsAvg,
				PingMsAvg:       pingAvg,
				SpeedAvgMps:     spAvg,
				SpeedMaxMps:     spMax,
				TeleportSuspect: tp,
				FlightSuspect:   fl,
				NoclipSuspect:   nc,
				AnomalyScore:    anomaly,
			}
			_ = insertJSONEachRow(chHTTP, db, chUser, chPass, "telemetry_events", evt)

			// простые правила
			if spMax > 8.0 {
				inc := incidentRow{
					Ts:           nowStr,
					SessionID:    sessionID,
					PlayerIDHash: playerID,
					Rule:         "speedhack",
					Value:        spMax,
					Details:      fmt.Sprintf("speed_max_mps=%.2f > 8.0", spMax),
				}
				_ = insertJSONEachRow(chHTTP, db, chUser, chPass, "incidents", inc)
			}
			if honeypot {
				inc := incidentRow{
					Ts:           nowStr,
					SessionID:    sessionID,
					PlayerIDHash: playerID,
					Rule:         "honeypot_hit",
					Value:        1,
					Details:      "honeypot object accessed",
				}
				_ = insertJSONEachRow(chHTTP, db, chUser, chPass, "incidents", inc)
			}
		}

		w.WriteHeader(http.StatusNoContent)
	})

	addr := ":8080"
	log.Println("ingest listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func insertJSONEachRow(chHTTP, db, user, pass, table string, row any) error {
	j, _ := json.Marshal(row)
	q := url.QueryEscape(fmt.Sprintf("INSERT INTO %s.%s FORMAT JSONEachRow", db, table))
	u := fmt.Sprintf("%s/?query=%s", chHTTP, q)

	req, _ := http.NewRequest(http.MethodPost, u, bytes.NewReader(j))
	req.Header.Set("Content-Type", "application/json")
	if user != "" {
		req.SetBasicAuth(user, pass)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("clickhouse http %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
