# concurrencyengine - basic concurrency engine
[![GoDoc](https://godoc.org/github.com/z-Wind/concurrencyengine?status.png)](http://godoc.org/github.com/z-Wind/concurrencyengine)

## Table of Contents

* [Installation](#installation)
* [Usage](#usage)
* [Example](#example)
* [Reference](#reference)

## Installation

    $ go get github.com/z-Wind/concurrencyengine

## Usage

1. set ELog to record log
2. declare ctx to pass to New function
3. declare reqToKey function for the key of record map
4. create engine
5. declare parse function to process request
6. pass requests to engine
7. get the result from channel

## Example

```go
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
```

## Reference
- [[Go] Concurrency Patterns](https://zwindr.blogspot.com/2018/12/go-concurrency-patterns.html)
