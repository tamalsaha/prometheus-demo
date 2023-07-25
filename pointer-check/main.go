// You can edit this code!
// Click here and start typing.
package main

import "fmt"

func hello(x *string) {
	println(*x)

	y := "hi"

	x = &y
}

func main2() {
	a := "hello"
	hello(&a)

	println(a)
}

type Stuff struct {
	Name string
}

func (s Stuff) GetName() string {
	return s.Name
}

type Object interface {
	GetName() string
}

func main() {
	arr := []string{"a", "b"}
	arr = append(arr, "c", "d")

	a := Stuff{Name: "a"}

	out := doStuff(&a)
	fmt.Println(a.Name, out.GetName())
}

func doStuff(in Object) Object {
	b := Stuff{Name: "b"}
	in = &b
	return b
}
