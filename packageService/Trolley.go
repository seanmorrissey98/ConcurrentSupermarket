package packageService

type Trolley struct {
	capacity int
	products []*Product
}

func NewTrolley(capacity int) *Trolley {
	t := Trolley{capacity, make([]*Product, 1, capacity)}
	return &t
}

func (t *Trolley) SetTrolleyCapacity(inVal int) {
	t.capacity = inVal
}

func (t *Trolley) GetTrolleyCapacity() int {
	return t.capacity
}

func (t *Trolley) SetProducts(inVal []*Product) {
	t.products = inVal
}

func (t *Trolley) GetProducts() []*Product {
	return t.products
}

func (t *Trolley) GetProduct(inVal int) *Product {
	return t.products[inVal]
}

func (t *Trolley) AddProductToTrolley(product *Product) {
	t.products = append(t.products, product)
}

func (t *Trolley) IsFull() bool {
	return t.capacity == len(t.products)
}

/*func (t *Trolley) FillTrolley(timeMult int) {
	for i := 0; i < t.capacity; i++ {
		t.products[i] = &Product{
			time: rand.Intn(timeMult)
		}
	}
}*/
