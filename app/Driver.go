package main

import (
	"fmt"
	"packageService"
)

func main() {
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
}
