package types

import "sync"

type ModemId = uint64
type ModemIndex = int32

type ModemList struct {
    sync sync.RWMutex
    byId map[ModemId]*ModemInfo
    byIdx map[ModemIndex]*ModemInfo
}

func NewModemList() *ModemList {
    return &ModemList{
        sync:  sync.RWMutex{},
        byId:  make(map[ModemId]*ModemInfo),
        byIdx: make(map[ModemIndex]*ModemInfo),
    }
}

func (m *ModemList) Replace(records []*ModemInfo) {
    m.sync.Lock()
    defer m.sync.Unlock()
    m.byId = make(map[ModemId]*ModemInfo)
    m.byIdx = make(map[ModemIndex]*ModemInfo)
    for _, rec := range records {
        m.byId[rec.DbId] = rec
        m.byIdx[rec.Index] = rec
    }
}

func (m *ModemList) ReplaceMap(records map[int]*ModemInfo) {
    m.sync.Lock()
    defer m.sync.Unlock()
    m.byId = make(map[ModemId]*ModemInfo)
    m.byIdx = make(map[ModemIndex]*ModemInfo)
    for _, rec := range records {
        m.byId[rec.DbId] = rec
        m.byIdx[rec.Index] = rec
    }
}


func (m *ModemList) ById(id ModemId) *ModemInfo {
    m.sync.RLock()
    defer m.sync.RUnlock()
    if rec, ok := m.byId[id]; ok {
        return rec
    } else {
        return nil
    }
}

func (m *ModemList) ByIndex(idx ModemIndex) *ModemInfo {
    m.sync.RLock()
    defer m.sync.RUnlock()
    if rec, ok := m.byIdx[idx]; ok {
        return rec
    } else {
        return nil
    }
}
