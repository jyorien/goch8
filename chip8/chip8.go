package chip8

import (
	"log"
	"math/rand/v2"
	"os"
)

const START_ADDRESS = 0x200
const FONT_SET_START_ADDRESS = 0x50
const VIDEO_WIDTH = 64
const VIDEO_HEIGHT = 32

var FONT_SET = []uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0,
	0x20, 0x60, 0x20, 0x20, 0x70,
	0xF0, 0x10, 0xF0, 0x80, 0xF0,
	0xF0, 0x10, 0xF0, 0x10, 0xF0,
	0x90, 0x90, 0xF0, 0x10, 0x10,
	0xF0, 0x80, 0xF0, 0x10, 0xF0,
	0xF0, 0x80, 0xF0, 0x90, 0xF0,
	0xF0, 0x10, 0x20, 0x40, 0x40,
	0xF0, 0x90, 0xF0, 0x90, 0xF0,
	0xF0, 0x90, 0xF0, 0x10, 0xF0,
	0xF0, 0x90, 0xF0, 0x90, 0x90,
	0xE0, 0x90, 0xE0, 0x90, 0xE0,
	0xF0, 0x80, 0x80, 0x80, 0xF0,
	0xE0, 0x90, 0x90, 0x90, 0xE0,
	0xF0, 0x80, 0xF0, 0x80, 0xF0,
	0xF0, 0x80, 0xF0, 0x80, 0x80}

type Chip8 struct {
	registers     [16]uint8
	indexRegister uint16
	pc            uint16
	memory        [4096]uint8
	stack         [16]uint16
	sp            uint8
	delayTimer    uint8
	soundTimer    uint8
	keypad        [16]uint8
	video         [64 * 32]uint32
	opcode        uint16
}

func (ch8 *Chip8) LoadROM() {
	data, err := os.ReadFile("../roms/ibm_logo.ch8")
	if err != nil {
		log.Fatal(err)
	}
	copy(ch8.memory[START_ADDRESS:], data)
}

func NewChip8() *Chip8 {
	ch8 := Chip8{pc: START_ADDRESS}
	copy(ch8.memory[FONT_SET_START_ADDRESS:], FONT_SET)
	return &ch8
}

func (ch8 *Chip8) Fetch() uint16 {
	op := uint16(ch8.memory[ch8.pc])<<8 | uint16(ch8.memory[ch8.pc+1])
	ch8.pc += 2
	return op

}

