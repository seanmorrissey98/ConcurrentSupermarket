package main

import (
	"ConcurrentSupermarket/packageService"
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"runtime/trace"
	"sort"
	"strconv"
	"sync"
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
	m := packageService.NewManager(1, &wg, productsRate, float64(customerRate), processSpeed)
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

func PrintCheckoutStats(checkouts []*packageService.Checkout, totalProcessedCustomers int64, totalProcessedProducts int64) {
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

	total := packageService.GetTotalNumberOfCustomersToday()
	fmt.Printf("Average Products Per Trolley: %.2f\n\n", float64(totalProcessedProducts)/float64(total))

	avgWait, avgProcess := packageService.GetCustomerTimesInSeconds()
	fmt.Printf("Average Customer Wait Time: %s, \nAverage Customer Process Time: %s\n", avgWait, avgProcess)

	fmt.Printf("Average Checkout Utilisation: %.2f%s\n", totalUtilization/float64(packageService.GetNumCheckouts()), "%")
}

func getTotalProcessedProducts(c []*packageService.Checkout) int64 {
	var total int64
	total = 0
	for i := range c {
		total += c[i].ProcessedProducts
	}
	return total
}

func getTotalProcessedCustomers(c []*packageService.Checkout) int64 {
	var total int64
	total = 0
	for i := range c {
		total += c[i].ProcessedCustomers
	}
	return total
}
