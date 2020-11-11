package packageService

type Trolley struct {
	capacity int
	products []*Product
}

// Trolley Constructor
func NewTrolley(capacity int) *Trolley {
	t := Trolley{capacity, make([]*Product, 1, capacity)}
	return &t
}

// Adds a product to a trolley
func (t *Trolley) AddProductToTrolley(product *Product) {
	t.products = append(t.products, product)
}

// Checks if trolley has reached capacity
func (t *Trolley) IsFull() bool {
	return t.capacity == len(t.products)
}

// Empties trolley by declaring the current slice as a new slice
func (t *Trolley) EmptyTrolley() {
	t.products = make([]*Product, 1, t.capacity)
}
