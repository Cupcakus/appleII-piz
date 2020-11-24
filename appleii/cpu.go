package appleii

/* cpu.go -- 6502 CPU Core
   Copyright (C) 2020 Cupcakus

   This program is free software; you can redistribute it and/or
   modify it under the terms of the GNU General Public License
   as published by the Free Software Foundation; Version 2
   of the License ONLY.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program; if not, write to the Free Software
   Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
*/

import (
	"fmt"
	"log"
)

// addressing modes
const (
	_ = iota
	modeAbsolute
	modeAbsoluteX
	modeAbsoluteY
	modeAccumulator
	modeImmediate
	modeImplied
	modeIndexedIndirect
	modeIndirect
	modeIndirectIndexed
	modeRelative
	modeZeroPage
	modeZeroPageX
	modeZeroPageY
)

const (
	vecNMI   = 0xFFFA
	vecRESET = 0xFFFC
	vecIRQ   = 0xFFFE
)

// instructionModes indicates the addressing mode for each instruction
var instructionModes = [256]byte{
	6, 7, 6, 7, 11, 11, 11, 11, 6, 5, 4, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 12, 12, 6, 3, 6, 3, 2, 2, 2, 2,
	1, 7, 6, 7, 11, 11, 11, 11, 6, 5, 4, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 12, 12, 6, 3, 6, 3, 2, 2, 2, 2,
	6, 7, 6, 7, 11, 11, 11, 11, 6, 5, 4, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 12, 12, 6, 3, 6, 3, 2, 2, 2, 2,
	6, 7, 6, 7, 11, 11, 11, 11, 6, 5, 4, 5, 8, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 12, 12, 6, 3, 6, 3, 2, 2, 2, 2,
	5, 7, 5, 7, 11, 11, 11, 11, 6, 5, 6, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 13, 13, 6, 3, 6, 3, 2, 2, 3, 3,
	5, 7, 5, 7, 11, 11, 11, 11, 6, 5, 6, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 13, 13, 6, 3, 6, 3, 2, 2, 3, 3,
	5, 7, 5, 7, 11, 11, 11, 11, 6, 5, 6, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 12, 12, 6, 3, 6, 3, 2, 2, 2, 2,
	5, 7, 5, 7, 11, 11, 11, 11, 6, 5, 6, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 12, 12, 6, 3, 6, 3, 2, 2, 2, 2,
}

// instructionSizes indicates the size of each instruction in bytes
var instructionSizes = [256]byte{
	2, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	3, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	1, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	1, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 0, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 0, 3, 0, 0,
	2, 2, 2, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
}

// instructionCycles indicates the number of cycles used by each instruction,
// not including conditional cycles
var instructionCycles = [256]byte{
	7, 6, 2, 8, 3, 3, 5, 5, 3, 2, 2, 2, 4, 4, 6, 6,
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	6, 6, 2, 8, 3, 3, 5, 5, 4, 2, 2, 2, 4, 4, 6, 6,
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	6, 6, 2, 8, 3, 3, 5, 5, 3, 2, 2, 2, 3, 4, 6, 6,
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	6, 6, 2, 8, 3, 3, 5, 5, 4, 2, 2, 2, 5, 4, 6, 6,
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	2, 6, 2, 6, 3, 3, 3, 3, 2, 2, 2, 2, 4, 4, 4, 4,
	2, 6, 2, 6, 4, 4, 4, 4, 2, 5, 2, 5, 5, 5, 5, 5,
	2, 6, 2, 6, 3, 3, 3, 3, 2, 2, 2, 2, 4, 4, 4, 4,
	2, 5, 2, 5, 4, 4, 4, 4, 2, 4, 2, 4, 4, 4, 4, 4,
	2, 6, 2, 8, 3, 3, 5, 5, 2, 2, 2, 2, 4, 4, 6, 6,
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	2, 6, 2, 8, 3, 3, 5, 5, 2, 2, 2, 2, 4, 4, 6, 6,
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
}

// instructionPageCycles indicates the number of cycles used by each
// instruction when a page is crossed
var instructionPageCycles = [256]byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 1, 0, 0, 0, 0, 0, 1, 0, 1, 1, 1, 1, 1,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0,
}

