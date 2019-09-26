// package main is just a simple experiment.
package main

import (
	"fmt"
	"time"
)

type Result interface {
	From() string
	Closest() []string
}

type Call interface {
	Do(address string) (ch chan Result, err error)
	Result(result Result) (done bool)
}

func routine(call Call) {
	address := "example.com"
	ch, err := call.Do(address)
	if err != nil {
		panic(err)
	}

	result := <-ch
	call.Result(result)
	fmt.Println(result.From())
	fmt.Println(result.Closest())
}

type ExampleResult struct {
	Something string
}

func (r *ExampleResult) From() string {
	return "me"
}

func (r *ExampleResult) Value() string {
	return "omg"
}

func (r *ExampleResult) Closest() []string {
	return []string{"hej", "svejs"}
}

type ExampleCall struct{}

func (q *ExampleCall) Do(address string) (chan Result, error) {
	ch := make(chan Result)
	go func(chan Result) {
		time.Sleep(100 * time.Millisecond)
		ch <- &ExampleResult{Something: "hello"}
	}(ch)
	return ch, nil
}

func (q *ExampleCall) Result(result Result) (done bool) {
	res, ok := result.(*ExampleResult)
	if !ok {
		return false
	}
	fmt.Println(res)
	return true
}

func main() {
	call := new(ExampleCall)
	routine(call)
}
