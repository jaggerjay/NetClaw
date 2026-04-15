package proxy

import (
	"strings"
	"sync"
	"time"
)

type fallbackCache struct {
	mu    sync.RWMutex
	items map[string]time.Time
}

func newFallbackCache() *fallbackCache {
	return &fallbackCache{items: make(map[string]time.Time)}
}

func (c *fallbackCache) Mark(host string, ttl time.Duration) {
	host = normalizeFallbackHost(host)
	if host == "" || ttl <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[host] = time.Now().Add(ttl)
}

func (c *fallbackCache) Active(host string) bool {
	host = normalizeFallbackHost(host)
	if host == "" {
		return false
	}
	c.mu.RLock()
	expiry, ok := c.items[host]
	c.mu.RUnlock()
	if !ok {
		return false
	}
	if time.Now().Before(expiry) {
		return true
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if expiry2, ok := c.items[host]; ok && !time.Now().Before(expiry2) {
		delete(c.items, host)
	}
	return false
}

func (s *Server) markMITMFailure(host string) {
	if s == nil || s.fallbackCache == nil || s.config.MITMFailureBackoff <= 0 {
		return
	}
	s.fallbackCache.Mark(host, s.config.MITMFailureBackoff)
}

func (s *Server) shouldTemporarilyBypassMITM(host string) bool {
	if s == nil || s.fallbackCache == nil {
		return false
	}
	return s.fallbackCache.Active(host)
}

func normalizeFallbackHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "" {
		return ""
	}
	if parsedHost, _ := splitHostPort(host); parsedHost != "" {
		return parsedHost
	}
	return host
}
