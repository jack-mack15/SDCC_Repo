package main

import "fmt"

var stringa string

func main() {

	var lista1 []int
	var lista2 []int

	lista1 = append(lista1, 1)
	lista2 = lista1

	fmt.Println(len(lista1))
	fmt.Println(len(lista2))

	lista1 = append(lista1, 2)
	fmt.Println(len(lista1))
	fmt.Println(len(lista2))
}