// instructionNames indicates the name of each instruction
var instructionNames = [256]string{
	"BRK", "ORA", "KIL", "SLO", "NOP", "ORA", "ASL", "SLO",
	"PHP", "ORA", "ASL", "ANC", "NOP", "ORA", "ASL", "SLO",
	"BPL", "ORA", "KIL", "SLO", "NOP", "ORA", "ASL", "SLO",
	"CLC", "ORA", "NOP", "SLO", "NOP", "ORA", "ASL", "SLO",
	"JSR", "AND", "KIL", "RLA", "BIT", "AND", "ROL", "RLA",
	"PLP", "AND", "ROL", "ANC", "BIT", "AND", "ROL", "RLA",
	"BMI", "AND", "KIL", "RLA", "NOP", "AND", "ROL", "RLA",
	"SEC", "AND", "NOP", "RLA", "NOP", "AND", "ROL", "RLA",
	"RTI", "EOR", "KIL", "SRE", "NOP", "EOR", "LSR", "SRE",
	"PHA", "EOR", "LSR", "ALR", "JMP", "EOR", "LSR", "SRE",
	"BVC", "EOR", "KIL", "SRE", "NOP", "EOR", "LSR", "SRE",
	"CLI", "EOR", "NOP", "SRE", "NOP", "EOR", "LSR", "SRE",
	"RTS", "ADC", "KIL", "RRA", "NOP", "ADC", "ROR", "RRA",
	"PLA", "ADC", "ROR", "ARR", "JMP", "ADC", "ROR", "RRA",
	"BVS", "ADC", "KIL", "RRA", "NOP", "ADC", "ROR", "RRA",
	"SEI", "ADC", "NOP", "RRA", "NOP", "ADC", "ROR", "RRA",
	"NOP", "STA", "NOP", "SAX", "STY", "STA", "STX", "SAX",
	"DEY", "NOP", "TXA", "XAA", "STY", "STA", "STX", "SAX",
	"BCC", "STA", "KIL", "AHX", "STY", "STA", "STX", "SAX",
	"TYA", "STA", "TXS", "TAS", "SHY", "STA", "SHX", "AHX",
	"LDY", "LDA", "LDX", "LAX", "LDY", "LDA", "LDX", "LAX",
	"TAY", "LDA", "TAX", "LAX", "LDY", "LDA", "LDX", "LAX",
	"BCS", "LDA", "KIL", "LAX", "LDY", "LDA", "LDX", "LAX",
	"CLV", "LDA", "TSX", "LAS", "LDY", "LDA", "LDX", "LAX",
	"CPY", "CMP", "NOP", "DCP", "CPY", "CMP", "DEC", "DCP",
	"INY", "CMP", "DEX", "AXS", "CPY", "CMP", "DEC", "DCP",
	"BNE", "CMP", "KIL", "DCP", "NOP", "CMP", "DEC", "DCP",
	"CLD", "CMP", "NOP", "DCP", "NOP", "CMP", "DEC", "DCP",
	"CPX", "SBC", "NOP", "ISC", "CPX", "SBC", "INC", "ISC",
	"INX", "SBC", "NOP", "SBC", "CPX", "SBC", "INC", "ISC",
	"BEQ", "SBC", "KIL", "ISC", "NOP", "SBC", "INC", "ISC",
	"SED", "SBC", "NOP", "ISC", "NOP", "SBC", "INC", "ISC",
}

//Status Register Flags
const (
	flagN      uint8 = 1 << 7 //Negative
	flagV      uint8 = 1 << 6 //Overflow
	flagUnused uint8 = 1 << 5 //Unused is always ON
	flagB      uint8 = 1 << 4 //Break
	flagD      uint8 = 1 << 3 //BCD (Decimal)
	flagI      uint8 = 1 << 2 //Interrupt Enable
	flagZ      uint8 = 1 << 1 //Zero
	flagC      uint8 = 1      //Carry
)

type regs struct {
	PC uint16 //Program counter
	AC uint8  //Accumulator
	X  uint8  //X register
	Y  uint8  //Y register
	SR uint8  // Status register
	SP uint8  // Stack pointer
}

