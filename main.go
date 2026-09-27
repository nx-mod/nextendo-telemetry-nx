// Command nextendo-telemetry-nx is the telemetry sink for Nextendo Network.
//
// Switch titles and the system upload telemetry to several Nintendo services:
// prepo / play-report (receive-*.dg.srv.nintendo.net), erpt / error-report
// (receive-*.er.srv.nintendo.net) and assorted analytics ("dragons" and
// friends). On the Nextendo stack these fall through to Nintendo; a game that
// insists on a successful upload can stall or retry until it does. This accepts
// every telemetry POST, counts it, optionally samples the body for research,
// discards it, and answers success — so nothing waits on Nintendo.
//
// It is deliberately permissive: any method, any path, always a 2xx. The
// dashboard shows what was received, grouped by host and path, which doubles as
// a map of what each title reports.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	httpPort  = envOrInt("TELEMETRY_PORT", 8472)
	certFile  = envOr("CERT_FILE", "")
	keyFile   = envOr("KEY_FILE", "")
	dumpDir   = envOr("TELEMETRY_DUMP", "") // "" = don't keep bodies
	maxBody   = int64(envOrInt("TELEMETRY_MAX_BODY", 1<<20))
	dashPort  = envOr("DASH_PORT", "8101")
	dashToken = envOr("DASH_TOKEN", "")

	total     atomic.Int64
	bytesRcvd atomic.Int64
	dashStart = time.Now()

	mu    sync.Mutex
	byKey = map[string]*endpointStat{} // "host path" -> stat
)

type endpointStat struct {
	Host  string `json:"host"`
	Path  string `json:"path"`
	Count int64  `json:"count"`
	Bytes int64  `json:"bytes"`
	Last  int64  `json:"lastUnix"`
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func envOrInt(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return d
}

func main() {
	log.SetOutput(os.Stdout)
	if dumpDir != "" {
		_ = os.MkdirAll(dumpDir, 0o755)
	}
	go startDashboard()

	mux := http.NewServeMux()
	mux.HandleFunc("/", sink)

	addr := fmt.Sprintf(":%d", httpPort)
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 15 * time.Second}
	if certFile != "" && keyFile != "" {
		log.Printf("[Telemetry] listening HTTPS %s (dump=%q)", addr, dumpDir)
		log.Fatal(srv.ListenAndServeTLS(certFile, keyFile))
	}
	log.Printf("[Telemetry] listening HTTP %s (TLS via sni-router; dump=%q)", addr, dumpDir)
	log.Fatal(srv.ListenAndServe())
}

// sink accepts any telemetry request and answers success.
func sink(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/healthz" {
		fmt.Fprintln(w, "ok")
		return
	}
	body, _ := io.ReadAll(io.LimitReader(r.Body, maxBody))
	r.Body.Close()

	total.Add(1)
	bytesRcvd.Add(int64(len(body)))
	record(r.Host, r.URL.Path, int64(len(body)))
	if dumpDir != "" && len(body) > 0 {
		dump(r, body)
	}

	// Most Nintendo telemetry endpoints are satisfied by a 200; erpt/prepo accept
	// an empty body. Answer JSON {} so a JSON-expecting client still parses it.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{}"))
}

func record(host, path string, n int64) {
	key := host + " " + path
	now := time.Now().Unix()
	mu.Lock()
	s := byKey[key]
	if s == nil {
		s = &endpointStat{Host: host, Path: path}
		byKey[key] = s
	}
	s.Count++
	s.Bytes += n
	s.Last = now
	mu.Unlock()
}

func dump(r *http.Request, body []byte) {
	name := fmt.Sprintf("%d_%s.bin", time.Now().UnixNano(),
		replaceAll(r.Host+r.URL.Path, "/:?&=", '_'))
	_ = os.WriteFile(filepath.Join(dumpDir, name), body, 0o644)
}

func replaceAll(s, set string, with byte) string {
	b := []byte(s)
	for i, c := range b {
		for _, bad := range []byte(set) {
			if byte(c) == bad {
				b[i] = with
			}
		}
	}
	return string(b)
}

func startDashboard() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		if dashToken != "" && r.URL.Query().Get("key") != dashToken {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		mu.Lock()
		eps := make([]endpointStat, 0, len(byKey))
		for _, s := range byKey {
			eps = append(eps, *s)
		}
		mu.Unlock()
		sort.Slice(eps, func(i, j int) bool { return eps[i].Count > eps[j].Count })
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"uptimeSeconds": int(time.Since(dashStart).Seconds()),
			"received":      total.Load(),
			"bytes":         bytesRcvd.Load(),
			"endpoints":     eps,
			"stack":         "telemetry",
		})
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "ok") })
	log.Printf("[Telemetry Dashboard] :%s", dashPort)
	if err := http.ListenAndServe(":"+dashPort, mux); err != nil {
		log.Printf("[Telemetry Dashboard] %v", err)
	}
}
