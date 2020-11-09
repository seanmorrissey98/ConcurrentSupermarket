package packageService

import (
	"math/rand"
)

type Trolley struct {
	trolleyCapacity int
	products        map[int]Product
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
