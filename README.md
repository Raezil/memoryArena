<p align="center">
  <img src="https://github.com/user-attachments/assets/c7f6f25b-e0ce-4159-be8e-7865c6e63236">
</p>


<h1 align="center">Memory Arena lib in Golang!</h1>
The purpose of this package is to isolate multiple related allocations into a single area of memory, so that they can be freed all at once.



Example
```
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
```

```
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
```

```
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
```

To install 
```
go get github.com/Raezil/memoryArena@latest
```

If you wish to test it, run commnad: go test
