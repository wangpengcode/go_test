package channel

import "sync"

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
