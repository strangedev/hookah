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
