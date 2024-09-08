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
	arena := NewMemoryArena[Person](512)
	obj, _ := NewObject[Person](arena, Person{"Kamil", 26})
	defer arena.Reset()
	fmt.Println(obj)
}

```

To install 
```
go get github.com/Raezil/memoryArena@latest
```
