import type { ProxyAssignStrategy, AutoAssignProxyConfig } from '@/types'

export type { ProxyAssignStrategy, AutoAssignProxyConfig }

interface ProxyCandidate {
  id: number
  account_count?: number
}

/**
 * Compute proxy assignments for a batch of accounts.
 * Mirrors backend service.AssignProxies logic for frontend-driven batch flows.
 */
export function assignProxies(
  proxies: ProxyCandidate[],
  accountCount: number,
  config: AutoAssignProxyConfig
): (number | null)[] {
  if (accountCount <= 0 || proxies.length === 0) return []

  let pool = filterPool(proxies, config.exclude_used)

  // For non-one-to-one strategies, fall back to all proxies if filtered pool is empty.
  if (pool.length === 0 && config.strategy !== 'one-to-one' && config.exclude_used) {
    pool = [...proxies]
  }

  if (pool.length === 0) return Array(accountCount).fill(null)

  switch (config.strategy) {
    case 'one-to-one':
      return assignOneToOne(pool, accountCount)
    case 'round-robin':
      return assignRoundRobin(pool, accountCount)
    case 'least-used':
      return assignLeastUsed(pool, accountCount)
    case 'random':
      return assignRandom(pool, accountCount)
    default:
      return Array(accountCount).fill(null)
  }
}

function filterPool(proxies: ProxyCandidate[], excludeUsed: boolean): ProxyCandidate[] {
  if (!excludeUsed) return [...proxies]
  return proxies.filter(p => (p.account_count ?? 0) === 0)
}

function assignOneToOne(pool: ProxyCandidate[], count: number): (number | null)[] {
  if (pool.length < count) {
    // Not enough proxies - assign what we can, rest get null.
    return Array.from({ length: count }, (_, i) => (i < pool.length ? pool[i].id : null))
  }
  return Array.from({ length: count }, (_, i) => pool[i].id)
}

function assignRoundRobin(pool: ProxyCandidate[], count: number): (number | null)[] {
  return Array.from({ length: count }, (_, i) => pool[i % pool.length].id)
}

function assignLeastUsed(pool: ProxyCandidate[], count: number): (number | null)[] {
  const candidates = pool.map(p => ({
    id: p.id,
    effectiveCount: p.account_count ?? 0
  }))

  const result: (number | null)[] = []
  for (let i = 0; i < count; i++) {
    candidates.sort((a, b) => a.effectiveCount - b.effectiveCount)
    result.push(candidates[0].id)
    candidates[0].effectiveCount++
  }
  return result
}

function assignRandom(pool: ProxyCandidate[], count: number): (number | null)[] {
  return Array.from({ length: count }, () => pool[Math.floor(Math.random() * pool.length)].id)
}
