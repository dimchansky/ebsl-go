package main

import (
	"fmt"

	"github.com/dimchansky/ebsl-go/evidence"
	"github.com/dimchansky/ebsl-go/opinion"
)

func main() {
	c := uint64(2)
	o1 := opinion.FromEvidence(c, evidence.New(5, 2))
	o2 := opinion.FromEvidence(c, evidence.New(10, 1))

	fmt.Printf("o1: %v\n", o1)
	fmt.Printf("o1: %v\n", o2)
	fmt.Printf("o1⊕o2: %v\n", o1.Plus(o2).ToEvidence(c))
	fmt.Printf("2·(o1⊕o2): %v\n", o1.Mul(2).ToEvidence(c))
}
