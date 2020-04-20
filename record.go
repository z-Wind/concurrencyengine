package concurrencyengine

import (
	"io/ioutil"
	"os"
	"sync"

	"github.com/pkg/errors"
)

// Record 記錄任務完成度
type Record struct {
	taskDone map[interface{}]bool
	lock     sync.RWMutex

	reqToKey func(req Request) interface{}

	jsonUnmarshal func([]byte) (map[interface{}]bool, error)
	jsonMarshal   func(map[interface{}]bool) ([]byte, error)
}

// NewRecord 建立 record
func NewRecord(reqToKey func(Request) interface{}) *Record {
	var r Record

	r.taskDone = make(map[interface{}]bool)
	r.reqToKey = reqToKey

	return &r
}

// JsonRWSetup 設定 json 以便可以 load & save
func (r *Record) JsonRWSetup(
	unmarshal func([]byte) (map[interface{}]bool, error),
	marshal func(map[interface{}]bool) ([]byte, error)) {
	r.jsonUnmarshal = unmarshal
	r.jsonMarshal = marshal
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

// Load 讀取記錄資料
func (r *Record) Load(filePath string) (map[interface{}]bool, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return r.taskDone, nil
	}

	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return r.taskDone, errors.Wrap(err, "ioutil.ReadFile")
	}

	r.taskDone, err = r.jsonUnmarshal(b)
	if err != nil {
		return r.taskDone, errors.Wrap(err, "jsonUnmarshal(")
	}

	ELog.Printf("load from %s\n", filePath)

	return r.taskDone, nil
}

// Save 儲存記錄資料
func (r *Record) Save(filePath string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	jsonBytes, err := r.jsonMarshal(r.taskDone)
	if err != nil {
		return errors.Wrap(err, "jsonMarshal")
	}

	err = ioutil.WriteFile(filePath, jsonBytes, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "ioutil.WriteFile")
	}
	ELog.LPrintf("save to %s\n", filePath)

	return nil
}
