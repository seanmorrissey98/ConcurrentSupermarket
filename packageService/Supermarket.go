package packageService

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

var trolleyMutex *sync.Mutex
var customerMutex *sync.RWMutex
var checkoutMutex *sync.RWMutex

type Supermarket struct {
	openStatus       bool
	checkoutOpen     []*Checkout
	checkoutClosed   []*Checkout
	customers        map[int]*Customer
	trolleys         []*Trolley
	finishedShopping chan int
	finishedCheckout chan int
}

func (s *Supermarket) GetAllCheckouts() []*Checkout {
	return append(s.checkoutOpen, s.checkoutClosed...)
}

// Constructor for Supermarket
func NewSupermarket() *Supermarket {
	trolleyMutex = &sync.Mutex{}
	customerMutex = &sync.RWMutex{}
	checkoutMutex = &sync.RWMutex{}

	s := Supermarket{true, make([]*Checkout, 0, 256), make([]*Checkout, 0, 256), make(map[int]*Customer), make([]*Trolley, NUM_TROLLEYS), make(chan int), make(chan int)}
	s.GenerateTrolleys()
	s.GenerateCheckouts()

	go s.GenerateCustomer()
	go s.FinishedShoppingListener()
	go s.FinishedCheckoutListener()

	return &s
}

// Create a customer and adds them to to the customers map in supermarket
func (s *Supermarket) GenerateCustomer() {

	for {
		if !s.openStatus {
			break
		}

		time.Sleep(time.Millisecond * time.Duration(rand.Intn(int((1.0/float64(customerRate))*10000))))
		if len(s.trolleys) == 0 {
			continue
		}

		// Create a new customer with an id = the number they are created at in the supermarket
		c := &Customer{id: totalNumberOfCustomersToday}

		//fmt.Printf("Total num of customers so far: %d\n", s.numOfTotalCustomers)

		// Create 3 different trolley sizes modelling a basket, small trolley and large trolley
		trolleySize := TROLLEY_SIZES[rand.Intn(len(TROLLEY_SIZES))]

		// A customer picks a trolley based on the amount of products they need
		outOfTrolleys := false
		for i, t := range s.trolleys {
			if t.capacity == trolleySize {
				c.trolley = t

				trolleyMutex.Lock()
				s.trolleys[i] = s.trolleys[len(s.trolleys)-1]
				s.trolleys = s.trolleys[:len(s.trolleys)-1]
				trolleyMutex.Unlock()
				break
			} else if i == len(s.trolleys)-1 {
				//fmt.Println("No More Trolleys of Size: ", trolleySize)
				outOfTrolleys = true
			}
		}
		if outOfTrolleys {
			continue
		}

		// Add customer to stat print
		newCustomerChan <- 1

		// Add customer to the customers map in supermarket, key=customer.id, value=customer
		customerMutex.Lock()
		s.customers[c.id] = c
		customerMutex.Unlock()

		// Customer can now go add products to the trolley
		go c.Shop(s.finishedShopping)

		// Decides to open or close checkouts
		s.CalculateOpenCheckout()
	}
}

// Sends customer to a checkout
func (s *Supermarket) SendToCheckout(id int) {
	customerMutex.RLock()
	c := s.customers[id]
	customerMutex.RUnlock()

	// Choose the best checkout for a customer to go to
	for {
		checkoutMutex.RLock()
		checkout, _ := s.ChooseCheckout()
		if checkout.tenOrLess && c.GetNumProducts() > 10 {
			continue
		}
		checkout.AddPersonToLine(c)
		checkoutMutex.RUnlock()
		break
	}
}

// Gets the best open checkout for a customer to go to at the current time
func (s *Supermarket) ChooseCheckout() (*Checkout, int) {
	min, pos := -1, -1
	for i := 0; i < len(s.checkoutOpen); i++ {
		if num := s.checkoutOpen[i].GetNumPeopleInLine(); num < min || min < 0 {
			min, pos = num, i
		}
	}

	return s.checkoutOpen[pos], pos
}

// Generates 200 trolleys in the supermarket
func (s *Supermarket) GenerateTrolleys() {
	for i := 0; i < NUM_TROLLEYS; i++ {
		s.trolleys[i] = NewTrolley(TROLLEY_SIZES[rand.Intn(len(TROLLEY_SIZES))])
	}
}

