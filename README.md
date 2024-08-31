# memoryArena
To install 
```
go get github.com/Raezil/memoryArena@latest
```

Example
```
type Person struct {
	Name string
	Age  int
}

// NewPerson creates a new Person object in the arena
func NewPerson(arena *MemoryArena, name string, age int) (*Person, error) {
	person := Person{Name: name, Age: age}
	ptr, err := arena.AllocateObject(person)
	if err != nil {
		return nil, err
	}
	return (*Person)(ptr), nil
}

func main() {
	arena := NewArena(512)
	person, _ := NewPerson(arena, "Kamil", 26)
	defer arena.Reset()
	fmt.Println(person)
}
```
