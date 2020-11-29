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

// validate user input, checks if value is within specified range
// returns string of user input if correct input is supplied
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

// Starts the Programme,asking for user inputs fro product,customer speed etc
// Starts a wait group which monitors supermarket for enter key
// Create Manager object and opens supermarket
// After supermarket closes we get metrics from supermarket simulation and print stats
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
	m := newManager(1, &wg, productsRate, float64(customerRate), processSpeed)
	m.openSupermarket()

	// Locks program running, must be at the end of main
	fmt.Println("\nPress Enter at any time to terminate simulation...")
	input := bufio.NewScanner(os.Stdin)
	// Waits for Enter to be clicked
	input.Scan()

	fmt.Println("\nSupermarket CLosing...")

	// Start graceful shutdown of the Supermarket
	m.closeSupermarket()

	// Wait for the Supermarket to close and the channels and go routines to shut down
	wg.Wait()

	// Get the supermarket metrics for Statistics print
	supermarket := m.getSupermarket()
	checkouts := supermarket.getAllCheckouts()
	totalProcessedCustomers := getTotalProcessedCustomers(checkouts)
	totalProcessedProducts := getTotalProcessedProducts(checkouts)
	fmt.Println()

	// Sort the Checkouts array for print
	sort.SliceStable(checkouts, func(i, j int) bool {
		return checkouts[i].Number < checkouts[j].Number
	})

	// Print the Checkout stats in order of checkout number
	printCheckoutStats(checkouts, totalProcessedCustomers, totalProcessedProducts)
}

// method to print all stats from each checkout after the supermarket closes
func printCheckoutStats(checkouts []*Checkout, totalProcessedCustomers int64, totalProcessedProducts int64) {
	var highest int64 = 0
	var totalUtilization float64
	for i := range checkouts {
		if checkouts[i].getFirstCustomerArrivalTime()+checkouts[i].getProcessedProductsTime() > highest {
			highest = checkouts[i].getFirstCustomerArrivalTime() + checkouts[i].getProcessedProductsTime()
		}
	}

	for i := range checkouts {
		checkout := checkouts[i]
		fmt.Printf("Checkout: #%d\n", checkout.Number)
		// Utilization based on the amount of customers the checkout processed in comparison to all the customers who were in the shop.
		//figure := float64(checkout.GetTotalCustomersProcessed()) / float64(totalProcessedCustomers) * 100

		// Utilization based on time checkout was open compared to time shop was open.
		figure := float64(checkout.getProcessedProductsTime()) / float64(highest) * 100
		totalUtilization += figure
		fmt.Printf("Utilisation: %.2f%s\n", figure, "%")
		productsProcessed := checkout.ProcessedProducts
		fmt.Printf("Products Processed: %d\n", productsProcessed)
		percentProducts := float64(productsProcessed) / float64(totalProcessedProducts) * 100
		fmt.Printf("Total Products Processed (%%): %.2f%s\n\n", percentProducts, "%")
	}

	total := getTotalNumberOfCustomersToday()
	fmt.Printf("Average Products Per Trolley: %d\n", int(float64(totalProcessedProducts)/float64(total)))

	avgWait, avgProcess := getCustomerTimesInSeconds()
	fmt.Printf("Average Customer Wait Time: %s, \nAverage Customer Process Time: %s\n", avgWait, avgProcess)

	fmt.Printf("Average Checkout Utilisation: %.2f%s\n", totalUtilization/float64(getNumCheckouts()), "%")
}

// returns total products proccessed at checkout
func getTotalProcessedProducts(c []*Checkout) int64 {
	var total int64
	total = 0
	for i := range c {
		total += c[i].ProcessedProducts
	}
	return total
}

// returns total proccessed customers at a checkout
func getTotalProcessedCustomers(c []*Checkout) int64 {
	var total int64
	total = 0
	for i := range c {
		total += c[i].ProcessedCustomers
	}
	return total
}

