package git

import (
	"container/list"
	"sync"
	"time"
)

// CachedRepo wraps a Repository with cache metadata
type CachedRepo struct {
	*Repository
	lastAccess time.Time
	path       string
}

// RepoPool provides LRU-cached repository access
type RepoPool struct {
	mu       sync.RWMutex
	repos    map[string]*list.Element
	lru      *list.List
	maxRepos int
	ttl      time.Duration
}

// NewRepoPool creates a new repository pool with the given capacity
func NewRepoPool(maxRepos int, ttl time.Duration) *RepoPool {
	if maxRepos <= 0 {
		maxRepos = 10
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &RepoPool{
		repos:    make(map[string]*list.Element),
		lru:      list.New(),
		maxRepos: maxRepos,
		ttl:      ttl,
	}
}

// Get retrieves a repository from the cache or opens it
func (p *RepoPool) Get(path string) (*Repository, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if already cached
	if elem, ok := p.repos[path]; ok {
		cached := elem.Value.(*CachedRepo)
		// Check if still valid (TTL not expired)
		if time.Since(cached.lastAccess) < p.ttl {
			cached.lastAccess = time.Now()
			p.lru.MoveToFront(elem)
			return cached.Repository, nil
		}
		// TTL expired, remove from cache
		p.removeElement(elem)
	}

	// Open repository
	repo, err := Open(path)
	if err != nil {
		return nil, err
	}

	// Add to cache
	p.addToCache(path, repo)

	return repo, nil
}

// addToCache adds a repository to the cache, evicting oldest if necessary
func (p *RepoPool) addToCache(path string, repo *Repository) {
	// Evict oldest entries if at capacity
	for p.lru.Len() >= p.maxRepos {
		oldest := p.lru.Back()
		if oldest != nil {
			p.removeElement(oldest)
		}
	}

	cached := &CachedRepo{
		Repository: repo,
		lastAccess: time.Now(),
		path:       path,
	}

	elem := p.lru.PushFront(cached)
	p.repos[path] = elem
}

// removeElement removes an element from the cache
func (p *RepoPool) removeElement(elem *list.Element) {
	cached := elem.Value.(*CachedRepo)
	delete(p.repos, cached.path)
	p.lru.Remove(elem)
}

// Invalidate removes a specific repository from the cache
func (p *RepoPool) Invalidate(path string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if elem, ok := p.repos[path]; ok {
		p.removeElement(elem)
	}
}

// Clear removes all repositories from the cache
func (p *RepoPool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.repos = make(map[string]*list.Element)
	p.lru.Init()
}

// Size returns the current number of cached repositories
func (p *RepoPool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lru.Len()
}

// Stats returns cache statistics
func (p *RepoPool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return PoolStats{
		Size:     p.lru.Len(),
		Capacity: p.maxRepos,
		TTL:      p.ttl,
	}
}

// PoolStats contains cache statistics
type PoolStats struct {
	Size     int
	Capacity int
	TTL      time.Duration
}

// DefaultPool is the global repository pool
var DefaultPool = NewRepoPool(10, 5*time.Minute)

// OpenCached opens a repository using the default pool
func OpenCached(path string) (*Repository, error) {
	return DefaultPool.Get(path)
}
