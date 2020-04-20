package concurrencyengine

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/pkg/errors"
)

func Test_record(t *testing.T) {
	tmpPath := "temp"
	if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
		os.MkdirAll(tmpPath, os.ModePerm)
	}

	req := Request{
		Item: "key",
	}

	reqToKey := func(req Request) interface{} { return req.Item.(string) }
	r := NewRecord(reqToKey)

	jsonUnmarshal := func(b []byte) (map[interface{}]bool, error) {
		result := make(map[interface{}]bool)

		m := make(map[string]bool)
		err := json.Unmarshal(b, &m)
		if err != nil {
			return result, errors.Wrap(err, "json.Unmarshal(")
		}

		for key, item := range m {
			result[key] = item

		}
		return result, nil
	}
	jsonMarshal := func(m map[interface{}]bool) ([]byte, error) {
		result := make(map[string]bool)

		for key, item := range m {
			result[key.(string)] = item
		}
		jsonBytes, err := json.Marshal(&result)
		if err != nil {
			return nil, errors.Wrap(err, "json.Marshal")
		}
		return jsonBytes, nil
	}
	r.JsonRWSetup(jsonUnmarshal, jsonMarshal)

	if got := r.IsProcessed(req); got {
		t.Errorf("r.IsProcessed() = %v, want %v", got, false)
	}
	r.Done(req.Item.(string))
	if got := r.IsProcessed(req); !got {
		t.Errorf("r.IsProcessed() = %v, want %v", got, true)
	}
	if err := r.Save(path.Join(tmpPath, "temp-record.dat")); err != nil {
		t.Errorf("r.Save() error = %v, wantErr %v", err, nil)
	}

	r = NewRecord(reqToKey)
	r.JsonRWSetup(jsonUnmarshal, jsonMarshal)
	if _, err := r.Load(path.Join(tmpPath, "temp-record.dat")); err != nil {
		t.Errorf("r.Load() error = %v, wantErr %v", err, nil)
	}
	req2 := Request{
		Item: "key",
	}
	fmt.Printf("%+v\n", r.taskDone)
	if got := r.IsProcessed(req2); !got {
		t.Errorf("r.IsProcessed() = %v, want %v", got, true)
	}

	// 移除暫存檔
	if err := os.RemoveAll(tmpPath); err != nil {
		t.Error(err)
	}
}
