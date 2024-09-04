package main

import "fmt"

var stringa string

func main() {

	var list []int
	list = append(list, 1)
	list = append(list, 2)
	list = append(list, 3)
	list = append(list, 4)
	list = append(list, 12)

	fmt.Printf("list:%v\n\n", list)

}
