package packageService

import (
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
	speed              float64
	isOpen             bool
	finishedProcessing chan int
}

// Checkout Constructor
func NewCheckout(number int, tenOrLess bool, isSelfCheckout bool, hasScanner bool, inUse bool, lineLength int, isLineFull bool, peopleInLine chan *Customer, averageWaitTime float32, processedProducts int, processedCustomers int, speed float64, isOpen bool, finishedProcessing chan int) *Checkout {
	c := Checkout{number, tenOrLess, isSelfCheckout, hasScanner, inUse, lineLength, isLineFull, peopleInLine, averageWaitTime, processedProducts, processedCustomers, speed, isOpen, finishedProcessing}
	// Starts a goroutine for processing all products in a trolley
	if c.hasScanner {
		c.speed = 0.5
	} else {
		c.speed = 1.0
	}
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

		//start := time.Now().UnixNano()

		for _, p := range products {
			time.Sleep(time.Millisecond * time.Duration(int(p.GetTime()*1000*c.speed)))
		}
		// Get the total time taken to process all products
		//totalTime := time.Now().UnixNano() - start
		//fmt.Printf("Customer #%d, Time: %v seconds\n", customer.id, totalTime/int64(time.Second))
		c.finishedProcessing <- customer.id
		finishedAtCheckoutChan <- customer.id
	}
}

/*func (c *Checkout) RemovePersonFromLine(customer *Customer) {
	delete(c.peopleInLine, c.lineLength)
	c.lineLength--
}*/

func (c *Checkout) Open() {
	c.isOpen = true
}

func (c *Checkout) Close() {
	c.isOpen = false
}
