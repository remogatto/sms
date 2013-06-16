package sms

import (
	"github.com/remogatto/application"
	"github.com/remogatto/z80"
)

var hblankcount = 0

const TStatesPerFrame = 227 // Number of T-states per frame
const PAGE_SIZE = 0x4000

const (
	JOYPAD_DOWN = iota
	JOYPAD_UP
)

type CmdRenderFrame struct{}

type CmdLoadROM struct {
	Filename string
}

type CmdJoypadEvent struct {
	Value int
	Event byte
}

type CmdPauseEmulation struct {
	Paused chan bool
}

type CmdShowCurrentInstruction struct{}

type SMS struct {
	cpu      *z80.Z80
	memory   *Memory
	vdp      *vdp
	ports    *Ports
	joystick int
	Paused   bool
	Command  chan interface{}
}

func NewSMS(displayLoop DisplayLoop) *SMS {
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
		Command:  make(chan interface{}),
	}
	sms.memory.init(cpu)
	sms.ports.init(sms)
	return sms
}

func (sms *SMS) LoadROM(fileName string) {
	application.Logf("Reading from file %s", fileName)
	data, err := readROM(fileName)
	if err != nil {
		panic(err)
	}
	size := len(data)
	// Calculate number of pages from file size and create array appropriately
	numROMBanks := size / PAGE_SIZE
	sms.memory.romBanks = make([][]byte, numROMBanks)
	application.Logf("Found %d ROM banks", numROMBanks)
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
	sms.memory.maskedPage0 = sms.memory.pages[0] & sms.memory.romPageMask
	sms.memory.maskedPage1 = sms.memory.pages[1] & sms.memory.romPageMask
	sms.memory.maskedPage2 = sms.memory.pages[2] & sms.memory.romPageMask
	sms.memory.romBank0 = make([]byte, PAGE_SIZE)
	copy(sms.memory.romBank0, sms.memory.romBanks[sms.memory.maskedPage0][:])
}

func (sms *SMS) RenderFrame() *DisplayData {
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

func (sms *SMS) Joypad(value int, event byte) {
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
