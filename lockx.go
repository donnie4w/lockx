// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// https://github.com/donnie4w/raftx
// https://tlnet.top/wiki/raftx
// github.com/donnie4w/lockx

package lockx

import (
	"github.com/donnie4w/gofer/hashmap"
	"github.com/donnie4w/gofer/util"
	"github.com/donnie4w/gofer/uuid"
	"github.com/donnie4w/raftx"
	"github.com/donnie4w/raftx/raft"
	"sync"
)

type Mutex struct {
	raft raftx.Raftx
	mp   *hashmap.Map[int64, *muxBean]
	mux  sync.Mutex
	rmap map[string]map[int64]byte
}

func NewMutex(listen string, peers []string) *Mutex {
	mux := &Mutex{raft: raftx.NewRaftx(&raft.Config{ListenAddr: listen, PeerAddr: peers}), mp: hashmap.NewMap[int64, *muxBean](), rmap: make(map[string]map[int64]byte, 0)}
	go mux.raft.Open()
	return mux
}

type muxBean struct {
	isTry bool
	ctx   chan bool
}

func (m *Mutex) register(lockstr string, id int64, timeout uint64) {
	m.mux.Lock()
	defer m.mux.Unlock()
	v, b := m.rmap[lockstr]
	if !b {
		v = make(map[int64]byte)
	}
	v[id] = 0
	m.rmap[lockstr] = v
	if b {
		return
	}

	//监听事件
	m.raft.MemWatch([]byte(lockstr), func(key, value []byte, watchType raft.WatchType) {
		//获取锁成功与否
		if watchType == raft.ADD {
			if mb, ok := m.mp.Get(util.BytesToInt64(value)); ok {
				m.del(string(key), util.BytesToInt64(value))
				close(mb.ctx)
			}
		}
		//锁释放，阻塞代码重新获取
		if watchType == raft.DELETE {
			m.mux.Lock()
			defer m.mux.Unlock()
			if ids, b := m.rmap[string(key)]; b {
				for k := range ids {
					m.raft.MemCommand(key, util.Int64ToBytes(k), timeout, raft.MEM_PUT)
					break
				}
			}
		}
		//tryLock获取锁失败
		if watchType == raft.UPDATE {
			if mb, ok := m.mp.Get(util.BytesToInt64(value)); ok {
				if mb.isTry {
					m.del(string(key), util.BytesToInt64(value))
					mb.ctx <- true
					close(mb.ctx)
				}
			}
		}
	}, false, raft.ADD, raft.DELETE, raft.UPDATE)
}

func (m *Mutex) del(lockstr string, id int64) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.mp.Del(id)
	if v, b := m.rmap[lockstr]; b {
		delete(v, id)
	}
}

// 资源锁释放，并它的删除监听事件
func (m *Mutex) unregister(lockstr string) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if v, b := m.rmap[lockstr]; b {
		if len(v) == 1 {
			delete(m.rmap, lockstr)
			m.raft.MemUnWatch([]byte(lockstr))
		}
	}
}

func (m *Mutex) lock(lockstr string, timeout uint64, isTry bool) (r bool) {
	id := uuid.NewUUID().Int64()
	ctx := make(chan bool, 1)
	m.mp.Put(id, &muxBean{isTry: isTry, ctx: ctx})
	m.register(lockstr, id, timeout)
	m.raft.MemCommand([]byte(lockstr), util.Int64ToBytes(id), timeout, raft.MEM_PUT)
	select {
	case r = <-ctx:
	}
	return
}

// Lock 获取指定资源分布式锁，获取时返回，未获取时阻塞
func (m *Mutex) Lock(lockStr string, timeout uint64) {
	m.lock(lockStr, timeout, false)
}

// TryLock 获取指定资源分布式锁，获取时返回true，未获取时返回false
func (m *Mutex) TryLock(lockStr string, timeout uint64) (r bool) {
	return !m.lock(lockStr, timeout, true)
}

// Unlock 释放指定资源当前分布式锁
func (m *Mutex) Unlock(lockStr string) {
	m.raft.MemCommand([]byte(lockStr), nil, 0, raft.MEM_DEL)
	m.unregister(lockStr)
}
