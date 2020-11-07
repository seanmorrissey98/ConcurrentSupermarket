package main

import (
	"fmt"
	"math/rand"
	"sync"
)

type Product struct {
	time int
}

type Trolley struct {
	trolleyCapacity int
	products        map[int]Product
}

type Customer struct {
	name      string
	trolley   Trolley
	age       int
	impatient bool
	gender    string
	mutex     sync.Mutex
}

func (p *Product) setTime(inVal int) {
	p.time = inVal
}

func (p *Product) getTime() int {
	return p.time
}

func (t *Trolley) setTrolleyCapacity(inVal int) {
	t.trolleyCapacity = inVal
}

func (t *Trolley) getTrolleyCapacity() int {
	return t.trolleyCapacity
}

func (t *Trolley) setProducts(inVal map[int]Product) {
	t.products = inVal
}

func (t *Trolley) initalizeProducts() {
	t.products = make(map[int]Product)
}

func (t *Trolley) getProducts() map[int]Product {
	return t.products
}

func (t *Trolley) getProduct(inVal int) Product {
	return t.products[inVal]
}

func (t *Trolley) addProductToTrolley(inVal Product, inVal2 int) {
	t.products[inVal2] = inVal
}

func (t *Trolley) fillTrolley(timeMult int) {
	for i := 0; i < t.trolleyCapacity; i++ {
		t.products[i] = Product{
			time: rand.Intn(timeMult),
		}
	}
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

func main() {
	trolley := new(Trolley)
	trolley.setTrolleyCapacity(1)
	apple := new(Product)
	apple.setTime(2)
	trolley.initalizeProducts()
	trolley.addProductToTrolley(*apple, 0)
	for i := 0; i < trolley.getTrolleyCapacity(); i++ {
		productInTrolley := trolley.getProduct(i)
		fmt.Println(productInTrolley.getTime())
	}
}
