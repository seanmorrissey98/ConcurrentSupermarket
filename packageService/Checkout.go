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

	if c.hasScanner {
		c.speed = 0.5
	} else {
		c.speed = 1.0
	}

	// Starts a goroutine for processing all products in a trolley
	if isOpen {
		go c.ProcessCheckout()
	}

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
}

// Processes all products in a customers trolley
func (c *Checkout) ProcessCheckout() {
	for {
		if !c.isOpen && c.lineLength == 0 {
			break
		}

		var processDuration time.Duration
		processDuration = 0

		// Get the first customer in line
		customer := <-c.peopleInLine
		if customer == nil {
			c.isOpen = false
			break
		} else {
			c.lineLength--
		}
		trolley := customer.trolley
		products := trolley.products

		for _, p := range products {
			processDuration += time.Millisecond * time.Duration(int(p.GetTime()*500*c.speed))
			time.Sleep(time.Millisecond * time.Duration(int(p.GetTime()*500*c.speed)))
		}

		customer.processTime = processDuration
		c.finishedProcessing <- customer.id
		c.processedCustomers++
	}
}

/*func (c *Checkout) RemovePersonFromLine(customer *Customer) {
	delete(c.peopleInLine, c.lineLength)
	c.lineLength--
}*/

func (c *Checkout) Open() {
	c.isOpen = true
	go c.ProcessCheckout()
}

func (c *Checkout) Close() {
	c.peopleInLine <- nil
}

func (c *Checkout) GetTotalCustomersProcessed() int {
	return c.processedCustomers
}

func (c *Checkout) GetCheckoutNumber() int {
	return c.number
}
