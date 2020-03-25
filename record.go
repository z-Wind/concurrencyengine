package concurrencyengine

import (
	"sync"
)

// Record 記錄任務完成度
type Record struct {
	taskDone map[interface{}]bool
	lock     sync.RWMutex

	reqToKey func(req Request) interface{}
}

// NewRecord 建立 record
func NewRecord(reqToKey func(Request) interface{}) *Record {
	var r Record

	r.taskDone = make(map[interface{}]bool)
	r.reqToKey = reqToKey

	return &r
}

// IsProcessed 是否已處理，未處理就加入
// 不存在表示未處理，存在但 False 表示處理中，存在且 True 表示已處理
func (r *Record) IsProcessed(req Request) bool {
	key := r.reqToKey(req)
	r.lock.Lock()
	defer r.lock.Unlock()

	_, ok := r.taskDone[key]
	if !ok {
		r.taskDone[key] = false
	}

	return ok
}

// IsDone 是否完成
// 不存在表示未處理，存在但 False 表示處理中，存在且 True 表示已處理
func (r *Record) IsDone(req Request) bool {
	key := r.reqToKey(req)
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.taskDone[key]
}

// Done 任務已完成
func (r *Record) Done(key interface{}) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.taskDone[key] = true
}