//CPU Main 6502 struct
type CPU struct {
	regs
	cycleCount uint64
	bus        *Bus
	al         uint16 // Internal address latch
	print      bool
	jumpTable  [256]func()
}

//NewCPU Creates a new 6502 and resets it
func NewCPU(b *Bus) *CPU {
	cpu := CPU{bus: b}
	cpu.jumpTable = [256]func(){
		// 0        1        2        3        4        5        6         7       8        9        A        B        C        D        E        F
		cpu.brk, cpu.ora, cpu.err, cpu.err, cpu.err, cpu.ora, cpu.asl, cpu.err, cpu.php, cpu.ora, cpu.aslA, cpu.err, cpu.err, cpu.ora, cpu.asl, cpu.err, //0
		cpu.bpl, cpu.ora, cpu.err, cpu.err, cpu.err, cpu.ora, cpu.asl, cpu.err, cpu.clc, cpu.ora, cpu.err, cpu.err, cpu.err, cpu.ora, cpu.asl, cpu.err, //1
		cpu.jsr, cpu.and, cpu.err, cpu.err, cpu.bit, cpu.and, cpu.rol, cpu.err, cpu.plp, cpu.and, cpu.rolA, cpu.err, cpu.bit, cpu.and, cpu.rol, cpu.err, //2
		cpu.bmi, cpu.and, cpu.err, cpu.err, cpu.err, cpu.and, cpu.rol, cpu.err, cpu.sec, cpu.and, cpu.err, cpu.err, cpu.err, cpu.and, cpu.rol, cpu.err, //3
		cpu.rti, cpu.eor, cpu.err, cpu.err, cpu.err, cpu.eor, cpu.lsr, cpu.err, cpu.pha, cpu.eor, cpu.lsrA, cpu.err, cpu.jmp, cpu.eor, cpu.lsr, cpu.err, //4
		cpu.bvc, cpu.eor, cpu.err, cpu.err, cpu.err, cpu.eor, cpu.lsr, cpu.err, cpu.cli, cpu.eor, cpu.err, cpu.err, cpu.err, cpu.eor, cpu.lsr, cpu.err, //5
		cpu.rts, cpu.adc, cpu.err, cpu.err, cpu.err, cpu.adc, cpu.ror, cpu.err, cpu.pla, cpu.adc, cpu.rorA, cpu.err, cpu.jmp, cpu.adc, cpu.ror, cpu.err, //6
		cpu.bvs, cpu.adc, cpu.err, cpu.err, cpu.err, cpu.adc, cpu.ror, cpu.err, cpu.sei, cpu.adc, cpu.err, cpu.err, cpu.err, cpu.adc, cpu.ror, cpu.err, //7
		cpu.err, cpu.sta, cpu.err, cpu.err, cpu.sty, cpu.sta, cpu.stx, cpu.err, cpu.dey, cpu.err, cpu.txa, cpu.err, cpu.sty, cpu.sta, cpu.stx, cpu.err, //8
		cpu.bcc, cpu.sta, cpu.err, cpu.err, cpu.sty, cpu.sta, cpu.stx, cpu.err, cpu.tya, cpu.sta, cpu.txs, cpu.err, cpu.err, cpu.sta, cpu.err, cpu.err, //9
		cpu.ldy, cpu.lda, cpu.ldx, cpu.err, cpu.ldy, cpu.lda, cpu.ldx, cpu.err, cpu.tay, cpu.lda, cpu.tax, cpu.err, cpu.ldy, cpu.lda, cpu.ldx, cpu.err, //A
		cpu.bcs, cpu.lda, cpu.err, cpu.err, cpu.ldy, cpu.lda, cpu.ldx, cpu.err, cpu.clv, cpu.lda, cpu.tsx, cpu.err, cpu.ldy, cpu.lda, cpu.ldx, cpu.err, //B
		cpu.cpy, cpu.cmp, cpu.err, cpu.err, cpu.cpy, cpu.cmp, cpu.dec, cpu.err, cpu.iny, cpu.cmp, cpu.dex, cpu.err, cpu.cpy, cpu.cmp, cpu.dec, cpu.err, //C
		cpu.bne, cpu.cmp, cpu.err, cpu.err, cpu.err, cpu.cmp, cpu.dec, cpu.err, cpu.cld, cpu.cmp, cpu.err, cpu.err, cpu.err, cpu.cmp, cpu.dec, cpu.err, //D
		cpu.cpx, cpu.sbc, cpu.err, cpu.err, cpu.cpx, cpu.sbc, cpu.inc, cpu.err, cpu.inx, cpu.sbc, cpu.nop, cpu.err, cpu.cpx, cpu.sbc, cpu.inc, cpu.err, //E
		cpu.beq, cpu.sbc, cpu.err, cpu.err, cpu.err, cpu.sbc, cpu.inc, cpu.err, cpu.sed, cpu.sbc, cpu.err, cpu.err, cpu.err, cpu.sbc, cpu.inc, cpu.err} //F
	return &cpu
}

