//go:generate go run github.com/strangedev/hookah cmd/main.go github.com/strangedev/hookah/test Person > generated.go
package test

type Person struct {
	name string
	age  int
}

func (p Person) Name() string {
	return p.name
}

func (p Person) Age() int {
	return p.age
}

func (p *Person) SetName(name string) error {
	p.name = name

	return nil
}
