package packageService

import (
	"fmt"
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
	//c.GetNumPeopleInLine() gets the number of people in front of you in the queue
	fmt.Printf("Before Checkout: %d\nLine Length: %d\nNumPeopleInLine: %d\n", c.number, c.lineLength, c.GetNumPeopleInLine())
	//It is always one less than the Line length as it does not include the person who is currently being processed
	//While line length does inclue the person who is currently being processed
	//This is because the person who is currently being processed gets popped off the channel
	//This happens in ProcessCheckout(), line "customer := <-c.peopleInLine"
	c.peopleInLine <- customer
	c.lineLength++
	fmt.Printf("After Checkout: %d\nLine Length: %d\nNumPeopleInLine: %d\n\n", c.number, c.lineLength, c.GetNumPeopleInLine())

	// TODO: If theres only one customer in the checkout line, start checkout
	//if c.lineLength == 1 {
	//	go c.ProcessCheckout()
	//}
}

// Processes all products in a customers trolley
func (c *Checkout) ProcessCheckout() {
	for {
		if !c.isOpen { //&& c.lineLength == 0 {
			break
		}

		var processDuration time.Duration
		processDuration = 0

		// Get the first customer in line
		c.lineLength--
		customer := <-c.peopleInLine
		if customer == nil {
			break
		}
		trolley := customer.trolley
		products := trolley.products

		//start := time.Now().UnixNano()

		for _, p := range products {
			processDuration += time.Millisecond * time.Duration(int(p.GetTime()*500*c.speed))
			time.Sleep(time.Millisecond * time.Duration(int(p.GetTime()*500*c.speed)))
		}
		// Get the total time taken to process all products
		//totalTime := time.Now().UnixNano() - start
		//fmt.Printf("Customer #%d, Time: %v seconds\n", customer.id, totalTime/int64(time.Second))
		customer.processTime = processDuration
		fmt.Println(customer.processTime)
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
	if c.lineLength == 0 {
		c.peopleInLine <- nil
	}

	c.isOpen = false
}

func (c *Checkout) GetTotalCustomersProcessed() int {
	return c.processedCustomers
}

func (c *Checkout) GetCheckoutNumber() int {
	return c.number
}
