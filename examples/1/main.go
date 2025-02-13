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
	arena, err := NewMemoryArena[[]Person](512)
	if err != nil {
		fmt.Printf(err.Error())
	}
	concurrentArena := NewConcurrentArena[[]Person](*arena)
	obj, _ := NewObject[[]Person](concurrentArena, []Person{Person{"Kamil", 27}, Person{"Lukasz", 28}})
	defer Reset(arena)
	fmt.Println(obj)

}
