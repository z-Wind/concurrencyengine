package concurrencyengine

import (
	"context"

	"golang.org/x/time/rate"
)

func NewQueueScheduler(ctx context.Context) *QueueScheduler {
	lim := rate.NewLimiter(rate.Inf, 1)

	return &QueueScheduler{Ctx: ctx, Limit: lim}
}

// QueueScheduler 分配 request 給 worker
type QueueScheduler struct {
	workerChan  chan chan Request
	requestChan chan Request
	Ctx         context.Context
	Limit       *rate.Limiter
}

// Submit 提交任務
func (s *QueueScheduler) Submit(r Request) {
	select {
	case s.requestChan <- r:
		ELog.LPrintf("%-30s s.requestChan <- r %+v\n", "Scheduler.Submit", r.Item)
	case <-s.Ctx.Done():
		ELog.LPrintf("%-30s QueueScheduler.Submit.Done\n", "Scheduler.Submit")
	}
}

// SetLimit 設定 worker/sec, 預設為 rate.Inf
func (s *QueueScheduler) SetLimit(r float64) {
	s.Limit.SetLimit(rate.Limit(r))
}

// WorkerArrange 將 worker 排進序列
func (s *QueueScheduler) WorkerArrange(w chan Request) {
	select {
	case s.workerChan <- w:
		ELog.LPrintf("%-30s s.workerChan <- worker(%v)\n", "Scheduler.WorkerReady", w)
	case <-s.Ctx.Done():
		ELog.LPrintf("%-30s QueueScheduler.WorkerReady.Done\n", "Scheduler.WorkerReady")
	}
}

// WorkerReady 將空閒的 worker 排進序列
func (s *QueueScheduler) WorkerReady(w chan Request) {
	err := s.Limit.Wait(s.Ctx)
	if err != nil {
		ELog.Fatalf("WorkerReady: s.Limit.Wait: err: %s\n", err)
	}

	s.WorkerArrange(w)
}

// Run 執行調配
func (s *QueueScheduler) Run() {
	s.requestChan = make(chan Request)
	s.workerChan = make(chan chan Request)

	go func() {
		// 用 queue 先存起來，防止阻塞
		var requestQ []Request
		var workerQ []chan Request

		for {
			var activeRequest Request
			// channel 初值為 nil，並不會觸發 select，除非賦於值
			var activeWorker chan<- Request
			if len(requestQ) > 0 && len(workerQ) > 0 {
				activeRequest = requestQ[0]
				activeWorker = workerQ[0]
			}

			select {
			case activeWorker <- activeRequest:
				ELog.LPrintf("%-30s Worker(%v) <- Request(%+v)\n", "Scheduler.Run", activeWorker, activeRequest.Item)
				requestQ = requestQ[1:]
				workerQ = workerQ[1:]
			case r := <-s.requestChan:
				ELog.LPrintf("%-30s Get Request(%+v)\n", "Scheduler.Run", r.Item)
				requestQ = append(requestQ, r)
			case w := <-s.workerChan:
				ELog.LPrintf("%-30s Worker(%v) Free\n", "Scheduler.Run", w)
				workerQ = append(workerQ, w)
			case <-s.Ctx.Done():
				ELog.LPrintf("%-30s QueueScheduler.Run.Done\n", "Scheduler.Run")
				return
			}
		}
	}()
}
