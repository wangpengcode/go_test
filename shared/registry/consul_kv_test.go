package registry

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestConsul_GetKVRaw_OK(t *testing.T) {
	cli := NewConsulWithClient("http://consul", &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodGet {
				t.Fatalf("method=%s", r.Method)
			}
			if r.URL.String() != "http://consul/v1/kv/go_test/user/blacklist?raw=1" {
				t.Fatalf("url=%s", r.URL.String())
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(strings.NewReader("1,2,3")),
				Header:     make(http.Header),
			}, nil
		}),
		Timeout: time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	v, ok, err := cli.GetKVRaw(ctx, "go_test/user/blacklist")
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if !ok {
		t.Fatalf("ok=false")
	}
	if v != "1,2,3" {
		t.Fatalf("val=%q", v)
	}
}

func TestConsul_GetKVRaw_NotFound(t *testing.T) {
	cli := NewConsulWithClient("http://consul", &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Status:     "404 Not Found",
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}, nil
		}),
		Timeout: time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, ok, err := cli.GetKVRaw(ctx, "missing")
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if ok {
		t.Fatalf("ok=true")
	}
}