// TOOD: broken, need to bitmask
func (ch8 *Chip8) Execute(opcode uint16) {
	ch8.opcode = opcode

	switch opcode & 0xF000 { // First nibble
	case 0x0000:
		switch opcode { // Exact match for 0x commands
		case 0x00E0: // Clear screen
			ch8.OP_00E0()
		case 0x00EE: // Return from subroutine
			ch8.OP_00EE()
		}
	case 0x1000: // 1nnn - Jump to nnn
		ch8.OP_1nnn()
	case 0x2000: // 2nnn - Call subroutine at nnn
		ch8.OP_2nnn()
	case 0x3000: // 3xkk - Skip if Vx == kk
		ch8.OP_3xkk()
	case 0x4000: // 4xkk - Skip if Vx != kk
		ch8.OP_4xkk()
	case 0x5000: // 5xy0 - Skip if Vx == Vy
		ch8.OP_5xy0()
	case 0x6000: // 6xkk - Set Vx = kk
		ch8.OP_6xkk()
	case 0x7000: // 7xkk - Add kk to Vx
		ch8.OP_7xkk()
	case 0x8000: // 8xyz - Multiple arithmetic operations
		switch opcode & 0x000F { // Look at last nibble
		case 0x0: // 8xy0
			ch8.OP_8xy0()
		case 0x1: // 8xy1
			ch8.OP_8xy1()
		case 0x2: // 8xy2
			ch8.OP_8xy2()
		case 0x3: // 8xy3
			ch8.OP_8xy3()
		case 0x4: // 8xy4
			ch8.OP_8xy4()
		case 0x5: // 8xy5
			ch8.OP_8xy5()
		case 0x6: // 8xy6
			ch8.OP_8xy6()
		case 0x7: // 8xy7
			ch8.OP_8xy7()
		case 0xE: // 8xyE
			ch8.OP_8xyE()
		}
	case 0x9000: // 9xy0 - Skip if Vx != Vy
		ch8.OP_9xy0()
	case 0xA000: // Annn - Set I = nnn
		ch8.OP_Annn()
	case 0xB000: // Bnnn - Jump to nnn + V0
		ch8.OP_Bnnn()
	case 0xC000: // Cxkk - Set Vx = random & kk
		ch8.OP_Cxkk()
	case 0xD000: // Dxyn - Draw sprite
		ch8.OP_Dxyn()
	case 0xE000: // Ex__ - Keyboard operations
		switch opcode & 0x00FF { // Look at last two nibbles
		case 0x9E: // Ex9E
			ch8.OP_Ex9E()
		case 0xA1: // ExA1
			ch8.OP_ExA1()
		}
	case 0xF000: // Fx__ - Timer/memory operations
		switch opcode & 0x00FF { // Look at last two nibbles
		case 0x07: // Fx07
			ch8.OP_Fx07()
		case 0x0A: // Fx0A
			ch8.OP_Fx0A()
		case 0x15: // Fx15
			ch8.OP_Fx15()
		case 0x18: // Fx18
			ch8.OP_Fx18()
		case 0x1E: // Fx1E
			ch8.OP_Fx1E()
		case 0x29: // Fx29
			ch8.OP_Fx29()
		case 0x33: // Fx33
			ch8.OP_Fx33()
		case 0x55: // Fx55
			ch8.OP_Fx55()
		case 0x65: // Fx65
			ch8.OP_Fx65()
		}
	}
}

// Clear Display
func (ch8 *Chip8) OP_00E0() {
	for i := 0; i < len(ch8.video); i++ {
		ch8.video[i] = 0
	}
}

// Return from Subroutine
func (ch8 *Chip8) OP_00EE() {
	ch8.sp--
	ch8.pc = ch8.stack[ch8.sp]
}

// Jump to nnn
func (ch8 *Chip8) OP_1nnn() {
	ch8.pc = ch8.opcode & 0x0FFF
}

// / Call Subroutine at nnn
func (ch8 *Chip8) OP_2nnn() {
	ch8.stack[ch8.sp] = ch8.pc
	ch8.sp++
	ch8.pc = ch8.opcode & 0x0FFF
}

// Skip instruction if register x == value kk
func (ch8 *Chip8) OP_3xkk() {
	// shift by 8 to push to last byte
	Vx := (ch8.opcode & 0x0F00) >> 8
	kk := ch8.opcode & 0x00FF

	if uint16(ch8.registers[Vx]) == kk {
		ch8.pc += 2
	}
}

// Skip instruction if register x != value kk
func (ch8 *Chip8) OP_4xkk() {
	// shift by 8 to push to last byte
	Vx := (ch8.opcode & 0x0F00) >> 8
	kk := ch8.opcode & 0x00FF

	if uint16(ch8.registers[Vx]) != kk {
		ch8.pc += 2
	}
}

// Skip instruction if register x == register y
func (ch8 *Chip8) OP_5xy0() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	Vy := (ch8.opcode & 0x00F0) >> 4
	if ch8.registers[Vx] == ch8.registers[Vy] {
		ch8.pc += 2
	}
}

// Set register x = kk
func (ch8 *Chip8) OP_6xkk() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	kk := ch8.opcode & 0x00FF
	ch8.registers[Vx] = uint8(kk)
}

