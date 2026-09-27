package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSinkAcceptsAndCounts(t *testing.T) {
	mu.Lock()
	byKey = map[string]*endpointStat{}
	mu.Unlock()

	req := httptest.NewRequest("POST", "http://receive-lp1.dg.srv.nintendo.net/v1/reports", strings.NewReader("play-report-bytes"))
	req.Host = "receive-lp1.dg.srv.nintendo.net"
	rec := httptest.NewRecorder()
	sink(rec, req)

	if rec.Code != 200 || rec.Body.String() != "{}" {
		t.Fatalf("code %d body %q", rec.Code, rec.Body.String())
	}
	mu.Lock()
	s := byKey["receive-lp1.dg.srv.nintendo.net /v1/reports"]
	mu.Unlock()
	if s == nil || s.Count != 1 || s.Bytes != int64(len("play-report-bytes")) {
		t.Fatalf("stat %+v", s)
	}
}

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	sink(rec, httptest.NewRequest("GET", "/healthz", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "ok") {
		t.Fatalf("healthz %d %q", rec.Code, rec.Body.String())
	}
}
