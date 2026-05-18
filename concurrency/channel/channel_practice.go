package channel

import (
	"fmt"
	"sync"
	"time"
)

func ChannelPractice() {
	sending := make(chan int)
	reception := make(chan int)
	wg := sync.WaitGroup{}

	wg.Add(2)
	go func() {
		defer wg.Done()
		defer close(reception)
		for i := 0; i < 10; i++ {
			sending <- i
		}
	}()

	go func() {
		defer wg.Done()
		defer close(sending)
		for seq := range sending {
			reception <- seq
		}
	}()

	//select {
	//case a := <-reception:
	//	println("reception in select", a)
	//}
	for rec := range reception {
		println("reception ", rec)
	}
	wg.Wait()
}

func doWork(done chan bool) {
	select {
	case <-done:
		println("doWork is done")
		return
	default:
		println("do work is running")
	}
}

func ChannelPracticeDoneSignal() {
	done := make(chan bool)
	go doWork(done)

	time.Sleep(time.Second * 10)
	close(done)
	fmt.Println("ChannelPracticeDoneSignal is over")
}