// Set register x = register x + kk
func (ch8 *Chip8) OP_7xkk() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	kk := ch8.opcode & 0x00FF
	ch8.registers[Vx] = ch8.registers[Vx] + uint8(kk)
}

// Set register x = register y
func (ch8 *Chip8) OP_8xy0() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	Vy := (ch8.opcode & 0x00F0) >> 4
	ch8.registers[Vx] = ch8.registers[Vy]
}

// Set register x = register x | register y
func (ch8 *Chip8) OP_8xy1() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	Vy := (ch8.opcode & 0x00F0) >> 4
	ch8.registers[Vx] = ch8.registers[Vx] | ch8.registers[Vy]

}

// Set register x = register x & register y
func (ch8 *Chip8) OP_8xy2() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	Vy := (ch8.opcode & 0x00F0) >> 4
	ch8.registers[Vx] = ch8.registers[Vx] & ch8.registers[Vy]

}

// Set register x = register x ^ register y
func (ch8 *Chip8) OP_8xy3() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	Vy := (ch8.opcode & 0x00F0) >> 4
	ch8.registers[Vx] = ch8.registers[Vx] ^ ch8.registers[Vy]
}

// Set register x = register x + register y, register f = carry
func (ch8 *Chip8) OP_8xy4() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	Vy := (ch8.opcode & 0x00F0) >> 4
	res := ch8.registers[Vx] + ch8.registers[Vy]
	if res > 0xFF {
		ch8.registers[0xF] = 1
	} else {
		ch8.registers[0xF] = 0
	}
	ch8.registers[Vx] = res & 0xFF

}

// Set register x = register x - register y, register f = Vx > Vy
func (ch8 *Chip8) OP_8xy5() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	Vy := (ch8.opcode & 0x00F0) >> 4
	if ch8.registers[Vx] > ch8.registers[Vy] {
		ch8.registers[0xF] = 1
	} else {
		ch8.registers[0xF] = 0
	}
	ch8.registers[Vx] = ch8.registers[Vx] - ch8.registers[Vy]
}

// Set register x = register x >> 1 (Shift Right)
func (ch8 *Chip8) OP_8xy6() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	ch8.registers[0xF] = ch8.registers[Vx] & 0x1
	ch8.registers[Vx] >>= 1
}

// Reverse subtract
func (ch8 *Chip8) OP_8xy7() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	Vy := (ch8.opcode & 0x00F0) >> 4
	if ch8.registers[Vy] > ch8.registers[Vx] {
		ch8.registers[0xF] = 1
	} else {
		ch8.registers[0xF] = 0
	}
	ch8.registers[Vx] = ch8.registers[Vy] - ch8.registers[Vx]
}

// Set register x = register x << 1 (Shift Left)
func (ch8 *Chip8) OP_8xyE() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	ch8.registers[0xF] = ch8.registers[Vx] >> 7
	ch8.registers[Vx] <<= 1
}

// Skip instruction if register x != register y
func (ch8 *Chip8) OP_9xy0() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	Vy := (ch8.opcode & 0x00F0) >> 4
	if ch8.registers[Vx] != ch8.registers[Vy] {
		ch8.pc += 2
	}
}

// Set index register address to nnnn
func (ch8 *Chip8) OP_Annn() {
	nnn := ch8.opcode & 0x0FFF
	ch8.indexRegister = nnn
}

// Jump to nnn + register 0
func (ch8 *Chip8) OP_Bnnn() {
	nnn := ch8.opcode & 0x0FFF
	ch8.pc = nnn + uint16(ch8.registers[0])
}

// Set register x = random byte & kk
func (ch8 *Chip8) OP_Cxkk() {
	Vx := ch8.opcode & 0x0F00
	kk := ch8.opcode & 0x00FF
	randByte := rand.IntN(256)
	ch8.registers[Vx] = uint8(randByte) & uint8(kk)
}

