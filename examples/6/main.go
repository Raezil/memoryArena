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
	arena, err := NewConcurrentArena[Person](72)
	if err != nil {
		fmt.Println(err)
	}
	person, err := NewObject(arena, Person{"John", 30})
	if err != nil {
		fmt.Println(err)
	}
	person1, err := NewObject(arena, Person{"Kamil", 27})
	if err != nil {
		fmt.Println(err)
	}

	person2, err := NewObject(arena, Person{"Lukasz", 28})
	if err != nil {
		fmt.Println(err)
	}
	arena.Reset()

	fmt.Println(*person)
	fmt.Println(*person1)
	fmt.Println(*person2)
}
