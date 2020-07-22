package db

import (
	"github.com/sedl/docsis-pnm/internal/types"
	"sync"
)

type CMTSUpstreamCache struct {
	cache map[int]*types.CMTSUpstreamRecord
	cacheDescr map[string]*types.CMTSUpstreamRecord
	lock sync.RWMutex
}

func NewCMTSUpstreamCache() *CMTSUpstreamCache {
	return &CMTSUpstreamCache{
		cache:      make(map[int]*types.CMTSUpstreamRecord),
		cacheDescr: make(map[string]*types.CMTSUpstreamRecord),
		lock:       sync.RWMutex{},
	}
}

func (c *CMTSUpstreamCache) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.cache)
}

func (c *CMTSUpstreamCache) GetByIndex(idx int) *types.CMTSUpstreamRecord {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if val, found := c.cache[idx]; found {
		return val
	}

	return nil
}

func (c *CMTSUpstreamCache) GetByDescr(descr string) *types.CMTSUpstreamRecord {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if val, found := c.cacheDescr[descr]; found {
		return val
	}

	return nil
}

func (c *CMTSUpstreamCache) Add(record *types.CMTSUpstreamRecord) {
	c.lock.Lock()
	c.cache[int(record.SNMPIndex)] = record
	c.cacheDescr[record.Description] = record
	c.lock.Unlock()
}
