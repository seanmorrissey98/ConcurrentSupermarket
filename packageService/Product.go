package packageService

type Product struct {
	time int
}

func (p *Product) setTime(inVal int) {
	p.time = inVal
}

func (p *Product) getTime() int {
	return p.time
}
