package main

import "fmt"

func inserisciOrdinato(slice []int, elemento int) []int {
	// Trova l'indice giusto in cui inserire l'elemento
	i := 0
	for i < len(slice) && slice[i] < elemento {
		i++
	}

	if i < len(slice) && slice[i] == elemento {
		return slice
	}

	// Inserisci l'elemento nella posizione trovata
	slice = append(slice[:i], append([]int{elemento}, slice[i:]...)...)
	return slice
}

func main() {
	// Slice iniziale
	slice := []int{1, 3, 5, 7, 9}

	// Aggiungi nuovi elementi mantenendo l'ordine crescente
	slice = inserisciOrdinato(slice, 4)
	slice = inserisciOrdinato(slice, 1)
	slice = inserisciOrdinato(slice, 10)

	fmt.Println(slice) // Output: [0 1 3 4 5 7 9 10]
}
