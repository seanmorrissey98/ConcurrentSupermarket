package main

import (
	"fmt"
)

func main() {
	trolley := new(Trolley)
	trolley.setTrolleyCapacity(1)
	apple := new(Product)
	apple.setTime(2)
	trolley.initalizeProducts()
	trolley.addProductToTrolley(*apple, 0)
	for i := 0; i < trolley.getTrolleyCapacity(); i++ {
		productInTrolley := trolley.getProduct(i)
		fmt.Println(productInTrolley.getTime())
	}
}
