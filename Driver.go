package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime/trace"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

func userInput(inVal string, rangeLower float64, rangeHigher float64, ok bool) string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println(inVal)
	var input string
	for scanner.Scan() {
		if ok {
			i, err := strconv.ParseInt(scanner.Text(), 10, 64)
			if err == nil && float64(i) >= rangeLower && float64(i) <= rangeHigher {
				input = scanner.Text()
				break
			}
		} else {
			i, err := strconv.ParseFloat(scanner.Text(), 64)
			if err == nil && i >= rangeLower && i <= rangeHigher {
				input = scanner.Text()
				break
			}
		}
		fmt.Println(inVal)
	}
	return input
}

func main() {
	// Trace for monitoring go routines
	f, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = trace.Start(f)
	if err != nil {
		panic(err)
	}
	defer trace.Stop()

	// Start random seed
	rand.Seed(time.Now().UnixNano())

	// Get required inputs from the user
	productsRate, _ := strconv.ParseInt(userInput("Please enter the range of products per trolley. (1-200):", 1, 200, true), 10, 64)
	customerRate, _ := strconv.Atoi(userInput("Please enter the rate customers arrive at checkouts. (1-60):", 1, 60, true))
	processSpeed, _ := strconv.ParseFloat(userInput("Please enter the range for product processing speed. (0.5-6):", 0.5, 6, false), 64)
	// Print the inputs back to the user
	fmt.Println("Products Rate:", productsRate)
	fmt.Println("Customer Rate:", customerRate)
	fmt.Printf("%s %f", "Process Speed:", processSpeed)

	// Add a WaitGroup for Supermarket closing when the Enter key is clicked
	var wg sync.WaitGroup
	wg.Add(1)

	// Create manager agent and start Open a Supermarket
	m := NewManager(1, &wg, productsRate, float64(customerRate), processSpeed)
	m.OpenSupermarket()

	// Locks program running, must be at the end of main
	fmt.Println("\nPress Enter at any time to terminate simulation...")
	input := bufio.NewScanner(os.Stdin)
	// Waits for Enter to be clicked
	input.Scan()

	fmt.Println("\nSupermarket CLosing...")

	// Start graceful shutdown of the Supermarket
	m.CloseSupermarket()

	// Wait for the Supermarket to close and the channels and go routines to shut down
	wg.Wait()

	// Get the supermarket metrics for Statistics print
	supermarket := m.GetSupermarket()
	checkouts := supermarket.GetAllCheckouts()
	totalProcessedCustomers := getTotalProcessedCustomers(checkouts)
	totalProcessedProducts := getTotalProcessedProducts(checkouts)
	fmt.Println()

	// Sort the Checkouts array for print
	sort.SliceStable(checkouts, func(i, j int) bool {
		return checkouts[i].Number < checkouts[j].Number
	})

	// Print the Checkout stats in order of checkout number
	PrintCheckoutStats(checkouts, totalProcessedCustomers, totalProcessedProducts)
}

