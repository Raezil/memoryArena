# memoryArena
To install 
```
go get github.com/Raezil/memoryArena@latest
```

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
	arena := NewArena(512)
	obj, _ := NewObject[Person](arena, Person{"Kamil", 26})
	defer arena.Reset()
	fmt.Println(obj)
}

```