//GetCycleCount gets the total running cycle count of the CPU
func (c *CPU) GetCycleCount() uint64 {
	return c.cycleCount
}

//Reset the CPU
func (c *CPU) Reset() {
	//On reset we set all flags and registers to 0
	c.regs.AC = 0
	c.regs.X = 0
	c.regs.Y = 0
	//Interrupts enabled by default
	c.regs.SR = uint8(flagI | flagUnused)
	//A little quirk of the 6502 on reset the stack pointer is set to 0 and then the state
	//is immediately popped off the stack without actually setting anything (3 bytes)
	//So the SP is initialized to 0xFD.
	c.regs.SP = 0xFD
	//Program counter is set to the reset vector
	c.regs.PC = c.read16(vecRESET)

	//Maybe something on the bus wants to reset
	c.bus.Reset()
}

var doPrint = false
var savedOpcode uint8 = 0

//Tick should be called for every clock cycle
func (c *CPU) Tick() int {
	//Fetch the next instruction
	if doPrint {
		c.printInstruction(c.regs.PC)
	}
	opcode := c.read8(c.regs.PC)

	var paged bool
	switch instructionModes[opcode] {
	case modeAbsolute:
		c.al = c.read16(c.PC + 1)
	case modeAccumulator:
		fallthrough
	case modeImplied:
		c.al = 0
	case modeIndexedIndirect:
		c.al = c.read16nowrap(uint16(c.read8(c.PC+1) + c.X))
	case modeAbsoluteX:
		c.al = c.read16(c.PC+1) + uint16(c.X)
		paged = pagesDiffer(c.al-uint16(c.X), c.al)
	case modeAbsoluteY:
		c.al = c.read16(c.PC+1) + uint16(c.Y)
		paged = pagesDiffer(c.al-uint16(c.Y), c.al)
	case modeImmediate:
		c.al = c.PC + 1
	case modeIndirect:
		c.al = c.read16nowrap(c.read16(c.PC + 1))
	case modeIndirectIndexed:
		c.al = c.read16nowrap(uint16(c.read8(c.PC+1))) + uint16(c.Y)
		paged = pagesDiffer(c.al-uint16(c.Y), c.al)
	case modeRelative:
		offset := uint16(c.read8(c.PC + 1))
		if offset < 0x80 {
			c.al = c.PC + 2 + offset
		} else {
			c.al = c.PC + 2 + offset - 0x100
		}
	case modeZeroPage:
		c.al = uint16(c.read8(c.PC + 1))
	case modeZeroPageX:
		c.al = uint16(c.read8(c.PC+1)+c.X) & 0xFF
	case modeZeroPageY:
		c.al = uint16(c.read8(c.PC+1)+c.Y) & 0xFF
	}

	c.regs.PC += uint16(instructionSizes[opcode])
	cycles := c.cycleCount
	c.cycleCount += uint64(instructionCycles[opcode])
	if paged {
		c.cycleCount += uint64(instructionPageCycles[opcode])
	}
	savedOpcode = opcode
	c.jumpTable[opcode]()

	return int(c.cycleCount - cycles)
}

func pagesDiffer(a, b uint16) bool {
	return a&0xFF00 != b&0xFF00
}

