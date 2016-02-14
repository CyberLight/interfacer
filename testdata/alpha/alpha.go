package alpha

import (
	"fmt"
)

type A struct{
	a int
}

type B struct {
	A
	b int
	c int
}

func (b *A) aaaa(av int) (bool, error) {
	fmt.Println("A:aaaa")
	return true, nil
}

func (b *B) ffff(av int, a *A) (*A, func(int, int) bool, error) {
	fmt.Println("B:ffff")
	return &A{}, func(a, b int) bool { return true }, nil
}
