package main

import (
	"github.com/remogatto/application"
	"github.com/remogatto/gospeccy/src/z80"
)

var hblankcount = 0

const TStatesPerFrame = 220 // Number of T-states per frame
const PAGE_SIZE = 0x4000

const (
	JOYPAD_DOWN = iota
	JOYPAD_UP
)

type cmdRenderFrame struct{}

type cmdLoadRom struct {
	fileName string
}

type cmdJoypadEvent struct {
	value int
	event byte
}

type SMS struct {
	cpu      *z80.Z80
	memory   *Memory
	vdp      *vdp
	ports    *Ports
	joystick int
	command  chan interface{}
}

func newSMS(displayLoop DisplayLoop) *SMS {
	memory := NewMemory()
	vdp := newVDP(displayLoop)
	ports := NewPorts()
	cpu := z80.NewZ80(memory, ports)

	sms := &SMS{
		cpu:      cpu,
		memory:   memory,
		ports:    ports,
		vdp:      vdp,
		joystick: 0xffff,
		command:  make(chan interface{}),
	}
	sms.memory.init(cpu)
	sms.ports.init(sms)
	return sms
}

func (sms *SMS) loadRom(fileName string) {
	data, err := readROM(fileName)
	if err != nil {
		panic(err)
	}
	size := len(data)
	// Calculate number of pages from file size and create array appropriately
	numROMBanks := size / PAGE_SIZE
	sms.memory.romBanks = make([][]byte, numROMBanks)
	for i := 0; i < numROMBanks; i++ {
		sms.memory.romBanks[i] = make([]byte, PAGE_SIZE)
		// Read file into pages array
		for j := 0; j < PAGE_SIZE; j++ {
			sms.memory.romBanks[i][j] = data[(i*PAGE_SIZE)+j]
		}

	}
	for i := 0; i < 3; i++ {
		sms.memory.pages[i] = byte(i % numROMBanks)
	}
	sms.memory.romPageMask = byte(numROMBanks - 1)
}

func (sms *SMS) frame() *DisplayData {
	sms.vdp.status = 0
	for (sms.vdp.status & 2) == 0 {
		sms.cpu.Tstates = (sms.cpu.Tstates % TStatesPerFrame)
		sms.cpu.EventNextEvent = TStatesPerFrame
		sms.doOpcodes()
		sms.vdp.status = sms.vdp.hblank()
		if sms.vdp.status != 0 {
			sms.cpu.Interrupt()
		}
	}
	return &sms.vdp.displayData
}

func (sms *SMS) joypad(value int, event byte) {
	switch event {
	case JOYPAD_DOWN:
		sms.joystick &= ^value
		break
	case JOYPAD_UP:
		sms.joystick |= value
		break
	default:
		application.Logf("%s", "Unknown joypad event")
		break
	}
}

func (sms *SMS) doOpcodes() {
	// Main instruction emulation loop
	{
		for (sms.cpu.Tstates < sms.cpu.EventNextEvent) && !sms.cpu.Halted {
			sms.memory.ContendRead(sms.cpu.PC(), 4)
			opcode := sms.memory.ReadByteInternal(sms.cpu.PC())
			sms.cpu.R = (sms.cpu.R + 1) & 0x7f
			sms.cpu.IncPC(1)
			z80.OpcodesMap[opcode](sms.cpu)
		}

		if sms.cpu.Halted {
			// Repeat emulating the HALT instruction until 'sms.cpu.eventNextEvent'
			for sms.cpu.Tstates < sms.cpu.EventNextEvent {
				sms.memory.ContendRead(sms.cpu.PC(), 4)
				sms.cpu.R = (sms.cpu.R + 1) & 0x7f
			}
		}
	}
}