// PrintInstruction prints the current CPU state
func (c *CPU) printInstruction(aPC uint16) {
	opcode := c.read8(aPC)
	bytes := instructionSizes[opcode]
	name := instructionNames[opcode]
	w0 := fmt.Sprintf("%02X", c.read8(aPC+0))
	w1 := fmt.Sprintf("%02X", c.read8(aPC+1))
	w2 := fmt.Sprintf("%02X", c.read8(aPC+2))
	if bytes < 2 {
		w1 = "  "
	}
	if bytes < 3 {
		w2 = "  "
	}
	fmt.Printf(
		"%4X  %s %s %s  %s %12s"+
			"A:%02X X:%02X Y:%02X SR:%02X SP:%02X CYC:%3d\n",
		aPC, w0, w1, w2, name, "",
		c.regs.AC, c.regs.X, c.regs.Y, c.regs.SR, c.regs.SP, c.cycleCount)

}

func (c *CPU) read16(aAddr uint16) uint16 {
	//Read 16bits from the BUS
	addr := aAddr
	read := true
	c.bus.Set(&addr, nil, &read)
	lo := uint16(c.bus.data)
	addr++
	c.bus.Set(&addr, nil, &read)
	hi := uint16(c.bus.data)
	return hi<<8 | lo
}

//An odd quirk of the 6502 is that the hi byte is not
//incremented when the lo byte overflows during indirect addressing
func (c *CPU) read16nowrap(aAddr uint16) uint16 {

	//If we aren't going to wrap just move on
	if aAddr&0xFF != 0xFF {
		return c.read16(aAddr)
	}

	//Read 16bits from the BUS
	addr := aAddr
	read := true
	c.bus.Set(&addr, nil, &read)
	lo := uint16(c.bus.data)
	addr &= 0xFF00
	c.bus.Set(&addr, nil, &read)
	hi := uint16(c.bus.data)
	return hi<<8 | lo
}

func (c *CPU) read8(aAddr uint16) uint8 {
	//Read 8bits from the BUS
	read := true
	c.bus.Set(&aAddr, nil, &read)
	return c.bus.data
}

func (c *CPU) write8(aAddr uint16, aData uint8) {
	read := false
	c.bus.Set(&aAddr, &aData, &read)
}

func (c *CPU) write16(aAddr uint16, aData uint16) {
	c.write8(aAddr, uint8(aData>>8))
	c.write8(aAddr, uint8(aData&0xFF))
}

func (c *CPU) pull8() uint8 {
	c.regs.SP++
	addr := 0x100 | uint16(c.regs.SP)
	read := true
	c.bus.Set(&addr, nil, &read)
	return c.bus.data
}

func (c *CPU) pull16() uint16 {
	lo := uint16(c.pull8())
	hi := uint16(c.pull8())
	return hi<<8 | lo
}

func (c *CPU) push16(aVal uint16) {
	c.push8(uint8(aVal >> 8))   //Push HI byte
	c.push8(uint8(aVal & 0xFF)) //Push LO byte
}

func (c *CPU) push8(aVal uint8) {
	addr := 0x100 | uint16(c.regs.SP)
	read := false
	c.bus.Set(&addr, &aVal, &read)
	c.regs.SP--
}

//After an operation that updates flags
func (c *CPU) updateFlags(aVal uint8) {
	if aVal&0x80 != 0 {
		c.regs.SR |= flagN
	} else {
		c.regs.SR &= ^flagN
	}

	if aVal == 0 {
		c.regs.SR |= flagZ
	} else {
		c.regs.SR &= ^flagZ
	}
}

/* CPU Operations */

// NOP - No Operation
func (c *CPU) nop() {
}

// err - Tried to run an unknown opcode
func (c *CPU) err() {
	//log.Fatalf("ILLEGAL OPCODE! 0x%x\n", savedOpcode)
}

// CLD - Clear Decimal
func (c *CPU) cld() {
	c.regs.SR &= ^flagD
}

// SED - Set Decimal
func (c *CPU) sed() {
	c.regs.SR |= flagD
}

// CLI - Clear Interrupts
func (c *CPU) cli() {
	c.regs.SR &= ^flagI
}

// SEI - Enable Interrupts
func (c *CPU) sei() {
	c.regs.SR |= flagI
}

// CLC - Clear Carry
func (c *CPU) clc() {
	c.regs.SR &= ^flagC
}

