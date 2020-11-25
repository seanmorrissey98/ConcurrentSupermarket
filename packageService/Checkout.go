package packageService

import (
	"math/rand"
	"sync"
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
	processedCustomers int
	speed              float64
	isOpen             bool
	finishedProcessing chan int
	mutexCheckout      sync.Mutex
}

// Checkout Constructor
func NewCheckout(number int, tenOrLess bool, isSelfCheckout bool, hasScanner bool, inUse bool, lineLength int, isLineFull bool, peopleInLine chan *Customer, averageWaitTime float32, processedProducts int64, processedCustomers int, speed float64, isOpen bool, finishedProcessing chan int) *Checkout {
	var mut = sync.Mutex{}
	c := Checkout{number, tenOrLess, isSelfCheckout, hasScanner, inUse, lineLength, isLineFull, peopleInLine, averageWaitTime, processedProducts, processedCustomers, speed, isOpen, finishedProcessing,mut}

	if c.hasScanner {
		c.speed = 0.5
	} else {
		c.speed = 1.0
	}

	rand.Seed(time.Now().UnixNano())
	c.tenOrLess = rand.Float64() < 0.25

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
	c.mutexCheckout.Lock()
	c.peopleInLine <- customer
	c.lineLength++
	if c.GetNumPeopleInLine() > 1 {
		sizeLine:= c.GetNumPeopleInLine()
		for x:=0; x <sizeLine; x++{
			cust := <- c.peopleInLine
			c.lineLength --
			if customer.id != cust.id{
			customer.peopleInFront=append(customer.peopleInFront,cust)
			}
			c.peopleInLine <- cust
			c.lineLength ++
		}
	}
	c.mutexCheckout.Unlock()
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

		var w sync.WaitGroup

		for _, p := range products {
			processDuration += time.Millisecond * time.Duration(int(p.GetTime()*500*c.speed))
			time.Sleep(time.Millisecond * time.Duration(int(p.GetTime()*500*c.speed)))

			w.Add(1)
			go c.increment(&w)

		}
		w.Wait()
		customer.processTime = processDuration
		var sum time.Duration
		sum=0
		for _,x := range customer.peopleInFront{
			sum=sum+x.processTime
		}
		customer.waitTime=sum
		c.finishedProcessing <- customer.id
		c.processedCustomers++
	}
}

/*func (c *Checkout) RemovePersonFromLine(customer *Customer) {
	delete(c.peopleInLine, c.lineLength)
	c.lineLength--
}*/

func (c *Checkout) increment(wg *sync.WaitGroup) {
	atomic.AddInt64(&c.processedProducts, 1)
	wg.Done()
}

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

func (c *Checkout) GetTotalProductsProcessed() int64 {
	return c.processedProducts
}