func PrintCheckoutStats(checkouts []*Checkout, totalProcessedCustomers int64, totalProcessedProducts int64) {
	var highest int64 = 0
	var totalUtilization float64
	for i := range checkouts {
		if checkouts[i].GetFirstCustomerArrivalTime()+checkouts[i].GetProcessedProductsTime() > highest {
			highest = checkouts[i].GetFirstCustomerArrivalTime() + checkouts[i].GetProcessedProductsTime()
		}
	}

	for i := range checkouts {
		checkout := checkouts[i]
		fmt.Printf("Checkout: #%d\n", checkout.Number)
		// Utilization based on the amount of customers the checkout processed in comparison to all the customers who were in the shop.
		//figure := float64(checkout.GetTotalCustomersProcessed()) / float64(totalProcessedCustomers) * 100

		// Utilization based on time checkout was open compared to time shop was open.
		figure := float64(checkout.GetProcessedProductsTime()) / float64(highest) * 100
		totalUtilization += figure
		fmt.Printf("Utilisation: %.2f%s\n", figure, "%")
		productsProcessed := checkout.ProcessedProducts
		fmt.Printf("Products Processed: %d\n", productsProcessed)
		percentProducts := float64(productsProcessed) / float64(totalProcessedProducts) * 100
		fmt.Printf("Total Products Processed (%%): %.2f%s\n\n", percentProducts, "%")
	}

	total := GetTotalNumberOfCustomersToday()
	fmt.Printf("Average Products Per Trolley: %.2f\n\n", float64(totalProcessedProducts)/float64(total))

	avgWait, avgProcess := GetCustomerTimesInSeconds()
	fmt.Printf("Average Customer Wait Time: %s, \nAverage Customer Process Time: %s\n", avgWait, avgProcess)

	fmt.Printf("Average Checkout Utilisation: %.2f%s\n", totalUtilization/float64(GetNumCheckouts()), "%")
}

func getTotalProcessedProducts(c []*Checkout) int64 {
	var total int64
	total = 0
	for i := range c {
		total += c[i].ProcessedProducts
	}
	return total
}

func getTotalProcessedCustomers(c []*Checkout) int64 {
	var total int64
	total = 0
	for i := range c {
		total += c[i].ProcessedCustomers
	}
	return total
}

type Checkout struct {
	Number                   int
	tenOrLess                bool
	isSeniorCheckout         bool
	isSelfCheckout           bool
	hasScanner               bool
	inUse                    bool
	lineLength               int
	isLineFull               bool
	peopleInLine             chan *Customer
	averageWaitTime          float32
	ProcessedProducts        int64
	ProcessedCustomers       int64
	speed                    float64
	isOpen                   bool
	finishedProcessing       chan int
	firstCustomerArrivalTime int64
	processedProductsTime    int64
}

// Checkout Constructor
func NewCheckout(number int, tenOrLess bool, isSeniorCheckout bool, isSelfCheckout bool, hasScanner bool, inUse bool, lineLength int, isLineFull bool, peopleInLine chan *Customer, averageWaitTime float32, processedProducts int64, processedCustomers int64, speed float64, isOpen bool, finishedProcessing chan int) *Checkout {
	c := Checkout{number, tenOrLess, isSeniorCheckout, isSelfCheckout, hasScanner, inUse, lineLength, isLineFull, peopleInLine, averageWaitTime, processedProducts, processedCustomers, speed, isOpen, finishedProcessing, 0, 0}

	if c.hasScanner {
		c.speed = 0.5
	} else {
		c.speed = 1.0
	}

	// Starts a goroutine for processing all products in a trolley
	if isOpen {
		go c.ProcessCheckout()
	}

	return &c
}

// Gets the number of customers in a checkout line
func (c *Checkout) GetNumPeopleInLine() int {
	return len(c.peopleInLine)
}

// Adds a customer a specific checkout line
func (c *Checkout) AddPersonToLine(customer *Customer) {
	// Use channel instead a list of customers to easily pop and send the customer
	customer.waitTime = time.Now().UnixNano()
	c.peopleInLine <- customer
	c.lineLength++
}

func (c *Checkout) GetProcessedProductsTime() int64 {
	return c.processedProductsTime
}

func (c *Checkout) GetFirstCustomerArrivalTime() int64 {
	return c.firstCustomerArrivalTime
}

