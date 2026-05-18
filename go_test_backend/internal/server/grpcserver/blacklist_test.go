package grpcserver

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go_test/shared/registry"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestParseUserIDs_JSONArray(t *testing.T) {
	got := parseUserIDs(`["1","2","2"," 3 "]`)
	if len(got) != 3 || got[0] != "1" || got[1] != "2" || got[2] != "3" {
		t.Fatalf("got=%v", got)
	}
}

func TestParseUserIDs_Split(t *testing.T) {
	got := parseUserIDs("1, 2\n3\t4;5")
	if len(got) != 5 {
		t.Fatalf("got=%v", got)
	}
}

func TestConsulUserBlacklist_Cache(t *testing.T) {
	var val atomic.Value
	val.Store("1,2")

	cli := registry.NewConsulWithClient("http://consul", &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if !strings.HasPrefix(r.URL.String(), "http://consul/v1/kv/go_test/user/blacklist") {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Status:     "404 Not Found",
					Body:       io.NopCloser(strings.NewReader("")),
					Header:     make(http.Header),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(strings.NewReader(val.Load().(string))),
				Header:     make(http.Header),
			}, nil
		}),
		Timeout: time.Second,
	})

	bl := NewConsulUserBlacklist(cli, "go_test/user/blacklist", 200*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ok, err := bl.IsBlacklisted(ctx, "1")
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if !ok {
		t.Fatalf("expected blacklisted")
	}

	// update underlying KV, but still within ttl -> should keep cached result
	val.Store("3")
	ok, err = bl.IsBlacklisted(ctx, "1")
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if !ok {
		t.Fatalf("expected cached blacklisted")
	}

	time.Sleep(250 * time.Millisecond)

	ok, err = bl.IsBlacklisted(ctx, "1")
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if ok {
		t.Fatalf("expected refreshed not blacklisted")
	}
}
