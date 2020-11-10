package packageService

import (
	"math/rand"
	"sync"
	"time"
)

type Customer struct {
	id        int
	name      string
	trolley   Trolley
	age       int
	impatient bool
	gender    string
	mutex     sync.Mutex
}

func (c *Customer) Shop(finishedShopping chan int) {
	for {
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
		p := NewProduct()
		c.trolley.AddProductToTrolley(p)

		if c.trolley.IsFull() {
			break
		}

		if rand.Float64() < 0.15 {
			break
		}
	}

	finishedShopping <- c.id
}

func (c *Customer) SetId(inVal int) {
	c.id = inVal
}

func (c *Customer) GetId() int {
	return c.id
}

func (c *Customer) SetName(inVal string) {
	c.name = inVal
}

func (c *Customer) GetName() string {
	return c.name
}

func (c *Customer) SetTrolley(inVal Trolley) {
	c.trolley = inVal
}

func (c *Customer) GetTrolley() Trolley {
	return c.trolley
}

func (c *Customer) GetNumProducts() int {
	return len(c.trolley.GetProducts())
}

func (c *Customer) SetAge(inVal int) {
	c.age = inVal
}

func (c *Customer) GetAge() int {
	return c.age
}

func (c *Customer) SetImpatient(inVal bool) {
	c.impatient = inVal
}

func (c *Customer) GetImpatient() bool {
	return c.impatient
}

func (c *Customer) SetGender(inVal string) {
	c.gender = inVal
}

func (c *Customer) GetGender() string {
	return c.gender
}
