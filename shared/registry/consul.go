package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Minimal Consul HTTP API client (no external deps).

type Consul struct {
	addr   string
	client *http.Client
}

func NewConsul(addr string) *Consul {
	return &Consul{
		addr: strings.TrimRight(addr, "/"),
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

type RegisterRequest struct {
	ID      string   `json:"ID,omitempty"`
	Name    string   `json:"Name"`
	Tags    []string `json:"Tags,omitempty"`
	Address string   `json:"Address"`
	Port    int      `json:"Port"`
	Check   *Check   `json:"Check,omitempty"`
}

type Check struct {
	GRPC                           string `json:"GRPC,omitempty"`
	HTTP                           string `json:"HTTP,omitempty"`
	TCP                            string `json:"TCP,omitempty"`
	Interval                       string `json:"Interval,omitempty"`
	Timeout                        string `json:"Timeout,omitempty"`
	DeregisterCriticalServiceAfter string `json:"DeregisterCriticalServiceAfter,omitempty"`
}

func (c *Consul) Register(ctx context.Context, req RegisterRequest) error {
	if c.addr == "" {
		return errors.New("consul addr is empty")
	}
	if req.Name == "" {
		return errors.New("service name is empty")
	}
	if req.Address == "" {
		req.Address = "127.0.0.1"
	}
	if req.Port <= 0 {
		return errors.New("invalid service port")
	}
	if req.ID == "" {
		req.ID = DefaultInstanceID(req.Name)
	}

	b, _ := json.Marshal(req)
	u := c.addr + "/v1/agent/service/register"
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(b))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("consul register failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *Consul) Deregister(ctx context.Context, serviceID string) error {
	if c.addr == "" {
		return errors.New("consul addr is empty")
	}
	if serviceID == "" {
		return errors.New("service id is empty")
	}
	u := c.addr + "/v1/agent/service/deregister/" + serviceID
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPut, u, nil)
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("consul deregister failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

type HealthServiceEntry struct {
	Service struct {
		ID      string `json:"ID"`
		Service string `json:"Service"`
		Address string `json:"Address"`
		Port    int    `json:"Port"`
	} `json:"Service"`
}

func (c *Consul) DiscoverOne(ctx context.Context, serviceName string) (string, error) {
	if c.addr == "" {
		return "", errors.New("consul addr is empty")
	}
	if serviceName == "" {
		return "", errors.New("service name is empty")
	}
	u := c.addr + "/v1/health/service/" + serviceName + "?passing=1"
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("consul discover failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var entries []HealthServiceEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", errors.New("no healthy instance found")
	}
	e := entries[0]
	addr := e.Service.Address
	if addr == "" {
		addr = "127.0.0.1"
	}
	return net.JoinHostPort(addr, strconv.Itoa(e.Service.Port)), nil
}

func DefaultInstanceID(serviceName string) string {
	host, _ := os.Hostname()
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%s-%s-%d-%d", serviceName, host, os.Getpid(), rand.Intn(1000000))
}
