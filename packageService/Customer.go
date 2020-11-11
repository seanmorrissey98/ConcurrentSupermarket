package packageService

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Customer struct {
	id        int
	name      string
	trolley   *Trolley
	age       int
	impatient bool
	gender    string
	mutex     sync.Mutex
}

func (c *Customer) Shop(finishedShopping chan int) {
	fmt.Printf("Customer #%d trolley size: %d\n", c.id, c.trolley.capacity)
	for {
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
		p := NewProduct()
		c.trolley.AddProductToTrolley(p)

		if c.trolley.IsFull() {
			break
		}

		if rand.Float64() < 0.05 {
			break
		}
	}

	finishedShopping <- c.id
}

func (c *Customer) GetNumProducts() int {
	return len(c.trolley.products)
}
