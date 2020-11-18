package packageService

import (
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

// Shop lets the customer get products and add them to their trolley until the reach capacity of trolley or break the random < 0.05
func (c *Customer) Shop(readyForCheckoutChan chan int) {
	//fmt.Printf("Customer #%d trolley size: %d\n", c.id, c.trolley.capacity)
	// Infinite loop of customer shopping
	for {
		// TODO: Add a 1 second sleep, will be replaces with a product wait time
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

	// Notify the channel in the supermarket FinishedShoppingListener() by sending the customer id to it
	readyForCheckoutChan <- c.id
	customerToCheckoutChan <- c.id
}

func (c *Customer) GetNumProducts() int {
	return len(c.trolley.products)
}