// Display sprite at location (Vx, Vy)
func (ch8 *Chip8) OP_Dxyn() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	Vy := (ch8.opcode & 0x00F0) >> 4
	n := (ch8.opcode & 0x000F)

	// wrap in case it goes off screen
	xPos := ch8.registers[Vx] % VIDEO_WIDTH
	yPos := ch8.registers[Vy] % VIDEO_HEIGHT

	// no collision
	ch8.registers[0xF] = 0

	for row := 0; uint16(row) < n; row++ {
		spriteByte := ch8.memory[ch8.indexRegister+uint16(row)]
		for col := 0; col < 8; col++ {
			spritePixel := spriteByte & (0x80 >> col)
			screenPixel := &ch8.video[(yPos+uint8(row))*VIDEO_WIDTH+(xPos+uint8(col))]

			if spritePixel != 0 {

				// collision
				if *screenPixel == 0xFFFFFFFF {
					ch8.registers[0xF] = 1
				}
				// xor
				*screenPixel ^= 0xFFFFFFFF

			}
		}
	}

}

// Skip instruction if key Vx pressed
func (ch8 *Chip8) OP_Ex9E() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	key := ch8.registers[Vx]
	if ch8.keypad[key] == 1 {
		ch8.pc += 2
	}
}

// Skip instruction if key Vx not pressed
func (ch8 *Chip8) OP_ExA1() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	key := ch8.registers[Vx]
	if ch8.keypad[key] == 0 {
		ch8.pc += 2
	}
}

// Set register x to delay timer value
func (ch8 *Chip8) OP_Fx07() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	ch8.registers[Vx] = ch8.delayTimer
}

// Wait for key press and store value of key in register x
func (ch8 *Chip8) OP_Fx0A() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	for i, ele := range ch8.registers {
		if ele == 1 {
			ch8.registers[Vx] = uint8(i)
			return
		}
	}
	// if no input, decrement pc and keep waiting
	ch8.pc -= 2
}

// Set delayTimer to register x
func (ch8 *Chip8) OP_Fx15() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	ch8.delayTimer = ch8.registers[Vx]
}

// Set soundTimer to register x
func (ch8 *Chip8) OP_Fx18() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	ch8.soundTimer = ch8.registers[Vx]
}

// Add register x to indexRegister
func (ch8 *Chip8) OP_Fx1E() {
	Vx := (ch8.opcode & 0x0F00) >> 8
	ch8.indexRegister += uint16(ch8.registers[Vx])
}

// Set indexRegister to location of sprite in Vx
func (ch8 *Chip8) OP_Fx29() {
	Vx := (ch8.opcode & 0xF00) >> 8
	ch8.indexRegister = FONT_SET_START_ADDRESS + uint16(5*ch8.registers[Vx])
}

// Store BCD of register x in indexRegister (hundreds), indexRegister+1 (tens), indexRegister+2 (ones)
func (ch8 *Chip8) OP_Fx33() {
	Vx := (ch8.opcode & 0xF00) >> 8
	value := ch8.registers[Vx]
	ch8.memory[ch8.indexRegister+2] = value % 10
	value /= 10
	ch8.memory[ch8.indexRegister+1] = value % 10
	value /= 10
	ch8.memory[ch8.indexRegister] = value % 10
}

// Store registers 0 to x in memory starting from location in indexRegister
func (ch8 *Chip8) OP_Fx55() {
	Vx := (ch8.opcode & 0xF00) >> 8
	for i := 0; uint16(i) <= Vx; i++ {
		ch8.memory[ch8.indexRegister+uint16(i)] = uint8(ch8.registers[i])
	}
}

// Read registers 0 to x in memory starting from location in indexRegister
func (ch8 *Chip8) OP_Fx65() {
	Vx := (ch8.opcode & 0xF00) >> 8
	for i := 0; uint16(i) <= Vx; i++ {
		ch8.registers[i] = uint8(ch8.memory[ch8.indexRegister+uint16(i)])
	}
}
