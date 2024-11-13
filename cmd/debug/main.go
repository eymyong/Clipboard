package main

import (
	"fmt"
	"sync"
)

// func main() {
// 	//  fmt.Println("1")
// 	fmt.Println("a")
// 	go fmt.Println("2")
// 	fmt.Println("b")
// }

func output(wg *sync.WaitGroup, msg string) {
	defer wg.Done()
	fmt.Println(msg)
}

func foo(wg sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Playing Dota2")
}

func f(wg *sync.WaitGroup, n int) {
	defer wg.Done()
	fmt.Println("n", n)
}

func main() {
	// Declare a waitgroup
	var wg sync.WaitGroup

	// Add(3) because of 3 functions of concurrent
	wg.Add(1)

	// Concurrent function list
	i := 1
	go func() {
		i = 100
	}()

	go func(wg *sync.WaitGroup, n int) {
		defer wg.Done()
		fmt.Println("n", n)
	}(&wg, i)

	i = 10
	i = 7

	// Waiting to finish
	wg.Wait()
}
