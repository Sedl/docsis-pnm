package db

import (
	"github.com/sedl/docsis-pnm/internal/types"
	"sync"
)

type CMTSCache struct {
	cache     map[string]*types.CMTSRecord
	lock sync.RWMutex
}

func NewCMTSCache() *CMTSCache {
	return &CMTSCache{
		cache:     make(map[string]*types.CMTSRecord),
		lock: sync.RWMutex{},
	}
}

func (m *CMTSCache) Get(hostname string) *types.CMTSRecord {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if val, ok := m.cache[hostname]; ok {
		return val
	}
	return nil
}

func (m *CMTSCache) Add(rec *types.CMTSRecord) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.cache[rec.Hostname] = rec
}
