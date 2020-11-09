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

func (c *Customer) setName(inVal string) {
	c.name = inVal
}

func (c *Customer) getName() string {
	return c.name
}

func (c *Customer) setTrolley(inVal Trolley) {
	c.trolley = inVal
}

func (c *Customer) getTrolley() Trolley {
	return c.trolley
}

func (c *Customer) setAge(inVal int) {
	c.age = inVal
}

func (c *Customer) getAge() int {
	return c.age
}

func (c *Customer) setImpatient(inVal bool) {
	c.impatient = inVal
}

func (c *Customer) getImpatient() bool {
	return c.impatient
}

func (c *Customer) setGender(inVal string) {
	c.gender = inVal
}

func (c *Customer) getGender() string {
	return c.gender
}
