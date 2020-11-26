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

const (
	CUSTOMER_NEW = iota
	CUSTOMER_CHECKOUT
	CUSTOMER_FINISHED
	CUSTOMER_LOST
	CUSTOMER_BAN
)

var (
	productsRate int64
	customerRate float64
	processSpeed float64

	customerStatusChan       chan int
	checkoutChangeStatusChan chan int

	numberOfCurrentCustomersShopping   int
	numberOfCurrentCustomersAtCheckout int
	totalNumberOfCustomersInStore      int
	totalNumberOfCustomersToday        int
	numberOfCheckoutsOpen              int
	numCustomersLost                   int
)

type Manager struct {
	id          int
	supermarket *Supermarket
	wg          *sync.WaitGroup
	//name string
}

// Manager Constructor
func NewManager(id int, wg *sync.WaitGroup, pr int64, cr float64, ps float64) *Manager {
	var weather Weather
	weather.InitializeWeather()
	weather.GenerateWeather()
	forecast, multiplier := weather.GetWeather()
	fmt.Printf("\nCURRENT FORECAST: %s\n", forecast)

	productsRate = pr
	customerRate = cr * multiplier
	processSpeed = ps

	customerStatusChan = make(chan int, 256)
	checkoutChangeStatusChan = make(chan int, 256)

	numberOfCheckoutsOpen = 1

	return &Manager{id: id, wg: wg}
}

func (m *Manager) CustomerStatusChangeListener() {
	for {
		input := <-customerStatusChan

		switch input {
		case CUSTOMER_NEW:
			numberOfCurrentCustomersShopping++
			totalNumberOfCustomersInStore++
			totalNumberOfCustomersToday++

		case CUSTOMER_CHECKOUT:
			numberOfCurrentCustomersShopping--
			numberOfCurrentCustomersAtCheckout++

		case CUSTOMER_FINISHED:
			numberOfCurrentCustomersAtCheckout--
			totalNumberOfCustomersInStore--

		case CUSTOMER_LOST:
			numCustomersLost++
			numberOfCurrentCustomersShopping--
			totalNumberOfCustomersInStore--

		case CUSTOMER_BAN:
			// TODO: Ban customers
		default:
			fmt.Println("UH-OH: THINGS JUST GOT SPICY. 🌶🌶🌶")
		}
	}
}

func (m *Manager) OpenCloseCheckoutListener() {
	for {
		numberOfCheckoutsOpen += <-checkoutChangeStatusChan

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

	go m.CustomerStatusChangeListener()
	go m.OpenCloseCheckoutListener()

	go m.StatPrint()
}

func (m *Manager) StatPrint() {
	for {
		fmt.Printf("Total Customers Today: %03d, Total Customers In Store: %03d, Total Customers Shopping: %02d,"+
			" Total Customers At Checkout: %02d, Checkouts Open: %d, Checkouts Closed: %d, Available Trolleys: %03d"+
			", Customers Lost: %02d\r",
			totalNumberOfCustomersToday, totalNumberOfCustomersInStore, numberOfCurrentCustomersShopping,
			numberOfCurrentCustomersAtCheckout, numberOfCheckoutsOpen, NUM_CHECKOUTS-numberOfCheckoutsOpen,
			NUM_TROLLEYS-totalNumberOfCustomersInStore, numCustomersLost)

		time.Sleep(time.Millisecond * 40)

		if !m.supermarket.openStatus && totalNumberOfCustomersInStore == 0 {
			fmt.Printf("\n")
			break
		}
	}

	m.wg.Done()
}

func GetTotalNumberOfCustomersToday() int {
	return totalNumberOfCustomersToday
}

func (m *Manager) CloseSupermarket() {
	m.supermarket.openStatus = false
}
