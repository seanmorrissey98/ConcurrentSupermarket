package packageService

import "time"

type Product struct {
	time int64
}

func NewProduct() *Product {
	p := Product{time.Now().UnixNano()}
	return &p
}

func (p *Product) GetTime() int64 {
	return p.time
}
