package packageService

type Trolley struct {
	capacity int
	products []*Product
}

func NewTrolley(capacity int) *Trolley {
	t := Trolley{capacity, make([]*Product, 1, capacity)}
	return &t
}

func (t *Trolley) AddProductToTrolley(product *Product) {
	t.products = append(t.products, product)
}

func (t *Trolley) IsFull() bool {
	return t.capacity == len(t.products)
}

func (t *Trolley) EmptyTrolley() {
	t.products = make([]*Product, 1, t.capacity)
}
