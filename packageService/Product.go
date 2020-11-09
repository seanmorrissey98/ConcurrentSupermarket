package packageService

type Product struct {
	time int
}

func (p *Product) SetTime(inVal int) {
	p.time = inVal
}

func (p *Product) GetTime() int {
	return p.time
}
