package main
import "github.com/jyorien/goch8/chip8"

func main() {
	ch8 := chip8.NewChip8()
	ch8.LoadROM()
}