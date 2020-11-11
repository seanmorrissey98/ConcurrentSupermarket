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

// Checkout Constructor
func NewCheckout(number int, tenOrLess bool, isSelfCheckout bool, hasScanner bool, inUse bool, lineLength int, isLineFull bool, peopleInLine chan *Customer, averageWaitTime float32, processedProducts int, processedCustomers int, speed int, isOpen bool, finishedProcessing chan int) *Checkout {
	c := Checkout{number, tenOrLess, isSelfCheckout, hasScanner, inUse, lineLength, isLineFull, peopleInLine, averageWaitTime, processedProducts, processedCustomers, speed, isOpen, finishedProcessing}

	// Starts a goroutine for processing all products in a trolley
	go c.ProcessCheckout()

	return &c
}

// Gets the number of customers in a checkout line
func (c *Checkout) GetNumPeopleInLine() int {
	return len(c.peopleInLine)
}

// Adds a customer a specific checkout line
func (c *Checkout) AddPersonToLine(customer *Customer) {
	// Use channel instead a list of customers to easily pop and send the customer
	c.peopleInLine <- customer
	c.lineLength++

	// TODO: If theres only one customer in the checkout line, start checkout
	//if c.lineLength == 1 {
	//	go c.ProcessCheckout()
	//}
}

// Processes all products in a customers trolley
func (c *Checkout) ProcessCheckout() {
	for {
		// Get the first customer in line
		customer := <-c.peopleInLine

		trolley := customer.trolley
		products := trolley.products

		start := time.Now().UnixNano()

		for range products {
			// TODO: product.time
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
		}

		// Get the total time taken to process all products
		totalTime := time.Now().UnixNano() - start
		fmt.Printf("Customer #%d, Time: %v seconds\n", customer.id, totalTime/int64(time.Second))
	}
}

/*func (c *Checkout) RemovePersonFromLine(customer *Customer) {
	delete(c.peopleInLine, c.lineLength)
	c.lineLength--
}*/

func Open(c *Checkout) {
	c.isOpen = true
}

func Close(c *Checkout) {
	c.isOpen = false
}
