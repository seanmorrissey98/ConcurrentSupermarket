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

func NewSupermarket() Supermarket {
	s := Supermarket{make([]*Checkout, 8, 256), make(map[int]*Checkout), make(map[int]*Customer), make([]*Trolley, 200), 0, make(chan int), make(chan int)}
	s.GenerateTrolleys()
	s.GenerateCheckouts()
	go s.FinishedShoppingListener()
	go s.FinishedCheckoutListener()

	return s
}

func (s *Supermarket) GenerateCustomer() {
	s.numOfTotalCustomers++
	c := &Customer{id: s.numOfTotalCustomers}
	s.customers[c.id] = c
	fmt.Printf("Total num of customers so far: %d\n", s.numOfTotalCustomers)

	trolleySizes := []int{10, 100, 200}
	rand.Seed(time.Now().UnixNano())
	trolleySize := trolleySizes[rand.Intn(3)]

	for i, t := range s.trolleys {
		if t.capacity == trolleySize {
			c.trolley = t
			s.trolleys = append(s.trolleys[:i], s.trolleys[i+1:]...)
			break
		}
	}

	go c.Shop(s.finishedShopping)
}

func (s *Supermarket) SendToCheckout(id int) {
	c := s.customers[id]
	checkout := s.ChooseCheckout()
	checkout.AddPersonToLine(c)

	fmt.Printf("Customer #%d is going to checkout #%d with %d items\n", id, checkout.number, s.customers[id].GetNumProducts())
}

func (s *Supermarket) ChooseCheckout() *Checkout {
	min, pos := -1, -1
	for i := 0; i < len(s.checkoutOpen); i++ {
		if num := s.checkoutOpen[i].GetNumPeopleInLine(); num < min || min < 0 {
			min, pos = num, i
		}
	}

	return s.checkoutOpen[pos]
}

func (s *Supermarket) GenerateTrolleys() {
	trolleySizes := []int{10, 100, 200}
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 200; i++ {
		s.trolleys[i] = NewTrolley(trolleySizes[rand.Intn(3)])
	}
}

func (s *Supermarket) AssignTrolley() {
	trolleySizes := []int{10, 100, 200}
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 200; i++ {
		s.trolleys[i] = NewTrolley(trolleySizes[rand.Intn(3)])
	}
}

func (s *Supermarket) GenerateCheckouts() {
	rand.Seed(time.Now().UnixNano())
	tenOrLess := rand.Float64() < 0.5

	// Default create 8 Checkouts when Supermarket is created
	for i := 0; i < 8; i++ {
		s.checkoutOpen[i] = NewCheckout(i+1, tenOrLess, false, true, true, 10, false, make(chan *Customer, 8), 0, 0, 0, 0, true, s.finishedCheckout)
	}
}

func (s *Supermarket) FinishedShoppingListener() {
	for {
		// Check if customer is finished adding products to trolley using channel
		id := <-s.finishedShopping

		// Send customer to a checkout
		s.SendToCheckout(id)
	}
}

func (s *Supermarket) FinishedCheckoutListener() {
	for {
		// Check if customer is finished adding products to trolley using channel
		id := <-s.finishedCheckout
		trolley := s.customers[id].trolley
		trolley.EmptyTrolley()

		s.trolleys = append(s.trolleys, trolley)
		delete(s.customers, id)
	}
}
