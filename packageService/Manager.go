package packageService

import (
	"fmt"
	"sync"
	"time"
)

// Const variables for checkouts, customer per checkout and trolleys
const (
	NUM_CHECKOUTS              = 8
	NUM_SMALL_CHECKOUTS        = 4
	MAX_CUSTOMERS_PER_CHECKOUT = 6
	NUM_TROLLEYS               = 500
)

// Global array for the 3 different trolley sizes, small, medium and large
var TROLLEY_SIZES = [...]int{10, 100, 200}

// Enum for channel switch (iota = 0, 1, 2, 3, 4)
const (
	CUSTOMER_NEW = iota
	CUSTOMER_CHECKOUT
	CUSTOMER_FINISHED
	CUSTOMER_LOST
	CUSTOMER_BAN
)

var (
	// Input from the user
	productsRate int64
	customerRate float64
	processSpeed float64

	// Channel for the customer status (Shopping, at checkout, finished)
	customerStatusChan chan int
	// Channel for the checkout status (Open, Close)
	checkoutChangeStatusChan chan int

	// Global stats
	numberOfCurrentCustomersShopping   int
	numberOfCurrentCustomersAtCheckout int
	totalNumberOfCustomersInStore      int
	totalNumberOfCustomersToday        int
	numberOfCheckoutsOpen              int
	numCustomersLost                   int
	numCustomersBanned                 int
	customerProcessTimeTotal           int64
	customerWaitTimeTotal              int64
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

	// Default to 1 Checkout when the store opens
	numberOfCheckoutsOpen = 1

	return &Manager{id: id, wg: wg}
}

// Creates a switch for updating various customer stats
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
			numCustomersBanned++
			numberOfCurrentCustomersShopping--
			totalNumberOfCustomersInStore--
		default:
			fmt.Println("UH-OH: THINGS JUST GOT SPICY. 🌶🌶🌶")
		}
	}
}

func GetNumCheckouts() int {
	return NUM_CHECKOUTS + NUM_SMALL_CHECKOUTS
}

// Listener for checkout open, close
func (m *Manager) OpenCloseCheckoutListener() {
	for {
		numberOfCheckoutsOpen += <-checkoutChangeStatusChan

		// Check is Supermarket is closing and no customer
		if !m.supermarket.openStatus && totalNumberOfCustomersInStore == 0 {
			break
		}
	}
}

func (m *Manager) GetSupermarket() *Supermarket {
	return m.supermarket
}

// Opens the Supermarket and start go routines for channels and printing the updated stats
func (m *Manager) OpenSupermarket() {
	// Create a Supermarket
	m.supermarket = NewSupermarket()

	go m.CustomerStatusChangeListener()
	go m.OpenCloseCheckoutListener()

	go m.StatPrint()
}

// Prints the current stats of the Supermarket using carriage return
func (m *Manager) StatPrint() {
	for {
		fmt.Printf("Total Customers Today: %03d, Total Customers In Store: %03d, Total Customers Shopping: %02d,"+
			" Total Customers At Checkout: %02d, Checkouts Open: %d, Checkouts Closed: %d, Available Trolleys: %03d"+
			", Customers Lost: %02d, Customers Banned: %d\r",
			totalNumberOfCustomersToday, totalNumberOfCustomersInStore, numberOfCurrentCustomersShopping,
			numberOfCurrentCustomersAtCheckout, numberOfCheckoutsOpen, NUM_CHECKOUTS+NUM_SMALL_CHECKOUTS-numberOfCheckoutsOpen,
			NUM_TROLLEYS-totalNumberOfCustomersInStore, numCustomersLost, numCustomersBanned)
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

// Gets the average customer wait time and process time
func GetCustomerTimesInSeconds() (string, string) {
	avgWait := float64(customerWaitTimeTotal) / float64(totalNumberOfCustomersToday-numCustomersLost)
	avgProcess := float64(customerProcessTimeTotal) / float64(totalNumberOfCustomersToday-numCustomersLost)

	avgWait /= float64(time.Second)
	avgProcess /= float64(time.Second)

	sWait := fmt.Sprintf("%dm %ds", int(avgWait)/60, int(avgWait)%60)
	sProcess := fmt.Sprintf("%dm %ds", int(avgProcess)/60, int(avgProcess)%60)
	return sWait, sProcess
}

// Closes the supermarket
func (m *Manager) CloseSupermarket() {
	m.supermarket.openStatus = false
}
