package main

import (
	"fmt"

	"github.com/jyorien/goch8/chip8"
)

func main() {
	ch8 := chip8.NewChip8()
	ch8.LoadROM()
	for {
		op := ch8.Fetch()
		fmt.Printf("op: %x\n", op)
		ch8.Execute(op)

	}
}
