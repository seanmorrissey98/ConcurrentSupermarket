package packageService

import (
	"sync"
)

type Customer struct {
	name      string
	trolley   Trolley
	age       int
	impatient bool
	gender    string
	mutex     sync.Mutex
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
