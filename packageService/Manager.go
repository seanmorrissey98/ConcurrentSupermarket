package packageService

import (
	"fmt"
	"sync"
	"time"
)

const NUM_CHECKOUTS = 8
const MAX_CUSTOMERS_PER_CHECKOUT = 6
const NUM_TROLLEYS = 500

var TROLLEY_SIZES = [...]int{10, 100, 200}

var (
	productsRate int64
	customerRate int
	processSpeed float64

	newCustomerChan          chan int
	customerToCheckoutChan   chan int
	finishedAtCheckoutChan   chan int
	checkoutChangeStatusChan chan int

	numberOfCurrentCustomersShopping   int
	numberOfCurrentCustomersAtCheckout int
	totalNumberOfCustomersInStore      int
	totalNumberOfCustomersToday        int
	numberOfCheckoutsOpen              int
	numberOfCheckoutsClosed            int
)

type Manager struct {
	id          int
	supermarket *Supermarket
	wg          *sync.WaitGroup
	//name string
}

// Manager Constructor
func NewManager(id int, wg *sync.WaitGroup, pr int64, cr int, ps float64) *Manager {
	productsRate = pr
	customerRate = cr
	processSpeed = ps

	newCustomerChan = make(chan int, 256)
	customerToCheckoutChan = make(chan int, 256)
	finishedAtCheckoutChan = make(chan int, 256)
	checkoutChangeStatusChan = make(chan int, 256)

	numberOfCheckoutsOpen = 1
	numberOfCheckoutsClosed = NUM_CHECKOUTS - 1

	return &Manager{id: id, wg: wg}
}

func (m *Manager) NewCustomerListener() {
	for {
		input := <-newCustomerChan
		if input < 0 {
			break
		}

		numberOfCurrentCustomersShopping++
		totalNumberOfCustomersInStore++
		totalNumberOfCustomersToday++
	}
}

func (m *Manager) CustomerToCheckoutListener() {
	for {
		<-customerToCheckoutChan
		numberOfCurrentCustomersShopping--
		numberOfCurrentCustomersAtCheckout++

		if !m.supermarket.openStatus && numberOfCurrentCustomersShopping == 0 {
			break
		}
	}
}

func (m *Manager) CustomerFinishedShoppingListener() {
	for {
		<-finishedAtCheckoutChan
		numberOfCurrentCustomersAtCheckout--
		totalNumberOfCustomersInStore--

		if !m.supermarket.openStatus && totalNumberOfCustomersInStore == 0 {
			break
		}
	}
}

// Shuts down all manager channels to avoid leaking
func (m *Manager) ShutDownChannels() {

}

func (m *Manager) OpenCloseCheckoutListener() {
	for {
		checkoutChange := <-checkoutChangeStatusChan
		if checkoutChange > 0 {
			numberOfCheckoutsOpen++
			numberOfCheckoutsClosed--
		} else {
			numberOfCheckoutsClosed++
			numberOfCheckoutsOpen--
		}

		if !m.supermarket.openStatus && totalNumberOfCustomersInStore == 0 {
			break
		}
	}
}

func (m *Manager) GetSupermarket() *Supermarket {
	return m.supermarket
}

func (m *Manager) OpenSupermarket() {
	// Create a Supermarket
	m.supermarket = NewSupermarket()

	go m.NewCustomerListener()
	go m.CustomerToCheckoutListener()
	go m.CustomerFinishedShoppingListener()
	go m.OpenCloseCheckoutListener()

	go m.StatPrint()
}

func (m *Manager) StatPrint() {
	for {
		fmt.Printf("Total Customers Today: %d, Total Customers In Store: %d, Total Customers Shopping: %d,"+
			" Total Customers At Checkout: %d, Checkouts Open: %d, Checkouts Closed: %d,"+
			" Available Trolleys: %d\r",
			totalNumberOfCustomersToday, totalNumberOfCustomersInStore, numberOfCurrentCustomersShopping,
			numberOfCurrentCustomersAtCheckout, numberOfCheckoutsOpen, numberOfCheckoutsClosed,
			NUM_TROLLEYS-totalNumberOfCustomersInStore)

		time.Sleep(time.Millisecond * 40)

		if !m.supermarket.openStatus && totalNumberOfCustomersInStore == 0 {
			break
		}
	}

	m.wg.Done()
}

func (m *Manager) CloseSupermarket() {
	m.supermarket.openStatus = false
	newCustomerChan <- -1
}