// Checkout struct for processing customers,products
// Stores channel of Customers in line
// has channel of finishedProcessing for ids of customers finshed at checkout
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
func newCheckout(number int, tenOrLess bool, isSeniorCheckout bool, isSelfCheckout bool, hasScanner bool, inUse bool, lineLength int, isLineFull bool, peopleInLine chan *Customer, averageWaitTime float32, processedProducts int64, processedCustomers int64, speed float64, isOpen bool, finishedProcessing chan int) *Checkout {
	c := Checkout{number, tenOrLess, isSeniorCheckout, isSelfCheckout, hasScanner, inUse, lineLength, isLineFull, peopleInLine, averageWaitTime, processedProducts, processedCustomers, speed, isOpen, finishedProcessing, 0, 0}

	if c.hasScanner {
		c.speed = 0.5
	} else {
		c.speed = 1.0
	}

	// Starts a goroutine for processing all products in a trolley
	if isOpen {
		go c.processCheckout()
	}

	return &c
}

// Gets the number of customers in a checkout line
func (c *Checkout) getNumPeopleInLine() int {
	return len(c.peopleInLine)
}

// Adds a customer a specific checkout line
func (c *Checkout) addPersonToLine(customer *Customer) {
	// Use channel instead a list of customers to easily pop and send the customer
	customer.waitTime = time.Now().UnixNano()
	c.peopleInLine <- customer
	c.lineLength++
}

// Gets time of products processed at checkout
func (c *Checkout) getProcessedProductsTime() int64 {
	return c.processedProductsTime
}

// Gets the time it tokk the first customer to get to the checkout
func (c *Checkout) getFirstCustomerArrivalTime() int64 {
	return c.firstCustomerArrivalTime
}

// Processes all products in a customers trolley
// calculates customer processtime
// Increments the processed customer after customer is finished at checkout
func (c *Checkout) processCheckout() {
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
			time.Sleep(time.Millisecond * time.Duration(p.getTime()*500*c.speed*ageMultiplier))
			atomic.AddInt64(&c.ProcessedProducts, 1)
			atomic.AddInt64(&c.processedProductsTime, int64(p.getTime()*500*c.speed))
		}

		// Stop customer process timer
		customer.processTime = time.Now().UnixNano() - customer.processTime

		// Send customer is to finished process channel
		c.finishedProcessing <- customer.id

		// Increments the processed customer after customer is finished ar checkout
		atomic.AddInt64(&c.ProcessedCustomers, 1)
	}
}

// open Checkout
func (c *Checkout) open() {
	c.isOpen = true
	go c.processCheckout()
}

// Passes a nil customer to the peopleInLine channel
func (c *Checkout) close() {
	c.peopleInLine <- nil
}

// Customer Struct storing trolley of products
// Contains mutex
// attributes such as gender and impatience affect shop and process speed
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
func (c *Customer) shop(readyForCheckoutChan chan int) {

	var speedMultiplier float64
	speedMultiplier = 1

	// Infinite loop of customer shopping
	for {
		if c.getNumProducts() == int(productsRate) {
			break
		}

		if c.age > 65 {
			speedMultiplier = 1.5
		}

		p := newProduct()
		time.Sleep(time.Millisecond * time.Duration(p.getTime()*200*speedMultiplier))
		c.trolley.addProductToTrolley(p)
		c.shopTime += int64(p.getTime() * 200)
		if c.trolley.isFull() {
			break
		}

		if rand.Float64() < 0.05 {
			break
		}
	}

	// Notify the channel in the supermarket FinishedShoppingListener() by sending the customer id to it
	readyForCheckoutChan <- c.id
}

// Get num of products of a customer
func (c *Customer) getNumProducts() int {
	return len(c.trolley.products)
}

// Const variables for checkouts, customer per checkout and trolleys
const (
	NumCheckouts            = 6
	NumSmallCheckouts       = 2
	MaxCustomersPerCheckout = 6
	NumTrolleys             = 500
)

// TrolleySizes is a global array for the 3 different trolley sizes, small, medium and large
var TrolleySizes = [...]int{10, 100, 200}

