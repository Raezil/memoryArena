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
	arena, err := NewMemoryArena[Person](72)
	concurentArena := NewConcurrentArena(*arena)
	if err != nil {
		fmt.Println(err)
	}
	person, err := NewObject(concurentArena, Person{"John", 30})
	if err != nil {
		fmt.Println(err)
	}
	person1, err := NewObject(concurentArena, Person{"Kamil", 27})
	if err != nil {
		fmt.Println(err)
	}

	person2, err := NewObject(concurentArena, Person{"Lukasz", 28})
	if err != nil {
		fmt.Println(err)
	}
	arena.Reset()

	fmt.Println(*person)
	fmt.Println(*person1)
	fmt.Println(*person2)
}
