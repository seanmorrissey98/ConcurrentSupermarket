package packageService

import (
	"fmt"
	"time"
)

const NUM_CHECKOUTS = 8
const MAX_CUSTOMERS_PER_CHECKOUT = 8
const NUM_TROLLEYS = 200

var TROLLEY_SIZES = [...]int{10, 100, 200}

var (
	productsRate int64
	customerRate int
	processSpeed float64

	customerToCheckoutChan   chan int
	finishedAtCheckoutChan   chan int
	checkoutChangeStatusChan chan int
)

type Manager struct {
	id                                 int
	numberOfCurrentCustomersShopping   int
	numberOfCurrentCustomersAtCheckout int
	totalNumberOfCustomersInStore      int
	totalNumberOfCustomersToday        int
	numberOfCheckoutsOpen              int
	numberOfCheckoutsClosed            int
	//name string
}

// Manager Constructor
func NewManager(id int, pr int64, cr int, ps float64) *Manager {
	productsRate = pr
	customerRate = cr
	processSpeed = ps

	customerToCheckoutChan = make(chan int, 256)
	finishedAtCheckoutChan = make(chan int, 256)
	checkoutChangeStatusChan = make(chan int, 256)

	return &Manager{id: id, numberOfCheckoutsOpen: 1, numberOfCheckoutsClosed: NUM_CHECKOUTS - 1}
}

func (m *Manager) NewCustomerStat() {
	m.numberOfCurrentCustomersShopping++
	m.totalNumberOfCustomersInStore++
	m.totalNumberOfCustomersToday++
}

func (m *Manager) CustomerToCheckoutListener() {
	for {
		<-customerToCheckoutChan
		m.numberOfCurrentCustomersShopping--
		m.numberOfCurrentCustomersAtCheckout++
	}
}

func (m *Manager) CustomerFinishedShoppingListener() {
	for {
		<-finishedAtCheckoutChan
		m.numberOfCurrentCustomersAtCheckout--
	}
}

func (m *Manager) OpenCloseCheckoutListener() {
	for {
		checkoutChange := <-checkoutChangeStatusChan
		if checkoutChange > 0 {
			m.numberOfCheckoutsOpen++
			m.numberOfCheckoutsClosed--
		} else {
			m.numberOfCheckoutsClosed++
			m.numberOfCheckoutsOpen--
		}
	}
}

// Generates infinite amount of customers in the supermarket at the rate provided bu the user
func generateCustomer(m *Manager, s *Supermarket) {

}

func (m *Manager) OpenSupermarket() {
	// Create a Supermarket
	NewSupermarket()
	// Start to create customers in the supermarket
	//go generateCustomer(m, &s)

	go m.CustomerToCheckoutListener()
	go m.CustomerFinishedShoppingListener()
	go m.OpenCloseCheckoutListener()

	//go m.StatPrint()
}

func (m *Manager) StatPrint() {
	for {
		fmt.Printf("Total Customers Today: %d, Total Customers In Store: %d, Total Customers Shopping: %d,"+
			" Total Customers At Checkout: %d, Checkouts Open: %d, Checkouts Closed: %d\r",
			m.totalNumberOfCustomersToday, m.totalNumberOfCustomersInStore, m.numberOfCurrentCustomersShopping,
			m.numberOfCurrentCustomersAtCheckout, m.numberOfCheckoutsOpen, m.numberOfCheckoutsClosed)

		time.Sleep(time.Millisecond * 40)
	}
}

func CloseSupermarket(supermarketId int) {}

func OpenCheckout(checkoutId int) {}

func CloseCheckout(checkoutId int) {}

func BanCustomer(customerId int) {}
