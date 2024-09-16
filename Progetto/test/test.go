package main

import "fmt"

// Definizione della struct che contiene una slice di interi
type MyStruct struct {
	numbers []int
}

// Funzione per rimuovere gli elementi che corrispondono a un certo valore
func removeElements(value int, ms *MyStruct) {
	for i := 0; i < len(ms.numbers); {
		if ms.numbers[i] == value {
			// Rimuovi l'elemento corrente usando append
			ms.numbers = append(ms.numbers[:i], ms.numbers[i+1:]...)
		} else {
			// Passa all'elemento successivo solo se non è stato rimosso
			i++
		}
	}
}

func main() {

	// Inizializza la struct con una slice di interi
	ms := &MyStruct{numbers: []int{1, 2, 3, 4, 3, 5, 3, 6}}

	fmt.Println("Prima della rimozione:", ms.numbers)

	// Rimuovi gli elementi che corrispondono al valore 3
	removeElements(3, ms)

	fmt.Println("Dopo la rimozione:", ms.numbers)

}