// SEC - Set Carry
func (c *CPU) sec() {
	c.regs.SR |= flagC
}

// CLV - Clear oVerflow
func (c *CPU) clv() {
	c.regs.SR &= ^flagV
}

// JSR - Jump subroutine
func (c *CPU) jsr() {
	c.push16(c.regs.PC - 1)
	c.regs.PC = c.al
}

// RTS - Return from Subroutine
func (c *CPU) rts() {
	c.regs.PC = c.pull16() + 1
}

// RTI - Return from interrupt
func (c *CPU) rti() {
	c.regs.SR = c.pull8()
	c.regs.PC = c.pull16()
	c.regs.SR &= ^flagB
}

// LDY - Load Y Register
func (c *CPU) ldy() {
	c.regs.Y = c.read8(c.al)
	c.updateFlags(c.regs.Y)
}

// LDX - Load X Register
func (c *CPU) ldx() {
	c.regs.X = c.read8(c.al)
	c.updateFlags(c.regs.X)
}

// LDA - Load Accumulator
func (c *CPU) lda() {
	c.regs.AC = c.read8(c.al)
	c.updateFlags(c.regs.AC)
}

// STY - Store Y Register
func (c *CPU) sty() {
	c.write8(c.al, c.regs.Y)
}

// STX - Store X Register
func (c *CPU) stx() {
	c.write8(c.al, c.regs.X)
}

// STA - Store A Register
func (c *CPU) sta() {
	c.write8(c.al, c.regs.AC)
}

// ADC - Add with carry
func (c *CPU) adc() {
	m := c.read8(c.al)
	result := uint16(c.regs.AC) + uint16(m)
	if c.regs.SR&flagC != 0 {
		//Add carry if needed
		result++
	}
	if result > 0xFF {
		//Set the carry flag if the result was > 255
		c.regs.SR |= flagC
	} else {
		c.regs.SR &= ^flagC
	}

	if (c.regs.AC^m)&0x80 == 0 && (uint16(c.regs.AC)^result)&0x80 != 0 {
		//Set the overflow flag if the result has a sign error
		c.regs.SR |= flagV
	} else {
		c.regs.SR &= ^flagV
	}

	c.regs.AC = uint8(result & 0xFF)
	c.updateFlags(c.regs.AC)
}

// SBC - Subtract with carry
func (c *CPU) sbc() {
	a := c.regs.AC
	m := c.read8(c.al)
	f := uint8(0)
	if c.regs.SR&flagC != 0 {
		f = 1
	}
	c.regs.AC = a - m - (1 - f)
	c.updateFlags(c.regs.AC)
	if int(a)-int(m)-int(1-f) >= 0 {
		c.regs.SR |= flagC
	} else {
		c.regs.SR &= ^flagC
	}
	if (a^m)&0x80 != 0 && (a^c.regs.AC)&0x80 != 0 {
		c.regs.SR |= flagV
	} else {
		c.regs.SR &= ^flagV
	}
}

// AND with accumulator
func (c *CPU) and() {
	c.regs.AC &= c.read8(c.al)
	c.updateFlags(c.regs.AC)
}

// OR with accumulator
func (c *CPU) ora() {
	c.regs.AC |= c.read8(c.al)
	c.updateFlags(c.regs.AC)
}

// ASL Shift left (Accumulator)
func (c *CPU) aslA() {
	if c.regs.AC&0x80 != 0 {
		c.regs.SR |= flagC
	} else {
		c.regs.SR &= ^flagC
	}
	c.regs.AC <<= 1
	c.updateFlags(c.regs.AC)
}

// ASL Shift left
func (c *CPU) asl() {
	m := c.read8(c.al)
	if m&0x80 != 0 {
		c.regs.SR |= flagC
	} else {
		c.regs.SR &= ^flagC
	}
	m <<= 1
	c.write8(c.al, m)
	c.updateFlags(m)
}

// LSR Shift right (Accumulator)
func (c *CPU) lsrA() {
	if c.regs.AC&0x1 != 0 {
		c.regs.SR |= flagC
	} else {
		c.regs.SR &= ^flagC
	}
	c.regs.AC >>= 1
	c.updateFlags(c.regs.AC)
}

