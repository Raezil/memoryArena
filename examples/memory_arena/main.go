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
	arena, err := NewMemoryArena[Person](48)
	if err != nil {
		panic(err)
	}

	// Allocate memory for the object
	ptr, err := arena.AllocateObject(Person{Name: "John", Age: 30})
	if err != nil {
		panic(err)
	}

	ptr2, err := arena.AllocateObject(Person{Name: "Kamil", Age: 27})
	if err != nil {
		panic(err)
	}

	// Get the object from the pointer
	person1 := *(*Person)(ptr)
	person2 := *(*Person)(ptr2)

	fmt.Printf("Person: %+v\n", person1)
	fmt.Printf("Person: %+v\n", person2)

}
