package locktest

import (
	"fmt"
	"sync"
)

func LockPractice() {
	lock := sync.Mutex{}
	wait := sync.WaitGroup{}
	var counter = 0

	wait.Add(1)
	go func(n *int) {
		defer wait.Done()
		for i := 0; i < 10000; i++ {
			lock.Lock()
			*n++
			//fmt.Println("counter:", counter)
			lock.Unlock()
		}
	}(&counter)

	wait.Add(1)
	go func(n *int) {
		defer wait.Done()
		for i := 0; i < 10000; i++ {
			lock.Lock()
			*n++
			//fmt.Println("counter:", counter)
			lock.Unlock()
		}
	}(&counter)
	wait.Wait()
	fmt.Println(" main counter:", counter)
}
