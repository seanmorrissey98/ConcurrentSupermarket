package packageService

import (
	"math/rand"
)

type Trolley struct {
	trolleyCapacity int
	products        map[int]Product
}

func (t *Trolley) SetTrolleyCapacity(inVal int) {
	t.trolleyCapacity = inVal
}

func (t *Trolley) GetTrolleyCapacity() int {
	return t.trolleyCapacity
}

func (t *Trolley) SetProducts(inVal map[int]Product) {
	t.products = inVal
}

func (t *Trolley) InitalizeProducts() {
	t.products = make(map[int]Product)
}

func (t *Trolley) GetProducts() map[int]Product {
	return t.products
}

func (t *Trolley) GetProduct(inVal int) Product {
	return t.products[inVal]
}

func (t *Trolley) AddProductToTrolley(inVal Product, inVal2 int) {
	t.products[inVal2] = inVal
}

func (t *Trolley) FillTrolley(timeMult int) {
	for i := 0; i < t.trolleyCapacity; i++ {
		t.products[i] = Product{
			time: rand.Intn(timeMult),
		}
	}
}
