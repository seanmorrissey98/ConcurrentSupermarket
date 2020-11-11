package packageService

import (
	"fmt"
	"math/rand"
	"time"
)

type Checkout struct {
	number             int
	tenOrLess          bool
	isSelfCheckout     bool
	hasScanner         bool
	inUse              bool
	lineLength         int
	isLineFull         bool
	peopleInLine       chan *Customer
	averageWaitTime    float32
	processedProducts  int
	processedCustomers int
	speed              int
	isOpen             bool
	finishedProcessing chan int
}

func NewCheckout(number int, tenOrLess bool, isSelfCheckout bool, hasScanner bool, inUse bool, lineLength int, isLineFull bool, peopleInLine chan *Customer, averageWaitTime float32, processedProducts int, processedCustomers int, speed int, isOpen bool, finishedProcessing chan int) *Checkout {
	c := Checkout{number, tenOrLess, isSelfCheckout, hasScanner, inUse, lineLength, isLineFull, peopleInLine, averageWaitTime, processedProducts, processedCustomers, speed, isOpen, finishedProcessing}

	go c.ProcessCheckout()

	return &c
}

func (c *Checkout) GetNumPeopleInLine() int {
	return len(c.peopleInLine)
}

func (c *Checkout) AddPersonToLine(customer *Customer) {
	c.peopleInLine <- customer
	c.lineLength++

	// If theres only one customer in the checkout line, start checkout
	//if c.lineLength == 1 {
	//	go c.ProcessCheckout()
	//}
}

func (c *Checkout) ProcessCheckout() {
	for {
		customer := <-c.peopleInLine

		trolley := customer.trolley
		products := trolley.products

		start := time.Now().UnixNano()

		for range products {
			//product.time
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
		}

		totalTime := time.Now().UnixNano() - start
		fmt.Printf("Customer #%d, Time: %v\n", customer.id, totalTime/int64(time.Second))
	}
}

/*func (c *Checkout) RemovePersonToLine(customer *Customer) {
	delete(c.peopleInLine, c.lineLength)
	c.lineLength--
}*/

func Open(c *Checkout) {
	c.isOpen = true
}

func Close(c *Checkout) {
	c.isOpen = false
}
