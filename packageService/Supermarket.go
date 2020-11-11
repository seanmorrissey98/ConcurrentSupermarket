package packageService

import (
	"fmt"
	"math/rand"
	"time"
)

type Supermarket struct {
	checkoutOpen        []*Checkout
	checkoutClosed      map[int]*Checkout
	customers           map[int]*Customer
	trolleys            []*Trolley
	numOfTotalCustomers int
	finishedShopping    chan int
	finishedCheckout    chan int
}

// Constructor for Supermarket
func NewSupermarket() Supermarket {
	s := Supermarket{make([]*Checkout, 8, 256), make(map[int]*Checkout), make(map[int]*Customer), make([]*Trolley, 200), 0, make(chan int), make(chan int)}
	s.GenerateTrolleys()
	s.GenerateCheckouts()
	go s.FinishedShoppingListener()
	go s.FinishedCheckoutListener()

	return s
}

// Create a customer and adds them to to the customers map in supermarket
func (s *Supermarket) GenerateCustomer() {
	// Increment the number of customers in the supermarket
	s.numOfTotalCustomers++
	// Create a new customer with an id = the number they are created at in the supermarket
	c := &Customer{id: s.numOfTotalCustomers}
	// Add customer to the customers map in supermarket, key=customer.id, value=customer
	s.customers[c.id] = c

	fmt.Printf("Total num of customers so far: %d\n", s.numOfTotalCustomers)

	// Create 3 different trolley sizes modelling a basket, small trolley and large trolley
	trolleySizes := []int{10, 100, 200}
	rand.Seed(time.Now().UnixNano())
	trolleySize := trolleySizes[rand.Intn(3)]

	// A customer picks a trolley based on the amount of products they need
	for i, t := range s.trolleys {
		if t.capacity == trolleySize {
			c.trolley = t
			s.trolleys = append(s.trolleys[:i], s.trolleys[i+1:]...)
			break
		}
	}

	// Customer can now go add products to the trolley
	go c.Shop(s.finishedShopping)
}

// Sends customer to a checkout
func (s *Supermarket) SendToCheckout(id int) {
	c := s.customers[id]

	// Choose the best checkout for a customer to go to
	checkout := s.ChooseCheckout()
	checkout.AddPersonToLine(c)

	fmt.Printf("Customer #%d is going to checkout #%d with %d items\n", id, checkout.number, s.customers[id].GetNumProducts())
}

// Gets the best open checkout for a customer to go to at the current time
func (s *Supermarket) ChooseCheckout() *Checkout {
	min, pos := -1, -1
	for i := 0; i < len(s.checkoutOpen); i++ {
		if num := s.checkoutOpen[i].GetNumPeopleInLine(); num < min || min < 0 {
			min, pos = num, i
		}
	}

	return s.checkoutOpen[pos]
}

// Generates 200 trolleys in the supermarket
func (s *Supermarket) GenerateTrolleys() {
	trolleySizes := []int{10, 100, 200}
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 200; i++ {
		s.trolleys[i] = NewTrolley(trolleySizes[rand.Intn(3)])
	}
}

// Generates 8 checkouts and opens them all by default
func (s *Supermarket) GenerateCheckouts() {
	rand.Seed(time.Now().UnixNano())
	tenOrLess := rand.Float64() < 0.5

	// Default create 8 Checkouts when Supermarket is created
	for i := 0; i < 8; i++ {
		s.checkoutOpen[i] = NewCheckout(i+1, tenOrLess, false, true, true, 10, false, make(chan *Customer, 8), 0, 0, 0, 0, true, s.finishedCheckout)
	}
}

// Waits for a customer to finish shopping using a channel, then sends the customer to a checkout
func (s *Supermarket) FinishedShoppingListener() {
	for {
		// Check if customer is finished adding products to trolley using channel from the shop() method in Customer.go
		id := <-s.finishedShopping
		// Send customer to a checkout
		s.SendToCheckout(id)
	}
}

// Waits for a customer to finnish at a checkout using a channel, then removes the customer from the supermarket
func (s *Supermarket) FinishedCheckoutListener() {
	for {
		// Check if customer is finished at a checkout when all products are processed
		id := <-s.finishedCheckout
		trolley := s.customers[id].trolley
		// Empty the customers trolley
		trolley.EmptyTrolley()

		// Adds the trolley the customer was using back into the trolleys slice in the supermarket
		s.trolleys = append(s.trolleys, trolley)
		// Remove customer from the supermarket
		delete(s.customers, id)
	}
}
