package packageService

import (
	"math/rand"
	"sync"
	"time"
)

type Customer struct {
	id          int
	name        string
	trolley     *Trolley
	age         int
	impatient   bool
	gender      string
	mutex       sync.Mutex
	processTime int64
	waitTime    int64
}

// Shop lets the customer get products and add them to their trolley until the reach capacity of trolley or break the random < 0.05
func (c *Customer) Shop(readyForCheckoutChan chan int) {
	// Infinite loop of customer shopping
	for {
		if c.GetNumProducts() == int(productsRate) {
			break
		}
		p := NewProduct()
		time.Sleep(time.Millisecond * time.Duration(int(p.GetTime()*200)))
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
}

func (c *Customer) GetNumProducts() int {
	return len(c.trolley.products)
}
