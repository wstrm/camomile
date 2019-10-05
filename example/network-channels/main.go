package main

import (
	"fmt"
	"time"
)

func do() chan string {
	return make(chan string)
}

func main() {
	ch := do()

	go func() {
		fmt.Println(<-ch)
	}()
	go func() {
		fmt.Println(<-ch)
	}()

	ch <- "Hej!"
	time.Sleep(time.Second)
}
