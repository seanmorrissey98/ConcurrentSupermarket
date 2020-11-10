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
}

func NewSupermarket() Supermarket {
	s := Supermarket{make([]*Checkout, 8, 256), make(map[int]*Checkout), make(map[int]*Customer), make([]*Trolley, 200), 0, make(chan int)}
	s.GenerateTrolleys()
	s.GenerateCheckouts()
	go s.FinishedShoppingListener()

	return s
}

func (s *Supermarket) GenerateCustomer() {
	s.numOfTotalCustomers++
	c := &Customer{id: s.numOfTotalCustomers}
	s.customers[c.id] = c
	fmt.Printf("Total num of customers so far: %d\n", s.numOfTotalCustomers)

	go c.Shop(s.finishedShopping)
}

func (s *Supermarket) SendToCheckout(id int) {
	c := s.customers[id]
	checkout := s.ChooseCheckout()
	checkout.AddPersonToLine(c)
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

func (s *Supermarket) GenerateCheckouts() {
	rand.Seed(time.Now().UnixNano())
	tenOrLess := rand.Float64() < 0.5

	for i := 0; i < 8; i++ {
		s.checkoutOpen[i] = NewCheckout(i+1, tenOrLess, false, true, true, 10, false, make([]*Customer, 1, 10), 0, 0, 0, 0, true)
	}
}

func (s *Supermarket) FinishedShoppingListener() {
	for {
		id := <-s.finishedShopping
		fmt.Printf("Customer #%d finished shopping with %d items\n", id, s.customers[id].GetNumProducts())
		s.SendToCheckout(id)
	}
}
