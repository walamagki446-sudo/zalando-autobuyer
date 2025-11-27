package main

import (
	"fmt"
	"sync"
)

// ProxyManager handles proxy rotation and management
type ProxyManager struct {
	proxies     []string
	currentIdx  int
	mutex       sync.RWMutex
	failedCount map[string]int
}

// NewProxyManager creates a new proxy manager
func NewProxyManager() *ProxyManager {
	return &ProxyManager{
		proxies:     make([]string, 0),
		failedCount: make(map[string]int),
	}
}

// AddProxy adds a proxy to the manager
func (pm *ProxyManager) AddProxy(proxy string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.proxies = append(pm.proxies, proxy)
}

// GetProxy returns the current proxy
func (pm *ProxyManager) GetProxy() string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	if len(pm.proxies) == 0 {
		return ""
	}

	return pm.proxies[pm.currentIdx%len(pm.proxies)]
}

// RotateProxy moves to the next proxy
func (pm *ProxyManager) RotateProxy() string {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if len(pm.proxies) == 0 {
		return ""
	}

	pm.currentIdx = (pm.currentIdx + 1) % len(pm.proxies)
	return pm.proxies[pm.currentIdx]
}

// MarkFailed marks a proxy as failed
func (pm *ProxyManager) MarkFailed(proxy string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.failedCount[proxy]++
}

// GetFailCount returns the fail count for a proxy
func (pm *ProxyManager) GetFailCount(proxy string) int {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.failedCount[proxy]
}

// RemoveProxy removes a proxy from the manager
func (pm *ProxyManager) RemoveProxy(proxy string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	for i, p := range pm.proxies {
		if p == proxy {
			pm.proxies = append(pm.proxies[:i], pm.proxies[i+1:]...)
			delete(pm.failedCount, proxy)
			if pm.currentIdx >= len(pm.proxies) && len(pm.proxies) > 0 {
				pm.currentIdx = pm.currentIdx % len(pm.proxies)
			}
			return
		}
	}
}

// Count returns the number of proxies
func (pm *ProxyManager) Count() int {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return len(pm.proxies)
}

// Clear removes all proxies
func (pm *ProxyManager) Clear() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.proxies = make([]string, 0)
	pm.failedCount = make(map[string]int)
	pm.currentIdx = 0
}

// ParseProxyString parses a proxy string in format: host:port:username:password
func ParseProxyString(proxyStr string) (host, port, username, password string, err error) {
	parts := splitProxyString(proxyStr)

	if len(parts) == 4 {
		return parts[0], parts[1], parts[2], parts[3], nil
	} else if len(parts) == 2 {
		return parts[0], parts[1], "", "", nil
	}

	return "", "", "", "", fmt.Errorf("invalid proxy format")
}

// DefaultProxy is an example proxy format placeholder
// Users should replace this with their own proxy credentials
const DefaultProxy = ""
