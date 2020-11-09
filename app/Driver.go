package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"packageService"
	"strconv"
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
	rand.Seed(time.Now().UnixNano())
	trolley := new(packageService.Trolley)
	trolley.SetTrolleyCapacity(1)
	apple := new(packageService.Product)
	apple.SetTime(2)
	trolley.InitalizeProducts()
	trolley.AddProductToTrolley(*apple, 0)
	for i := 0; i < trolley.GetTrolleyCapacity(); i++ {
		productInTrolley := trolley.GetProduct(i)
		fmt.Println(productInTrolley.GetTime())
	}

	productsRate, _ := strconv.ParseInt(userInput("Please enter the range of products per trolley. (1-200):", 1, 200, true), 10, 64)
	customerRate, _ := strconv.ParseInt(userInput("Please enter the rate customers arrive at checkouts. (0-60):", 0, 60, true), 10, 64)
	processSpeed, _ := strconv.ParseFloat(userInput("Please enter the range for product processing speed. (0.5-6):", 0, 60, false), 64)
	fmt.Println("Products rate:", productsRate)
	fmt.Println("Customer rate:", customerRate)
	fmt.Printf("%s %f", "Process Speed:", processSpeed)
}
