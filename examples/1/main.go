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
	arena, err := NewConcurrentArena[[]Person](100)
	if err != nil {
		fmt.Printf(err.Error())
	}
	obj, _ := NewObject[[]Person](arena, []Person{Person{"Kamil", 27}, Person{"Lukasz", 28}})
	defer arena.Reset()
	fmt.Println(obj)

}
