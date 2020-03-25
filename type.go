package concurrencyengine

// Request 需執行的任務
type Request struct {
	// 執行的參數
	Item interface{}
	// 執行的函數
	ParseFunc func(Request) (ParseResult, error)
}

// ParseResult worker 回傳的執行結果
type ParseResult struct {
	// 結果，若是 nil 不會傳進 channel
	Item interface{}
	// 新增的任務
	ExtraRequests []Request
	// 需重作的原任務
	RedoRequests []Request
	// 記錄任務是否已完成
	Done bool
}

// Scheduler 調配工作
type Scheduler interface {
	// 提交任務
	Submit(Request)
	// 將空閒的 worker 排進序列
	WorkerReady(chan Request)
	Run()
}

// Recorder 記錄工作
type Recorder interface {
	// 是否處理中
	IsProcessed(Request) bool
	// 是否已完成
	IsDone(req Request) bool
	// 將任務設定已完成
	Done(key interface{})
}
