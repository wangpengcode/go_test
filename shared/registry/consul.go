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

// 极简版 Consul HTTP API 客户端（不依赖第三方库）。

type Consul struct {
	addr   string
	client *http.Client
}

// NewConsul 创建一个极简的 Consul HTTP API 客户端。
//
// addr 例子："http://127.0.0.1:8500"。
func NewConsul(addr string) *Consul {
	return &Consul{
		addr: strings.TrimRight(addr, "/"),
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

// NewConsulWithClient 与 NewConsul 类似，但允许注入自定义 http.Client（便于测试/自定义）。
func NewConsulWithClient(addr string, client *http.Client) *Consul {
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	return &Consul{
		addr:   strings.TrimRight(addr, "/"),
		client: client,
	}
}

// GetKVRaw 从 Consul KV 读取一个 key 的原始字符串内容（使用 ?raw）。
//
// ok=false 表示 key 不存在（Consul 返回 404）。
func (c *Consul) GetKVRaw(ctx context.Context, key string) (val string, ok bool, err error) {
	if c.addr == "" {
		return "", false, errors.New("consul addr is empty")
	}
	key = strings.TrimLeft(key, "/")
	if key == "" {
		return "", false, errors.New("consul kv key is empty")
	}

	u := c.addr + "/v1/kv/" + key + "?raw=1"
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", false, nil
	}
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(resp.Body)
		return "", false, fmt.Errorf("consul kv get failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false, err
	}
	return string(body), true, nil
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

// Register 向 Consul 注册（或更新）一个服务实例。
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

// Deregister 从 Consul 注销一个服务实例（按 serviceID）。
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

// DiscoverOne 从 Consul 里找一个“健康”的 serviceName 实例，并返回 "host:port"。
//
// 给刚接触 Go 的同学：
// - Consul 的 health API 返回的是实例列表；这里为了简单，取第一个。
// - 后续你可以扩展为负载均衡（随机、轮询等）。
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

// DefaultInstanceID 生成一个“尽量不重复”的实例 ID 字符串。
func DefaultInstanceID(serviceName string) string {
	host, _ := os.Hostname()
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%s-%s-%d-%d", serviceName, host, os.Getpid(), rand.Intn(1000000))
}
