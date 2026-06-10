package service

import (
	"fmt"
	"math/rand/v2"
	"sort"
)

// ProxyAssignStrategy defines the strategy for auto-assigning proxies to accounts.
type ProxyAssignStrategy string

const (
	ProxyAssignOneToOne   ProxyAssignStrategy = "one-to-one"
	ProxyAssignRoundRobin ProxyAssignStrategy = "round-robin"
	ProxyAssignLeastUsed  ProxyAssignStrategy = "least-used"
	ProxyAssignRandom     ProxyAssignStrategy = "random"
)

// AutoAssignProxyConfig is the user-facing configuration for auto-assign proxy.
type AutoAssignProxyConfig struct {
	Strategy    ProxyAssignStrategy `json:"strategy"`
	ExcludeUsed bool                `json:"exclude_used"`
}

// ProxyAssignInput holds all data needed by AssignProxies.
type ProxyAssignInput struct {
	Strategy     ProxyAssignStrategy
	Proxies      []ProxyWithAccountCount
	AccountCount int
	ExcludeUsed  bool
}

// AssignProxies returns a slice of proxy IDs (one per account).
// The caller must pre-fetch active proxies with account counts.
func AssignProxies(input ProxyAssignInput) ([]*int64, error) {
	if input.AccountCount <= 0 {
		return nil, nil
	}

	pool := filterProxyPool(input.Proxies, input.ExcludeUsed)

	// For non-one-to-one strategies, fall back to all proxies if filtered pool is empty.
	if len(pool) == 0 && input.Strategy != ProxyAssignOneToOne && input.ExcludeUsed {
		pool = filterActiveProxies(input.Proxies)
	}

	if len(pool) == 0 {
		return nil, fmt.Errorf("no available proxies for auto-assignment")
	}

	switch input.Strategy {
	case ProxyAssignOneToOne:
		return assignOneToOne(pool, input.AccountCount)
	case ProxyAssignRoundRobin:
		return assignRoundRobin(pool, input.AccountCount), nil
	case ProxyAssignLeastUsed:
		return assignLeastUsed(pool, input.AccountCount), nil
	case ProxyAssignRandom:
		return assignRandom(pool, input.AccountCount), nil
	default:
		return nil, fmt.Errorf("unknown proxy assign strategy: %s", input.Strategy)
	}
}

func filterProxyPool(proxies []ProxyWithAccountCount, excludeUsed bool) []ProxyWithAccountCount {
	if !excludeUsed {
		return filterActiveProxies(proxies)
	}
	var result []ProxyWithAccountCount
	for _, p := range proxies {
		if p.AccountCount == 0 {
			result = append(result, p)
		}
	}
	return result
}

func filterActiveProxies(proxies []ProxyWithAccountCount) []ProxyWithAccountCount {
	// Return a copy to avoid mutating the input.
	result := make([]ProxyWithAccountCount, len(proxies))
	copy(result, proxies)
	return result
}

func assignOneToOne(pool []ProxyWithAccountCount, count int) ([]*int64, error) {
	if len(pool) < count {
		return nil, fmt.Errorf(
			"not enough proxies for one-to-one assignment: need %d, available %d",
			count, len(pool),
		)
	}
	result := make([]*int64, count)
	for i := 0; i < count; i++ {
		id := pool[i].ID
		result[i] = &id
	}
	return result, nil
}

func assignRoundRobin(pool []ProxyWithAccountCount, count int) []*int64 {
	result := make([]*int64, count)
	for i := 0; i < count; i++ {
		id := pool[i%len(pool)].ID
		result[i] = &id
	}
	return result
}

func assignLeastUsed(pool []ProxyWithAccountCount, count int) []*int64 {
	// Track effective count: original account_count + assignments in this batch.
	type candidate struct {
		id             int64
		effectiveCount int64
	}
	candidates := make([]candidate, len(pool))
	for i, p := range pool {
		candidates[i] = candidate{id: p.ID, effectiveCount: p.AccountCount}
	}

	result := make([]*int64, count)
	for i := 0; i < count; i++ {
		// Find the candidate with the lowest effective count.
		sort.Slice(candidates, func(a, b int) bool {
			return candidates[a].effectiveCount < candidates[b].effectiveCount
		})
		id := candidates[0].id
		result[i] = &id
		candidates[0].effectiveCount++
	}
	return result
}

func assignRandom(pool []ProxyWithAccountCount, count int) []*int64 {
	result := make([]*int64, count)
	for i := 0; i < count; i++ {
		idx := rand.IntN(len(pool))
		id := pool[idx].ID
		result[i] = &id
	}
	return result
}
