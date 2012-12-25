package main

import (
	"github.com/remogatto/z80"
)

type Memory struct {
	ram               [0x2000]byte
	cartridgeRam      [0x8000]byte
	pages             [4]byte
	romBanks          [][]byte
	romPageMask       byte
	ramSelectRegister byte
	cpu               *z80.Z80
}

func NewMemory() *Memory {
	return &Memory{}
}

func (memory *Memory) init(cpu *z80.Z80) {
	memory.cpu = cpu
}

func (memory *Memory) reset() {}

func (memory *Memory) ReadByteInternal(address uint16) byte {
	if address < 0x0400 {
		return memory.romBanks[0][address]
	}
	if address < 0x4000 {
		return memory.romBanks[memory.pages[0]&memory.romPageMask][address]
	}
	if address < 0x8000 {
		return memory.romBanks[memory.pages[1]&memory.romPageMask][address-0x4000]
	}
	if address < 0xc000 {
		if (memory.ramSelectRegister & 12) == 8 {
			return memory.cartridgeRam[address-0x8000]
		} else if (memory.ramSelectRegister & 12) == 12 {
			return memory.cartridgeRam[address-0x4000]
		} else {
			return memory.romBanks[memory.pages[2]&memory.romPageMask][address-0x8000]
		}
	}
	if address < 0xe000 {
		return memory.ram[address-0xc000]
	}
	if address < 0xfffc {
		return memory.ram[address-0xe000]
	}
	switch address {
	case 0xfffc:
		return memory.ramSelectRegister
	case 0xfffd:
		return memory.pages[0]
	case 0xfffe:
		return memory.pages[1]
	case 0xffff:
		return memory.pages[2]
	default:
		panic("zoiks")
	}
	return 0
}

func (memory *Memory) WriteByteInternal(address uint16, b byte) {
	if address >= 0xfffc {
		switch address {
		case 0xfffc:
			memory.ramSelectRegister = b
			break
		case 0xfffd:
			memory.pages[0] = b
			break
		case 0xfffe:
			memory.pages[1] = b
			break
		case 0xffff:
			memory.pages[2] = b
			break
		default:
			panic("zoiks")
		}
		return
	}
	if address < 0xc000 {
		return // Ignore ROM writes
	}
	memory.ram[address&0x1fff] = b
}

func (memory *Memory) ReadByte(address uint16) byte {
	return memory.ReadByteInternal(address)
}

func (memory *Memory) WriteByte(address uint16, b byte) {
	memory.WriteByteInternal(address, b)
}

func (memory *Memory) Read(address uint16) byte {
	return 0
}

func (memory *Memory) Write(address uint16, value byte, protectROM bool) {
}

func (memory *Memory) Data() *[0x10000]byte {
	return nil
}

func contendMemory(z80 *z80.Z80, address uint16, time uint) {
	tstates_p := &z80.Tstates
	tstates := *tstates_p
	tstates += time
	*tstates_p = tstates
}

func (memory *Memory) ContendRead(address uint16, time uint) {
	contendMemory(memory.cpu, address, time)
}

func (memory *Memory) ContendReadNoMreq(address uint16, time uint)                   {}
func (memory *Memory) ContendReadNoMreq_loop(address uint16, time uint, count uint)  {}
func (memory *Memory) ContendWriteNoMreq(address uint16, time uint)                  {}
func (memory *Memory) ContendWriteNoMreq_loop(address uint16, time uint, count uint) {}