// Generates 8 checkouts and opens them all by default
func (s *Supermarket) GenerateCheckouts() {
	// Default create 8 Checkouts when Supermarket is created
	for i := 0; i < NUM_CHECKOUTS; i++ {
		hasScanner := rand.Float64() < 0.5
		if i == 0 {
			s.checkoutOpen = append(s.checkoutOpen, NewCheckout(i+1, false, false, hasScanner, false, 0, false, make(chan *Customer, MAX_CUSTOMERS_PER_CHECKOUT), 0, 0, 0, 0, true, s.finishedCheckout))
		} else {
			s.checkoutClosed = append(s.checkoutClosed, NewCheckout(i+1, false, false, hasScanner, false, 0, false, make(chan *Customer, MAX_CUSTOMERS_PER_CHECKOUT), 0, 0, 0, 0, false, s.finishedCheckout))
		}
	}
}

// Waits for a customer to finish shopping using a channel, then sends the customer to a checkout
func (s *Supermarket) FinishedShoppingListener() {
	for {
		if !s.openStatus && numberOfCurrentCustomersShopping == 0 {
			break
		}

		// Check if customer is finished adding products to trolley using channel from the shop() method in Customer.go
		id := <-s.finishedShopping

		customerToCheckoutChan <- id

		// Send customer to a checkout
		s.SendToCheckout(id)
	}
}

// Waits for a customer to finish at a checkout using a channel, then removes the customer from the supermarket
func (s *Supermarket) FinishedCheckoutListener() {
	for {
		if !s.openStatus && totalNumberOfCustomersInStore == 0 {
			break
		}

		// Check if customer is finished at a checkout when all products are processed
		id := <-s.finishedCheckout
		//customerMutex.RLock()
		if s.customers[id] != nil {
			customerMutex.RLock()
			trolley := s.customers[id].trolley
			customerMutex.RUnlock()
			// Empty the customers trolley
			trolley.EmptyTrolley()
			// Adds the trolley the customer was using back into the trolleys slice in the supermarket
			trolleyMutex.Lock()
			s.trolleys = append(s.trolleys, trolley)
			trolleyMutex.Unlock()
		}

		// Remove customer from the supermarket
		customerMutex.Lock()
		delete(s.customers, id)
		customerMutex.Unlock()

		finishedAtCheckoutChan <- id

		s.CalculateOpenCheckout()
	}
}

// Calculates the threshold for opening / closing a checkout
func (s *Supermarket) CalculateOpenCheckout() {
	numOfCurrentCustomers := len(s.customers)
	numOfOpenCheckouts := len(s.checkoutOpen)
	calculationOfThreshold := int(math.Ceil(float64(numOfCurrentCustomers) / MAX_CUSTOMERS_PER_CHECKOUT))

	// Ensure at least 1 checkout stays open
	if numOfCurrentCustomers == 0 && s.openStatus {
		return
	}

	// Calculate threshold for opening a checkout
	if calculationOfThreshold > numOfOpenCheckouts {
		// If there are no more checkouts to open
		if len(s.checkoutClosed) == 0 {
			//fmt.Printf("All checkouts currently open. The current number of customers is: %d\n", numOfCurrentCustomers)
			return
		}

		// Open first checkout in closed checkout slice
		s.checkoutClosed[0].Open()
		s.checkoutOpen = append(s.checkoutOpen, s.checkoutClosed[0])
		s.checkoutClosed = s.checkoutClosed[1:]

		checkoutChangeStatusChan <- 1

		//fmt.Printf("1 new chekout opened. We now have %d open checkouts.\n", len(s.checkoutOpen))

		return
	}

	// Calculate threshold for closing a checkout
	if calculationOfThreshold < numOfOpenCheckouts {
		if len(s.checkoutOpen) == 1 && s.openStatus {
			//fmt.Printf("We only have one checkout open. Number of customer: %d\n", numOfCurrentCustomers)
			return
		}

		// Choose best checkout to close
		checkout, pos := s.ChooseCheckout()
		checkout.Close()
		s.checkoutClosed = append(s.checkoutClosed, checkout)
		s.checkoutOpen = append(s.checkoutOpen[0:pos], s.checkoutOpen[pos+1:]...)

		checkoutChangeStatusChan <- -1

		//fmt.Printf("1 checkout just closed. We now have %d open checkouts.\n", len(s.checkoutOpen))

		return
	}
}
