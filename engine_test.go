package concurrencyengine

import (
	"context"
	"fmt"
	"os"
	"time"
)

func ExampleConcurrencyEngine() {
	// set log
	ELog.Start("engine.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	ELog.SetFlags(0)
	defer ELog.Stop()

	// set ctx
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// set reqToKey for record mapping key
	reqToKey := func(req Request) interface{} { return req.Item.(int) }

	// create engine
	e := New(ctx, 10, reqToKey)
	e.Scheduler.SetLimit(2.0)

	// set simple parse function for processing request
	// when ParseResult.Item is nil, you would not get the item
	parseFunc := func(req Request) (ParseResult, error) {
		parseResult := ParseResult{
			Item:          nil,
			ExtraRequests: []Request{},
			RedoRequests:  []Request{},
			Done:          false,
		}

		n := req.Item.(int)

		//if n%2 == 0 {
		//	parseResult.Done = true
		//	parseResult.ExtraRequests = append(parseResult.ExtraRequests, Request{n + 11, req.ParseFunc})

		//	return parseResult, nil
		//}

		parseResult.Item = n
		parseResult.Done = true

		return parseResult, nil
	}

	// create initial requests
	requests := []Request{}
	for i := 0; i < 10; i++ {
		requests = append(requests, Request{
			Item:      i,
			ParseFunc: parseFunc,
		})
	}

	// the returned result is ParseResult.Item
	rspChan := e.Run(requests...)
	fmt.Println("completed  elapsed       actual rate")
	start := time.Now()
	for rsp := range rspChan {
		result := rsp.(int)

		e.Recorder.Done(result)

		elapsed := time.Since(start)
		actual := 1.0 / elapsed.Seconds()
		fmt.Printf("%5d      %8v  %v\n", result, elapsed, actual)
		start = time.Now()
	}

	// 偵測是否有未關閉的 goroutine
	// debug.SetTraceback("all")
	// panic(1)

	// Unordered Output:
	// Initial Tasks: 10
	// completed  elapsed       actual rate
	//     0      494.043269ms  0.18182080889746208
	//     1      499.860877ms  2.0005566468847693
	//     2      499.992179ms  2.000031284489352
	//     3      499.921848ms  2.0003126568695193
	//     4      500.343271ms  1.9986278580330903
	//     5      499.779441ms  2.0008826253419256
	//     6      499.720914ms  2.001116967459961
	//     7      500.203099ms  1.9991879338596419
	//     8      499.930578ms  2.0002777265606664
	//     9      509.194768ms  1.9638850648991744
	// Finish =============================================

}