// Processes all products in a customers trolley
func (c *Checkout) ProcessCheckout() {
	for {
		if !c.isOpen && c.lineLength == 0 {
			break
		}
		// Get the first customer in line
		customer := <-c.peopleInLine
		// Check if customer is nil, break open of for loop and set checkout open to false
		if customer == nil {
			c.isOpen = false
			break
		}

		if c.ProcessedCustomers == 0 {
			c.firstCustomerArrivalTime = customer.shopTime
		}
		c.lineLength--

		// Start customer wait timer
		customer.waitTime = time.Now().UnixNano() - customer.waitTime

		trolley := customer.trolley
		products := trolley.products

		age := customer.age
		var ageMultiplier float64
		ageMultiplier = 1
		if age > 65 {
			ageMultiplier = 1.5
		}

		// Start customer process timer
		customer.processTime = time.Now().UnixNano()

		// Get all products in trolley and calculate the time to wait
		for _, p := range products {
			time.Sleep(time.Millisecond * time.Duration(p.GetTime()*500*c.speed*ageMultiplier))
			atomic.AddInt64(&c.ProcessedProducts, 1)
			atomic.AddInt64(&c.processedProductsTime, int64(p.GetTime()*500*c.speed))
		}

		// Stop customer process timer
		customer.processTime = time.Now().UnixNano() - customer.processTime

		// Send customer is to finished process channel
		c.finishedProcessing <- customer.id

		// Increments the processed customer after customer is finished ar checkout
		atomic.AddInt64(&c.ProcessedCustomers, 1)
	}
}

func (c *Checkout) Open() {
	c.isOpen = true
	go c.ProcessCheckout()
}

// Passes a nil customer to the peopleInLine channel
func (c *Checkout) Close() {
	c.peopleInLine <- nil
}

type Customer struct {
	id          int
	name        string
	trolley     *Trolley
	age         int
	impatient   bool
	gender      string
	mutex       sync.Mutex
	processTime int64
	waitTime    int64
	shopTime    int64
}

// Shop lets the customer get products and add them to their trolley until the reach capacity of trolley or break the random < 0.05
func (c *Customer) Shop(readyForCheckoutChan chan int) {

	var speedMultiplier float64
	speedMultiplier = 1

	// Infinite loop of customer shopping
	for {
		if c.GetNumProducts() == int(productsRate) {
			break
		}

		if c.age > 65 {
			speedMultiplier = 1.5
		}

		p := NewProduct()
		time.Sleep(time.Millisecond * time.Duration(p.GetTime()*200*speedMultiplier))
		c.trolley.AddProductToTrolley(p)
		c.shopTime += int64(p.GetTime() * 200)
		if c.trolley.IsFull() {
			break
		}

		if rand.Float64() < 0.05 {
			break
		}
	}

	// Notify the channel in the supermarket FinishedShoppingListener() by sending the customer id to it
	readyForCheckoutChan <- c.id
}

func (c *Customer) GetNumProducts() int {
	return len(c.trolley.products)
}