// LSR Shift right
func (c *CPU) lsr() {
	m := c.read8(c.al)
	if m&0x1 != 0 {
		c.regs.SR |= flagC
	} else {
		c.regs.SR &= ^flagC
	}
	m >>= 1
	c.write8(c.al, m)
	c.updateFlags(m)
}

// BCC Branch on carry clear
func (c *CPU) bcc() {
	if c.regs.SR&flagC != 0 {
		return
	}
	c.cycleCount++
	if pagesDiffer(c.regs.PC, c.al) {
		c.cycleCount++
	}
	c.regs.PC = c.al
}

// BCS Branch on carry set
func (c *CPU) bcs() {
	if c.regs.SR&flagC == 0 {
		return
	}
	c.cycleCount++
	if pagesDiffer(c.regs.PC, c.al) {
		c.cycleCount++
	}
	c.regs.PC = c.al
}

// BEQ Branch on equal
func (c *CPU) beq() {
	if c.regs.SR&flagZ == 0 {
		return
	}
	c.cycleCount++
	if pagesDiffer(c.regs.PC, c.al) {
		c.cycleCount++
	}
	c.regs.PC = c.al
}

// BNE Branch on carry clear
func (c *CPU) bne() {
	if c.regs.SR&flagZ != 0 {
		return
	}
	c.cycleCount++
	if pagesDiffer(c.regs.PC, c.al) {
		c.cycleCount++
	}
	c.regs.PC = c.al
}

// BMI Branch on minus
func (c *CPU) bmi() {
	if c.regs.SR&flagN == 0 {
		return
	}
	c.cycleCount++
	if pagesDiffer(c.regs.PC, c.al) {
		c.cycleCount++
	}
	c.regs.PC = c.al
}

// BPL Branch on plus
func (c *CPU) bpl() {
	if c.regs.SR&flagN != 0 {
		return
	}
	c.cycleCount++
	if pagesDiffer(c.regs.PC, c.al) {
		c.cycleCount++
	}
	c.regs.PC = c.al
}

// BVC Branch on overflow clear
func (c *CPU) bvc() {
	if c.regs.SR&flagV != 0 {
		return
	}
	c.cycleCount++
	if pagesDiffer(c.regs.PC, c.al) {
		c.cycleCount++
	}
	c.regs.PC = c.al
}

// BVS Branch on overflow set
func (c *CPU) bvs() {
	if c.regs.SR&flagV == 0 {
		return
	}
	c.cycleCount++
	if pagesDiffer(c.regs.PC, c.al) {
		c.cycleCount++
	}
	c.regs.PC = c.al
}

// BIT Bit test
func (c *CPU) bit() {
	m := c.read8(c.al)
	c.regs.SR = m&0xC0 + c.regs.SR&0x3F
	m &= c.regs.AC
	if m != 0 {
		c.regs.SR &= ^flagZ
	} else {
		c.regs.SR |= flagZ
	}
}

// BRK Break
func (c *CPU) brk() {
	log.Fatal("BRK HALTED")
	c.push16(c.regs.PC)
	c.push8(c.regs.SR | flagB) //B flag is set on the value saved to the stack during a BRK
	c.regs.SR |= flagI
	c.regs.PC = c.read16(vecIRQ)
}

// CMP compare (with accumulator)
func (c *CPU) cmp() {
	m := c.read8(c.al)
	c.updateFlags(c.regs.AC - m)
	if c.regs.AC >= m {
		c.regs.SR |= flagC
	} else {
		c.regs.SR &= ^flagC
	}
}

// CPX compare (with X register)
func (c *CPU) cpx() {
	m := c.read8(c.al)
	c.updateFlags(c.regs.X - m)
	if c.regs.X >= m {
		c.regs.SR |= flagC
	} else {
		c.regs.SR &= ^flagC
	}
}

// CPY compare (with Y register)
func (c *CPU) cpy() {
	m := c.read8(c.al)
	c.updateFlags(c.regs.Y - m)
	if c.regs.Y >= m {
		c.regs.SR |= flagC
	} else {
		c.regs.SR &= ^flagC
	}
}

// DEC Decrement memory
func (c *CPU) dec() {
	m := c.read8(c.al) - 1
	c.write8(c.al, m)
	c.updateFlags(m)
}