// Enum for channel switch (iota = 0, 1, 2, 3, 4)
const (
	CustomerNew = iota
	CustomerCheckout
	CustomerFinished
	CustomerLost
	CustomerBan
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

// Manager struct
type Manager struct {
	id          int
	supermarket *Supermarket
	wg          *sync.WaitGroup
	//name string
}

// Manager Constructor
// Returns new Manager
func newManager(id int, wg *sync.WaitGroup, pr int64, cr float64, ps float64) *Manager {
	var weather Weather
	weather.initializeWeather()
	weather.generateWeather()
	forecast, multiplier := weather.getWeather()
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
func (m *Manager) customerStatusChangeListener() {
	for {
		input := <-customerStatusChan

		switch input {
		case CustomerNew:
			numberOfCurrentCustomersShopping++
			totalNumberOfCustomersInStore++
			totalNumberOfCustomersToday++

		case CustomerCheckout:
			numberOfCurrentCustomersShopping--
			numberOfCurrentCustomersAtCheckout++

		case CustomerFinished:
			numberOfCurrentCustomersAtCheckout--
			totalNumberOfCustomersInStore--

		case CustomerLost:
			numCustomersLost++
			numberOfCurrentCustomersShopping--
			totalNumberOfCustomersInStore--

		case CustomerBan:
			numCustomersBanned++
			numberOfCurrentCustomersShopping--
			totalNumberOfCustomersInStore--
		default:
			fmt.Println("UH-OH: THINGS JUST GOT SPICY. 🌶🌶🌶")
		}
	}
}

// Returns total number of Checkout objects in the Supermarket
func getNumCheckouts() int {
	return NumCheckouts + NumSmallCheckouts
}

// Listener for checkout open, close
func (m *Manager) openCloseCheckoutListener() {
	for {
		numberOfCheckoutsOpen += <-checkoutChangeStatusChan

		// Check is Supermarket is closing and no customer
		if !m.supermarket.openStatus && totalNumberOfCustomersInStore == 0 {
			break
		}
	}
}

// Returns the Manager Supermarket object
func (m *Manager) getSupermarket() *Supermarket {
	return m.supermarket
}

// Opens the Supermarket and start go routines for channels and printing the updated stats
func (m *Manager) openSupermarket() {
	// Create a Supermarket
	m.supermarket = newSupermarket()

	go m.customerStatusChangeListener()
	go m.openCloseCheckoutListener()

	go m.statPrint()
}

// Prints the current stats of the Supermarket using carriage return
func (m *Manager) statPrint() {
	for {
		fmt.Printf("Total Customers Today: %03d, Total Customers In Store: %03d, Total Customers Shopping: %02d,"+
			" Total Customers At Checkout: %02d, Checkouts Open: %d, Checkouts Closed: %d, Available Trolleys: %03d"+
			", Customers Lost: %02d, Customers Banned: %d\r",
			totalNumberOfCustomersToday, totalNumberOfCustomersInStore, numberOfCurrentCustomersShopping,
			numberOfCurrentCustomersAtCheckout, numberOfCheckoutsOpen, NumCheckouts+NumSmallCheckouts-numberOfCheckoutsOpen,
			NumTrolleys-totalNumberOfCustomersInStore, numCustomersLost, numCustomersBanned)
		time.Sleep(time.Millisecond * 40)

		if !m.supermarket.openStatus && totalNumberOfCustomersInStore == 0 {
			fmt.Printf("\n")
			break
		}
	}

	m.wg.Done()
}

// Returns the total number of Customers in the Supermarket today
func getTotalNumberOfCustomersToday() int {
	return totalNumberOfCustomersToday
}

// Gets the average customer wait time and process time
func getCustomerTimesInSeconds() (string, string) {
	avgWait := float64(customerWaitTimeTotal) / float64(totalNumberOfCustomersToday-numCustomersLost)
	avgProcess := float64(customerProcessTimeTotal) / float64(totalNumberOfCustomersToday-numCustomersLost)

	avgWait /= float64(time.Second)
	avgProcess /= float64(time.Second)

	sWait := fmt.Sprintf("%dm %ds", int(avgWait)/60, int(avgWait)%60)
	sProcess := fmt.Sprintf("%dm %ds", int(avgProcess)/60, int(avgProcess)%60)
	return sWait, sProcess
}

// Closes the supermarket
func (m *Manager) closeSupermarket() {
	m.supermarket.openStatus = false
}

// Product struct holds processing time for that product
type Product struct {
	time float64
}

// Product Constructor
func newProduct() *Product {
	p := Product{rand.Float64() * processSpeed}
	return &p
}

// returns the products processing time
func (p *Product) getTime() float64 {
	return p.time
}

var trolleyMutex *sync.Mutex
var customerMutex *sync.RWMutex
var checkoutMutex *sync.RWMutex

// Supermarket struct
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
// Returns new Supermarket
func newSupermarket() *Supermarket {
	trolleyMutex = &sync.Mutex{}
	customerMutex = &sync.RWMutex{}
	checkoutMutex = &sync.RWMutex{}

	s := Supermarket{0, true, make([]*Checkout, 0, 256), make([]*Checkout, 0, 256), make(map[int]*Customer), make([]*Trolley, NumTrolleys), make(chan int), make(chan int)}
	s.generateTrolleys()
	s.generateCheckouts()

	go s.generateCustomer()
	go s.finishedShoppingListener()
	go s.finishedCheckoutListener()

	return &s
}

// Create a customer and adds them to to the customers map in supermarket
func (s *Supermarket) generateCustomer() {
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
		trolleySize := TrolleySizes[rand.Intn(len(TrolleySizes))]

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
		customerStatusChan <- CustomerNew

		// Add customer to the customers map in supermarket, key=customer.id, value=customer
		customerMutex.Lock()
		s.customers[c.id] = c
		customerMutex.Unlock()

		// Customer can now go add products to the trolley
		go c.shop(s.finishedShopping)

		// Decides to open or close checkouts
		s.calculateOpenCheckout()
	}
}

// Sends customer to a checkout
func (s *Supermarket) sendToCheckout(id int) {
	customerMutex.RLock()
	c := s.customers[id]
	customerMutex.RUnlock()

	// Choose the best checkout for a customer to go to
	checkout, pos := s.chooseCheckout(c.getNumProducts(), c.impatient)
	// No checkout with < max number in queue - The number of lost customers (Customers will leave the store if they need to join a queue more than six deep)
	if pos < 0 {
		s.customerLeavesStore(id)
		customerStatusChan <- CustomerLost
		return
	}

	// Checks if customer is impatient and joins a ten or less checkout with more tha 10 items
	// Manager has a 50% chance of finding them and banning them
	if c.impatient && c.getNumProducts() > 10 && checkout.tenOrLess && rand.Float64() < 0.5 {
		s.customerLeavesStore(id)
		customerStatusChan <- CustomerBan
		return
	}

	checkout.addPersonToLine(c)

	// Change the status channel of customer, sends a 1
	customerStatusChan <- CustomerCheckout
}

// Gets the best open checkout for a customer to go to at the current time
func (s *Supermarket) chooseCheckout(numProducts int, isImpatient bool) (*Checkout, int) {
	min, pos := -1, -1

	checkoutMutex.RLock()
	for i := 0; i < len(s.checkoutOpen); i++ {
		// Gets the number of people in the checkout and if the checkout is 'tenOrLess'
		// Checks if the customer can join the checkout (less than max number (6) allowed)
		// Ensure only customers with 10 or less items can go to the 10 or less checkouts
		// Added impatience variable
		// Finds the checkout with the least amount of people
		if num, tenOrLess := s.checkoutOpen[i].getNumPeopleInLine(), s.checkoutOpen[i].tenOrLess; ((tenOrLess && (numProducts <= 10 || isImpatient)) || !tenOrLess) && (num < min || min < 0) && num < MaxCustomersPerCheckout {
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
func (s *Supermarket) generateTrolleys() {
	for i := 0; i < NumTrolleys; i++ {
		s.trolleys[i] = newTrolley(TrolleySizes[rand.Intn(len(TrolleySizes))])
	}
}

// Generates 8 checkouts
func (s *Supermarket) generateCheckouts() {
	var hasScanner bool
	// Default create 8 Checkouts when Supermarket is created
	for i := 0; i < NumCheckouts+NumSmallCheckouts-1; i++ {
		hasScanner := rand.Float64() < 0.5
		if i == 0 {
			s.checkoutOpen = append(s.checkoutOpen, newCheckout(i+1, false, false, false, hasScanner, false, 0, false, make(chan *Customer, MaxCustomersPerCheckout), 0, 0, 0, 0, true, s.finishedCheckout))
		} else {
			s.checkoutClosed = append(s.checkoutClosed, newCheckout(i+1, i >= NumCheckouts, false, false, hasScanner, false, 0, false, make(chan *Customer, MaxCustomersPerCheckout), 0, 0, 0, 0, false, s.finishedCheckout))
		}
	}

	s.checkoutClosed = append(s.checkoutClosed, newCheckout(NumCheckouts+NumSmallCheckouts, true, true, false, hasScanner, false, 0, false, make(chan *Customer, MaxCustomersPerCheckout), 0, 0, 0, 0, false, s.finishedCheckout))
}

// Waits for a customer to finish shopping using a channel, then sends the customer to a checkout
func (s *Supermarket) finishedShoppingListener() {
	for {
		if !s.openStatus && numberOfCurrentCustomersShopping == 0 {
			break
		}

		// Check if customer is finished adding products to trolley using channel from the shop() method in Customer.go
		id := <-s.finishedShopping

		// Send customer to a checkout
		s.sendToCheckout(id)
	}
}

// Waits for a customer to finish at a checkout using a channel, then removes the customer from the supermarket
func (s *Supermarket) finishedCheckoutListener() {
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
		s.customerLeavesStore(id)

		customerStatusChan <- CustomerFinished

		s.calculateOpenCheckout()
	}
}

// Cleans up customer and trolley items when they leave a shop
func (s *Supermarket) customerLeavesStore(id int) {
	customerMutex.RLock()
	trolley := s.customers[id].trolley
	customerMutex.RUnlock()

	trolley.emptyTrolley()
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
func (s *Supermarket) calculateOpenCheckout() {
	numOfCurrentCustomers := len(s.customers)
	numOfOpenCheckouts := len(s.checkoutOpen)
	calculationOfThreshold := int(math.Ceil(float64(numOfCurrentCustomers) / MaxCustomersPerCheckout))

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
		s.checkoutClosed[0].open()
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
		checkout, pos := s.chooseCheckout(0, false)
		if pos < 0 {
			return
		}
		checkout.close()
		s.checkoutClosed = append(s.checkoutClosed, checkout)
		s.checkoutOpen = append(s.checkoutOpen[0:pos], s.checkoutOpen[pos+1:]...)

		checkoutChangeStatusChan <- -1

		//fmt.Printf("1 checkout just closed. We now have %d open checkouts.\n", len(s.checkoutOpen))

		return
	}
}

// returns a slice of all of the checkouts in the supermarket
func (s *Supermarket) getAllCheckouts() []*Checkout {
	return append(s.checkoutOpen, s.checkoutClosed...)
}

// Trolley struct for holding products
type Trolley struct {
	capacity int
	products []*Product
}

// Trolley Constructor
func newTrolley(capacity int) *Trolley {
	t := Trolley{capacity, make([]*Product, 0, capacity)}
	return &t
}

// Adds a product to a trolley
func (t *Trolley) addProductToTrolley(product *Product) {
	t.products = append(t.products, product)
}

// Checks if trolley has reached capacity
func (t *Trolley) isFull() bool {
	return t.capacity == len(t.products)
}

// Empties trolley by declaring the current slice as a new slice
func (t *Trolley) emptyTrolley() {
	t.products = make([]*Product, 0, t.capacity)
}

// Weather struct which affects customer generation
type Weather struct {
	status    int
	forecasts [4]string
}

// Initializes the forecast array of string to
// include 4 different weather types.
func (w *Weather) initializeWeather() {
	w.forecasts[0] = "SUNNY DAYS" //1.25
	w.forecasts[1] = "RAINY DAYS" // .75
	w.forecasts[2] = "CLEAR DAY"  // 1
	w.forecasts[3] = "SNOWY DAY"  //.5
}

// Returns a string of the current weather forecast i.e. "SUNNY DAYS"
// and also returns a float64 value which is used as a multiplyer for
// customers entering the shop
func (w *Weather) getWeather() (string, float64) {
	//return w.forecasts[w.status]
	multipliers := [4]float64{1.25, 0.75, 1, 0.5}
	return w.forecasts[w.status], multipliers[w.status]
}

// Sets the weathers status equal to forecastIndex
func (w *Weather) changeWeather(forecastIndex int) {
	w.status = forecastIndex
}

// Generates a random number between 0-3 and sets the weather
// status to the random number.
func (w *Weather) generateWeather() {
	rand.Seed(time.Now().UnixNano())
	weatherRand := rand.Intn(4)
	w.changeWeather(weatherRand)
}