// Const variables for checkouts, customer per checkout and trolleys
const (
	NUM_CHECKOUTS              = 6
	NUM_SMALL_CHECKOUTS        = 2
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

type Product struct {
	time float64
}

// Product Constructor
func NewProduct() *Product {
	p := Product{rand.Float64() * processSpeed}
	return &p
}

func (p *Product) GetTime() float64 {
	return p.time
}

var trolleyMutex *sync.Mutex
var customerMutex *sync.RWMutex
var checkoutMutex *sync.RWMutex

type Supermarket struct {
	customerCount    int
	openStatus       bool
	checkoutOpen     []*Checkout
	checkoutClosed   []*Checkout
	customers        map[int]*Customer
	trolleys         []*Trolley
	finishedShopping chan int
	finishedCheckout chan int
}

// Constructor for Supermarket
func NewSupermarket() *Supermarket {
	trolleyMutex = &sync.Mutex{}
	customerMutex = &sync.RWMutex{}
	checkoutMutex = &sync.RWMutex{}

	s := Supermarket{0, true, make([]*Checkout, 0, 256), make([]*Checkout, 0, 256), make(map[int]*Customer), make([]*Trolley, NUM_TROLLEYS), make(chan int), make(chan int)}
	s.GenerateTrolleys()
	s.GenerateCheckouts()

	go s.GenerateCustomer()
	go s.FinishedShoppingListener()
	go s.FinishedCheckoutListener()

	return &s
}

// Create a customer and adds them to to the customers map in supermarket
func (s *Supermarket) GenerateCustomer() {
	for {
		// Check is Supermarket is closing
		if !s.openStatus {
			break
		}

		time.Sleep(time.Millisecond * time.Duration(rand.Intn(int((1.0/customerRate)*10000))))

		// Checks if the Supermarket it out of trolleys, customer will not enter store
		if len(s.trolleys) == 0 {
			continue
		}

		isImpatient := rand.Float64() < 0.15

		// Create a new customer with an id = the number they are created at in the supermarket
		c := &Customer{id: s.customerCount, impatient: isImpatient, age: 20 + rand.Intn(50)}
		//fmt.Printf("Total num of customers so far: %d\n", s.numOfTotalCustomers)

		// Create 3 different trolley sizes modelling a basket, small trolley and large trolley
		trolleySize := TROLLEY_SIZES[rand.Intn(len(TROLLEY_SIZES))]

		// A customer picks a trolley based on the amount of products they need
		outOfTrolleys := false
		for i, t := range s.trolleys {
			if t.capacity == trolleySize {
				c.trolley = t

				trolleyMutex.Lock()
				s.trolleys[i] = s.trolleys[len(s.trolleys)-1]
				s.trolleys = s.trolleys[:len(s.trolleys)-1]
				trolleyMutex.Unlock()
				break
			} else if i == len(s.trolleys)-1 {
				//fmt.Println("No More Trolleys of Size: ", trolleySize)
				outOfTrolleys = true
			}
		}
		if outOfTrolleys {
			continue
		}

		s.customerCount++
		// Add customer to stat print
		customerStatusChan <- CUSTOMER_NEW

		// Add customer to the customers map in supermarket, key=customer.id, value=customer
		customerMutex.Lock()
		s.customers[c.id] = c
		customerMutex.Unlock()

		// Customer can now go add products to the trolley
		go c.Shop(s.finishedShopping)

		// Decides to open or close checkouts
		s.CalculateOpenCheckout()
	}
}

// Sends customer to a checkout
func (s *Supermarket) SendToCheckout(id int) {
	customerMutex.RLock()
	c := s.customers[id]
	customerMutex.RUnlock()

	// Choose the best checkout for a customer to go to
	checkout, pos := s.ChooseCheckout(c.GetNumProducts(), c.impatient)
	// No checkout with < max number in queue - The number of lost customers (Customers will leave the store if they need to join a queue more than six deep)
	if pos < 0 {
		s.CustomerLeavesStore(id)
		customerStatusChan <- CUSTOMER_LOST
		return
	}

	// Checks if customer is impatient and joins a ten or less checkout with more tha 10 items
	// Manager has a 50% chance of finding them and banning them
	if c.impatient && c.GetNumProducts() > 10 && checkout.tenOrLess && rand.Float64() < 0.5 {
		s.CustomerLeavesStore(id)
		customerStatusChan <- CUSTOMER_BAN
		return
	}

	checkout.AddPersonToLine(c)

	// Change the status channel of customer, sends a 1
	customerStatusChan <- CUSTOMER_CHECKOUT
}

// Gets the best open checkout for a customer to go to at the current time
func (s *Supermarket) ChooseCheckout(numProducts int, isImpatient bool) (*Checkout, int) {
	min, pos := -1, -1

	checkoutMutex.RLock()
	for i := 0; i < len(s.checkoutOpen); i++ {
		// Gets the number of people in the checkout and if the checkout is 'tenOrLess'
		// Checks if the customer can join the checkout (less than max number (6) allowed)
		// Ensure only customers with 10 or less items can go to the 10 or less checkouts
		// Added impatience variable
		// Finds the checkout with the least amount of people
		if num, tenOrLess := s.checkoutOpen[i].GetNumPeopleInLine(), s.checkoutOpen[i].tenOrLess; ((tenOrLess && (numProducts <= 10 || isImpatient)) || !tenOrLess) && (num < min || min < 0) && num < MAX_CUSTOMERS_PER_CHECKOUT {
			min, pos = num, i
		}
	}
	checkoutMutex.RUnlock()

	var c *Checkout
	if pos >= 0 {
		c = s.checkoutOpen[pos]
	}

	return c, pos
}

// Generates 200 trolleys in the supermarket
func (s *Supermarket) GenerateTrolleys() {
	for i := 0; i < NUM_TROLLEYS; i++ {
		s.trolleys[i] = NewTrolley(TROLLEY_SIZES[rand.Intn(len(TROLLEY_SIZES))])
	}
}

// Generates 8 checkouts
func (s *Supermarket) GenerateCheckouts() {
	var hasScanner bool
	// Default create 8 Checkouts when Supermarket is created
	for i := 0; i < NUM_CHECKOUTS+NUM_SMALL_CHECKOUTS-1; i++ {
		hasScanner := rand.Float64() < 0.5
		if i == 0 {
			s.checkoutOpen = append(s.checkoutOpen, NewCheckout(i+1, false, false, false, hasScanner, false, 0, false, make(chan *Customer, MAX_CUSTOMERS_PER_CHECKOUT), 0, 0, 0, 0, true, s.finishedCheckout))
		} else {
			s.checkoutClosed = append(s.checkoutClosed, NewCheckout(i+1, i >= NUM_CHECKOUTS, false, false, hasScanner, false, 0, false, make(chan *Customer, MAX_CUSTOMERS_PER_CHECKOUT), 0, 0, 0, 0, false, s.finishedCheckout))
		}
	}

	s.checkoutClosed = append(s.checkoutClosed, NewCheckout(NUM_CHECKOUTS+NUM_SMALL_CHECKOUTS, true, true, false, hasScanner, false, 0, false, make(chan *Customer, MAX_CUSTOMERS_PER_CHECKOUT), 0, 0, 0, 0, false, s.finishedCheckout))
}

// Waits for a customer to finish shopping using a channel, then sends the customer to a checkout
func (s *Supermarket) FinishedShoppingListener() {
	for {
		if !s.openStatus && numberOfCurrentCustomersShopping == 0 {
			break
		}

		// Check if customer is finished adding products to trolley using channel from the shop() method in Customer.go
		id := <-s.finishedShopping

		// Send customer to a checkout
		s.SendToCheckout(id)
	}
}

// Waits for a customer to finish at a checkout using a channel, then removes the customer from the supermarket
func (s *Supermarket) FinishedCheckoutListener() {
	for {
		if !s.openStatus && totalNumberOfCustomersInStore == 0 {
			break
		}

		// Check if customer is finished at a checkout when all products are processed
		id := <-s.finishedCheckout

		// Updating total wait and process time to get the average later
		// Doesn't need mutex, accessed 1 at a time here
		customer := s.customers[id]
		customerProcessTimeTotal += customer.processTime
		customerWaitTimeTotal += customer.waitTime

		// Empty the customers trolley
		s.CustomerLeavesStore(id)

		customerStatusChan <- CUSTOMER_FINISHED

		s.CalculateOpenCheckout()
	}
}

// Cleans up customer and trolley items when they leave a shop
func (s *Supermarket) CustomerLeavesStore(id int) {
	customerMutex.RLock()
	trolley := s.customers[id].trolley
	customerMutex.RUnlock()

	trolley.EmptyTrolley()
	// Adds the trolley the customer was using back into the trolleys slice in the supermarket
	trolleyMutex.Lock()
	s.trolleys = append(s.trolleys, trolley)
	trolleyMutex.Unlock()

	// Remove customer from the supermarket
	customerMutex.Lock()
	delete(s.customers, id)
	customerMutex.Unlock()
}

// Calculates the threshold for opening / closing a checkout
func (s *Supermarket) CalculateOpenCheckout() {
	numOfCurrentCustomers := len(s.customers)
	numOfOpenCheckouts := len(s.checkoutOpen)
	calculationOfThreshold := int(math.Ceil(float64(numOfCurrentCustomers) / MAX_CUSTOMERS_PER_CHECKOUT))

	// Ensure at least 1 checkout stays open
	if numOfCurrentCustomers == 0 && s.openStatus {
		return
	}

	if len(s.checkoutOpen) == 1 {
		if s.checkoutOpen[0].isSeniorCheckout {
			s.checkoutOpen[0].isSeniorCheckout = false
		}

		if s.checkoutOpen[0].tenOrLess {
			s.checkoutOpen[0].tenOrLess = false
		}
	}

	// Calculate threshold for opening a checkout
	if calculationOfThreshold > numOfOpenCheckouts {
		// If there are no more checkouts to open
		if len(s.checkoutClosed) == 0 {
			//fmt.Printf("All checkouts currently open. The current number of customers is: %d\n", numOfCurrentCustomers)
			return
		}

		// Open first checkout in closed checkout slice
		s.checkoutClosed[0].Open()
		s.checkoutOpen = append(s.checkoutOpen, s.checkoutClosed[0])
		s.checkoutClosed = s.checkoutClosed[1:]

		checkoutChangeStatusChan <- 1

		return
	}

	// Calculate threshold for closing a checkout
	if calculationOfThreshold < numOfOpenCheckouts {
		if len(s.checkoutOpen) == 1 && s.openStatus {
			//fmt.Printf("We only have one checkout open. Number of customer: %d\n", numOfCurrentCustomers)
			return
		}

		// Choose best checkout to close
		checkout, pos := s.ChooseCheckout(0, false)
		if pos < 0 {
			return
		}
		checkout.Close()
		s.checkoutClosed = append(s.checkoutClosed, checkout)
		s.checkoutOpen = append(s.checkoutOpen[0:pos], s.checkoutOpen[pos+1:]...)

		checkoutChangeStatusChan <- -1

		//fmt.Printf("1 checkout just closed. We now have %d open checkouts.\n", len(s.checkoutOpen))

		return
	}
}

// returns a slice of all of the checkouts in the supermarket
func (s *Supermarket) GetAllCheckouts() []*Checkout {
	return append(s.checkoutOpen, s.checkoutClosed...)
}

// Trolley struct for holding products
type Trolley struct {
	capacity int
	products []*Product
}

// Trolley Constructor
func NewTrolley(capacity int) *Trolley {
	t := Trolley{capacity, make([]*Product, 0, capacity)}
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
	t.products = make([]*Product, 0, t.capacity)
}

// Weather struct which affects customer generation
type Weather struct {
	status    int
	forecasts [4]string
}

// Initializes the forecast array of string to
// include 4 different weather types.
func (w *Weather) InitializeWeather() {
	w.forecasts[0] = "SUNNY DAYS" //1.25
	w.forecasts[1] = "RAINY DAYS" // .75
	w.forecasts[2] = "CLEAR DAY"  // 1
	w.forecasts[3] = "SNOWY DAY"  //.5
}

// Returns a string of the current weather forecast i.e. "SUNNY DAYS"
// and alos returns a float64 value which is used as a multiplyer for
// customers entering the shop
func (w *Weather) GetWeather() (string, float64) {
	//return w.forecasts[w.status]
	multipliers := [4]float64{1.25, 0.75, 1, 0.5}
	return w.forecasts[w.status], multipliers[w.status]
}

// Sets the weathers status equal to forecastIndex
func (w *Weather) ChangeWeather(forecastIndex int) {
	w.status = forecastIndex
}

// Generates a random number between 0-3 and sets the weather
// status to the random number.
func (w *Weather) GenerateWeather() {
	rand.Seed(time.Now().UnixNano())
	weatherRand := rand.Intn(4)
	w.ChangeWeather(weatherRand)
}
