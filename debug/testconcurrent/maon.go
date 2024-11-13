package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

func main() {
	b := make([]byte, 1000000)
	for i := range b {
		b[i] = 'f'
	}
	start := time.Now()
	writeConcurrent2([]string{"foo1", "foo2", "foo3"}, b)
	fmt.Printf("duration: %s\n", time.Since(start).String())
}

func writeBlocking(b []byte) {
	err := os.WriteFile("foo1", b, os.ModePerm)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("foo2", b, os.ModePerm)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("foo3", b, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func writeConcurrent(b []byte) {
	var wg sync.WaitGroup
	wg.Add(3)
	go func(wg *sync.WaitGroup, b []byte) {
		defer wg.Done()
		w("foo1", b)
	}(&wg, b)
	go func(wg *sync.WaitGroup, b []byte) {
		defer wg.Done()
		w("foo2", b)
	}(&wg, b)
	go func(wg *sync.WaitGroup, b []byte) {
		defer wg.Done()
		w("foo3", b)
	}(&wg, b)

	wg.Wait()
}

func writeConcurrent2(files []string, b []byte) {
	var wg sync.WaitGroup
	wg.Add(len(files))
	for i := range files {
		go func(filename string) {
			defer wg.Done()
			w(filename, b)
		}(files[i])
	}

	wg.Wait()
}

func w(f string, b []byte) {
	err := os.WriteFile(f, b, os.ModePerm)
	if err != nil {
		panic(err)
	}
}
