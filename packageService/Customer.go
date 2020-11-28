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
	shopTime    int64
}

// Shop lets the customer get products and add them to their trolley until the reach capacity of trolley or break the random < 0.05
func (c *Customer) Shop(readyForCheckoutChan chan int) {

	var speedMultiplier float64
	speedMultiplier = 1

	// Infinite loop of customer shopping
	for {
		if c.GetNumProducts() == int(productsRate) {
			break
		}

		if c.age > 65 {
			speedMultiplier = 1.5
		}

		p := NewProduct()
		time.Sleep(time.Millisecond * time.Duration(p.GetTime()*200*speedMultiplier))
		c.trolley.AddProductToTrolley(p)
		c.shopTime += int64(p.GetTime() * 200)
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

func (c *Customer) GetAge() int {
	return c.age
}
