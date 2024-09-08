# memoryArena
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
