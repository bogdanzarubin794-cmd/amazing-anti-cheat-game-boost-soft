package main

import (
"encoding/base64"
"encoding/json"
"fmt"
"io"
"log"
"net/http"
"net/url"
"os"
"time"
)

type row struct {
BatchID      string    `json:"batch_id"`
SessionID    string    `json:"session_id"`
PlayerIDHash string    `json:"player_id_hash"`
ReceivedAt   time.Time `json:"received_at"`
PayloadBytes string    `json:"payload_bytes"`
}

func (r row) MarshalJSON() ([]byte, error) {
type alias row
a := struct {
alias
ReceivedAt string `json:"received_at"`
}{
alias:      (alias)(r),
ReceivedAt: r.ReceivedAt.Format("2006-01-02 15:04:05"),
}
return json.Marshal(a)
}

func main() {
chHTTP := env("CH_HTTP", "http://clickhouse:8123")
db := env("CH_DB", "anticheat")
chUser := env("CH_USER", "default")
chPass := env("CH_PASS", "")

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
data := row{
BatchID:      req.Header.Get("X-Batch-Id"),
SessionID:    req.Header.Get("X-Session-Id"),
PlayerIDHash: req.Header.Get("X-Player-Id"),
ReceivedAt:   time.Now().UTC().Truncate(time.Second),
PayloadBytes: base64.StdEncoding.EncodeToString(body),
}
j, _ := json.Marshal(data)

q := url.QueryEscape(fmt.Sprintf("INSERT INTO %s.telemetry_raw FORMAT JSONEachRow", db))
insertURL := fmt.Sprintf("%s/?query=%s", chHTTP, q)

httpReq, _ := http.NewRequest(http.MethodPost, insertURL, bytesReader(j))
httpReq.Header.Set("Content-Type", "application/json")
if chUser != "" {
httpReq.SetBasicAuth(chUser, chPass)
}
resp, err := http.DefaultClient.Do(httpReq)
if err != nil {
log.Println("insert error:", err)
http.Error(w, "db error", http.StatusInternalServerError)
return
}
defer resp.Body.Close()
if resp.StatusCode >= 300 {
b, _ := io.ReadAll(resp.Body)
log.Printf("clickhouse http status=%d body=%s\n", resp.StatusCode, string(b))
http.Error(w, "db error", http.StatusInternalServerError)
return
}
w.WriteHeader(http.StatusNoContent)
})

addr := ":8080"
log.Println("ingest listening on", addr)
log.Fatal(http.ListenAndServe(addr, nil))
}

func bytesReader(b []byte) io.Reader { return &rdr{b, 0} }
type rdr struct{ b []byte; i int }
func (r *rdr) Read(p []byte) (int, error) {
if r.i >= len(r.b) { return 0, io.EOF }
n := copy(p, r.b[r.i:])
r.i += n
return n, nil
}

func env(k, d string) string { if v := os.Getenv(k); v != "" { return v }; return d }