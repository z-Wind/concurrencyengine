package concurrencyengine

import (
	"context"
	"fmt"
	"os"
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

		if n%2 == 0 {
			parseResult.Done = true
			parseResult.ExtraRequests = append(parseResult.ExtraRequests, Request{n + 11, req.ParseFunc})

			return parseResult, nil
		}

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
	for rsp := range rspChan {
		result := rsp.(int)

		e.Recorder.Done(result)

		fmt.Printf("odd %d\n", result)
	}

	// 偵測是否有未關閉的 goroutine
	// debug.SetTraceback("all")
	// panic(1)

	// Unordered Output:
	// Initial Tasks: 10
	// odd 1
	// odd 11
	// odd 3
	// odd 13
	// odd 5
	// odd 15
	// odd 7
	// odd 17
	// odd 9
	// odd 19
	// Finish =============================================
}
