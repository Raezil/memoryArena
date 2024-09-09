![Screenshot_2024-09-09_at_20-14-40_ChatGPT-removebg-preview](https://github.com/user-attachments/assets/c7f6f25b-e0ce-4159-be8e-7865c6e63236)

# Memory Arena lib in Golang!
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
	if err != nil {
		fmt.Printf(err.Error())
	}
	obj, _ := NewObject[[]Person](concurrentArena, []Person{Person{"Kamil", 27}, Person{"Lukasz", 28}})
	defer Reset(arena)
	fmt.Println(obj)

}
```

To install 
```
go get github.com/Raezil/memoryArena@latest
```
