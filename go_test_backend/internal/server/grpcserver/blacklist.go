package grpcserver

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"go_test/shared/registry"
)

type UserBlacklistChecker interface {
	// IsBlacklisted returns ok=true when userID is in blacklist.
	IsBlacklisted(ctx context.Context, userID string) (ok bool, err error)
}

type ConsulUserBlacklist struct {
	consul *registry.Consul
	key    string
	ttl    time.Duration

	mu        sync.RWMutex
	expiresAt time.Time
	set       map[string]struct{}
}

func NewConsulUserBlacklist(consul *registry.Consul, key string, ttl time.Duration) *ConsulUserBlacklist {
	return &ConsulUserBlacklist{
		consul: consul,
		key:    strings.TrimSpace(key),
		ttl:    ttl,
	}
}

func (b *ConsulUserBlacklist) IsBlacklisted(ctx context.Context, userID string) (bool, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" || b == nil {
		return false, nil
	}
	if b.consul == nil || b.key == "" {
		return false, nil
	}

	now := time.Now()
	b.mu.RLock()
	if b.set != nil && now.Before(b.expiresAt) {
		_, ok := b.set[userID]
		b.mu.RUnlock()
		return ok, nil
	}
	b.mu.RUnlock()

	if err := b.refresh(ctx, now); err != nil {
		return false, err
	}

	b.mu.RLock()
	_, ok := b.set[userID]
	b.mu.RUnlock()
	return ok, nil
}

func (b *ConsulUserBlacklist) refresh(ctx context.Context, now time.Time) error {
	// double-check to reduce duplicate refresh under contention
	b.mu.RLock()
	if b.set != nil && now.Before(b.expiresAt) {
		b.mu.RUnlock()
		return nil
	}
	b.mu.RUnlock()

	val, ok, err := b.consul.GetKVRaw(ctx, b.key)
	if err != nil {
		return err
	}

	next := make(map[string]struct{})
	if ok {
		for _, id := range parseUserIDs(val) {
			next[id] = struct{}{}
		}
	}

	ttl := b.ttl
	if ttl <= 0 {
		ttl = 5 * time.Second
	}

	b.mu.Lock()
	b.set = next
	b.expiresAt = now.Add(ttl)
	b.mu.Unlock()
	return nil
}

func parseUserIDs(val string) []string {
	val = strings.TrimSpace(val)
	if val == "" {
		return nil
	}

	// 兼容 JSON 数组（["1","2"] 或 [1,2]）。
	if strings.HasPrefix(val, "[") {
		var sids []string
		if err := json.Unmarshal([]byte(val), &sids); err == nil {
			return normalizeIDs(sids)
		}
		var iids []any
		if err := json.Unmarshal([]byte(val), &iids); err == nil {
			out := make([]string, 0, len(iids))
			for _, v := range iids {
				switch t := v.(type) {
				case string:
					out = append(out, t)
				case float64:
					out = append(out, strings.TrimRight(strings.TrimRight(strconv.FormatFloat(t, 'f', -1, 64), "0"), "."))
				default:
					// ignore
				}
			}
			return normalizeIDs(out)
		}
	}

	// 兼容逗号、空格、换行分隔。
	parts := strings.FieldsFunc(val, func(r rune) bool {
		switch r {
		case ',', ';', '\n', '\r', '\t', ' ':
			return true
		default:
			return false
		}
	})
	return normalizeIDs(parts)
}

func normalizeIDs(in []string) []string {
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
