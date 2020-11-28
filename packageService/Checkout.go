package packageService

import (
	"sync/atomic"
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
	processedProducts  int64
	processedCustomers int64
	speed              float64
	isOpen             bool
	finishedProcessing chan int
}

// Checkout Constructor
func NewCheckout(number int, tenOrLess bool, isSelfCheckout bool, hasScanner bool, inUse bool, lineLength int, isLineFull bool, peopleInLine chan *Customer, averageWaitTime float32, processedProducts int64, processedCustomers int64, speed float64, isOpen bool, finishedProcessing chan int) *Checkout {
	c := Checkout{number, tenOrLess, isSelfCheckout, hasScanner, inUse, lineLength, isLineFull, peopleInLine, averageWaitTime, processedProducts, processedCustomers, speed, isOpen, finishedProcessing}

	if c.hasScanner {
		c.speed = 0.5
	} else {
		c.speed = 1.0
	}

	//rand.Seed(time.Now().UnixNano())
	//c.tenOrLess = rand.Float64() < 0.25

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
	customer.waitTime = time.Now().UnixNano()
	c.peopleInLine <- customer
	c.lineLength++
}

// Processes all products in a customers trolley
func (c *Checkout) ProcessCheckout() {
	for {
		if !c.isOpen && c.lineLength == 0 {
			break
		}

		// Get the first customer in line
		customer := <-c.peopleInLine
		// Check if customer is nil, break open of for loop and set checkout open to false
		if customer == nil {
			c.isOpen = false
			break
		}
		c.lineLength--

		// Start customer wait timer
		customer.waitTime = time.Now().UnixNano() - customer.waitTime

		trolley := customer.trolley
		products := trolley.products

		age := customer.GetAge()
		var ageMultiplier float64
		ageMultiplier = 1
		if age > 65 {
			ageMultiplier = 1.5
		}

		// Start customer process timer
		customer.processTime = time.Now().UnixNano()

		// Get all products in trolley and calculate the time to wait
		for _, p := range products {
			time.Sleep(time.Millisecond * time.Duration(p.GetTime()*500*c.speed*ageMultiplier))
			atomic.AddInt64(&c.processedProducts, 1)
		}

		// Stop customer process timer
		customer.processTime = time.Now().UnixNano() - customer.processTime

		// Send customer is to finished process channel
		c.finishedProcessing <- customer.id

		// Increments the processed customer after customer is finished ar checkout
		atomic.AddInt64(&c.processedCustomers, 1)
	}
}

func (c *Checkout) Open() {
	c.isOpen = true
	go c.ProcessCheckout()
}

// Passes a nil customer to the peopleInLine channel
func (c *Checkout) Close() {
	c.peopleInLine <- nil
}

func (c *Checkout) GetTotalCustomersProcessed() int64 {
	return c.processedCustomers
}

func (c *Checkout) GetCheckoutNumber() int {
	return c.number
}

func (c *Checkout) GetTotalProductsProcessed() int64 {
	return c.processedProducts
}

func (c *Checkout) GetId() int {
	return c.number
}