// DEX Decrement X
func (c *CPU) dex() {
	c.regs.X--
	c.updateFlags(c.regs.X)
}

// DEY Decrement Y
func (c *CPU) dey() {
	c.regs.Y--
	c.updateFlags(c.regs.Y)
}

// INC Increment memory
func (c *CPU) inc() {
	m := c.read8(c.al) + 1
	c.write8(c.al, m)
	c.updateFlags(m)
}

// DEX Increment X
func (c *CPU) inx() {
	c.regs.X++
	c.updateFlags(c.regs.X)
}

// DEY Increment Y
func (c *CPU) iny() {
	c.regs.Y++
	c.updateFlags(c.regs.Y)
}

// EOR Exclusive OR (with accumulator)
func (c *CPU) eor() {
	c.regs.AC ^= c.read8(c.al)
	c.updateFlags(c.regs.AC)
}

// JMP Everybody Jump! Jump!
func (c *CPU) jmp() {
	c.regs.PC = c.al
}

// PHA Push Accumulator
func (c *CPU) pha() {
	c.push8(c.regs.AC)
}

// PLA Pull Accumulator
func (c *CPU) pla() {
	c.regs.AC = c.pull8()
	c.updateFlags(c.regs.AC)
}

// PHP Push status
func (c *CPU) php() {
	c.regs.SR |= flagB
	c.push8(c.regs.SR)
	c.regs.SR &= ^flagB
}

// PLP Pull status
func (c *CPU) plp() {
	c.regs.SR = c.pull8()
	c.regs.SR &= ^flagB
	c.regs.SR |= flagUnused
}

// ROL Rotate left (Accumulator)
func (c *CPU) rolA() {
	f := c.regs.AC & 0x80
	c.regs.AC <<= 1
	if c.regs.SR&flagC != 0 {
		c.regs.AC |= 1
	}
	if f != 0 {
		c.regs.SR |= flagC
	} else {
		c.regs.SR &= ^flagC
	}
	c.updateFlags(c.regs.AC)
}

// ROL Rotate left
func (c *CPU) rol() {
	m := c.read8(c.al)
	f := m & 0x80
	m <<= 1
	if c.regs.SR&flagC != 0 {
		m |= 1
	}
	if f != 0 {
		c.regs.SR |= flagC
	} else {
		c.regs.SR &= ^flagC
	}
	c.write8(c.al, m)
	c.updateFlags(m)
}

// ROR Rotate right (Accumulator)
func (c *CPU) rorA() {
	f := c.regs.AC & 0x1
	c.regs.AC >>= 1
	if c.regs.SR&flagC != 0 {
		c.regs.AC |= 1 << 7
	}
	if f != 0 {
		c.regs.SR |= flagC
	} else {
		c.regs.SR &= ^flagC
	}
	c.updateFlags(c.regs.AC)
}

// ROR Rotate right
func (c *CPU) ror() {
	m := c.read8(c.al)
	f := m & 0x1
	m >>= 1
	if c.regs.SR&flagC != 0 {
		m |= 1 << 7
	}
	if f != 0 {
		c.regs.SR |= flagC
	} else {
		c.regs.SR &= ^flagC
	}
	c.write8(c.al, m)
	c.updateFlags(m)
}

// TAX Transfer AC to X
func (c *CPU) tax() {
	c.regs.X = c.regs.AC
	c.updateFlags(c.regs.X)
}

// TAY Transfer AC to Y
func (c *CPU) tay() {
	c.regs.Y = c.regs.AC
	c.updateFlags(c.regs.Y)
}

// TSX Transfer SP to X
func (c *CPU) tsx() {
	c.regs.X = c.regs.SP
	c.updateFlags(c.regs.X)
}

// TXA Transfer X to AC
func (c *CPU) txa() {
	c.regs.AC = c.regs.X
	c.updateFlags(c.regs.AC)
}

// TYA Transfer Y to AC
func (c *CPU) tya() {
	c.regs.AC = c.regs.Y
	c.updateFlags(c.regs.AC)
}

// TXS Transfer X to SP
func (c *CPU) txs() {
	c.regs.SP = c.regs.X
}
