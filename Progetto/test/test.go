package main

import "fmt"

var stringa string

func remove(slice []int, index int) []int {
	// Verifica che l'indice sia all'interno del range della slice
	if index < 0 || index >= len(slice) {
		fmt.Println("Indice fuori dal range")
		return slice
	}

	// Rimuovi l'elemento dall'indice specificato
	return append(slice[:index], slice[index+1:]...)
}

func main() {

	numbers := []int{1}

	// Stampa la slice originale
	fmt.Println("Slice originale:", numbers)

	// Rimuove l'elemento all'indice 2 (il numero 3 in questo caso)
	numbers = remove(numbers, 0)

	// Stampa la slice risultante dopo la rimozione
	fmt.Println("Slice dopo rimozione:", numbers)

}
