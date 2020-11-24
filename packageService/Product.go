package packageService

import (
	"math/rand"
)

type Product struct {
	time float64
}

// Product Constructor
func NewProduct() *Product {
	p := Product{rand.Float64() * processSpeed}
	return &p
}

func (p *Product) GetTime() float64 {
	return p.time
}
