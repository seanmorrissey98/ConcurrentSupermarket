package packageService

import (
	"fmt"
)

type Checkout struct {
	number int
	tenOrLess bool
	isSelfCheckout bool
	hasScanner bool
	inUse bool
	lineLength int
	isLineFull bool
	peopleInLine map[int]*Customer
	averageWaitTime float32
	processedProducts int
	processedCustomers int
	speed int
	isOpen bool

}

func (c *Checkout) SetNumber(inVal int) {
	c.number = inVal
}

func (c *Checkout) GetNumber() int {
	return c.number
}

func (c *Checkout) InitalizeCheckout() {
	c.peopleInLine = make(map[int]*Customer)
}

func (c *Checkout) SetTenOLess(inVal bool) {
	c.tenOrLess = inVal
}

func (c *Checkout) GetTenOLess() bool {
	return c.tenOrLess
}
func (c *Checkout) SetIsSelfCheckout(inVal bool) {
	c.isSelfCheckout = inVal
}

func (c *Checkout) GetIsSelfCheckout() bool {
	return c.isSelfCheckout
}

func (c *Checkout) SetHasScanner(inVal bool) {
	c.hasScanner = inVal
}

func (c *Checkout) GetHasScanner() bool{
	return c.hasScanner
}

func (c *Checkout) SetInUse(inVal bool) {
	c.inUse = inVal
}

func (c *Checkout) GetInUse() bool {
	return c.inUse
}

func (c *Checkout) SetIsLineFull(inVal bool) {
	c.isLineFull = inVal
}

func (c *Checkout) GetIsLineFull() bool {
	return c.isLineFull
}

func (c *Checkout) SetPeopleInLine(inVal map[int]*Customer)  {
	c.peopleInLine= inVal
}
func (c *Checkout) GetPeopleInLine() map[int]*Customer {
	return c.peopleInLine
}

func (c *Checkout) GetAverageWaitTime() float32 {
	return c.averageWaitTime
}

func (c *Checkout) SetAverageWaitTime(inVal float32) {
	c.averageWaitTime=inVal
}

func (c *Checkout) GetProcessedProducts() int {
	return c.processedProducts
}
func (c *Checkout) SetProcessedProducts(inVal int)  {
	c.processedProducts=inVal
}
func (c *Checkout) GetLineLenght() int {
	return c.lineLength
}
func (c *Checkout) SetLineLength(inVal int)  {
	c.lineLength=inVal
}

func (c *Checkout) GetProcessedCustomers() int {
	return c.processedCustomers
}
func (c *Checkout) SetProcessedCustomers(inVal int)  {
	c.processedCustomers=inVal
}
func (c *Checkout) GetSpeed() int {
	return c.speed
}
func (c *Checkout) SetSpeed(inVal int)  {
	c.speed=inVal
}

func (c *Checkout) GetIsOpen() bool {
	return c.isOpen
}

func (c *Checkout) SetIsOpen(inVal bool) {
	c.isOpen= inVal
}

func (c * Checkout) AddPersonToLine(customer *Customer){
	c.peopleInLine[c.lineLength]=customer
	c.lineLength++
}

func (c * Checkout) RemovePersonToLine(customer *Customer){
	delete(c.peopleInLine, c.lineLength)
	c.lineLength--
}

func Open(c * Checkout){
	c.SetIsOpen(true)
}

func Close(c * Checkout){
	c.SetIsOpen(false)
}

 func (c * Checkout) ProcessCustomer(customer *Customer){
	var timeProcessed =0
	trolley :=customer.GetTrolley()
	products :=trolley.GetProducts()
	for key,element := range(products){
		timeProcessed=timeProcessed+element.GetTime()
	}
 }



