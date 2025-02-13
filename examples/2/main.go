package main

import (
	"fmt"

	. "github.com/Raezil/memoryArena"
)

type Person struct {
	Name string
	Age  int
}

func main() {
	arena, err := NewMemoryArena[[]Person](1024)
	if err != nil {
		panic(err)
	}

	// Allocate memory for the object
	ptr, err := arena.AllocateObject([]Person{
		{"Alice", 25},
		{"Bob", 30},
		{"Charlie", 35},
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get the value from the pointer
	people := *(*[]Person)(ptr)
	fmt.Println(people)
	AppendSlice(&Person{"David", 40}, arena, &people)

	// Reset the arena
	defer arena.Reset()
	fmt.Println(people)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
}
