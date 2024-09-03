package main

import "fmt"

var stringa string

func main() {

	type myStruct struct {
		a int
		b string
	}

	var lista []myStruct

	currStruct1 := myStruct{a: 1, b: "bho"}
	currStruct2 := myStruct{a: 2, b: "ciao"}
	currStruct3 := myStruct{a: 3, b: "dest"}
	lista = append(lista, currStruct1, currStruct2, currStruct3)
	fmt.Printf("len prima is: %d \n", len(lista))

	pointer := &lista[0]
	fmt.Printf("a is: %d \n", pointer.a)

	lista = append(lista[:1], lista[2:]...)
	fmt.Printf("len dopo is: %d \n", len(lista))
	fmt.Printf("a del pointer post delete is: %d \n", pointer.a)

}
