package main

import (
	"flag"
	"fmt"
	"github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"
	"github.com/remogatto/application"
	sms "github.com/remogatto/sms/segamastersystem"
	"github.com/remogatto/z80"
	"log"
	"os"
	"runtime/pprof"
	"time"
)

const NUM_FRAMES_FOR_PROFILING = 10000

// drainTicker drains the remaining ticks from the given tick.
func drainTicker(ticker *time.Ticker) {
loop:
	for {
		select {
		case <-ticker.C:
		default:
			break loop
		}
	}
}

// showDisassembled prints out disassembled code.
func showDisassembled(instructions []z80.DebugInstruction) {
	arrow := ""
	for line, instr := range instructions {
		if line == 0 {
			arrow = "=>>"
		} else {
			arrow = ""
		}
		application.Logf(arrow+"\t0x%04x %s\n", instr.Address, instr.Mnemonic)
	}
}

// emulatorLoop sends a cmdRenderFrame command to the rendering backend
// (displayLoop) each 1/50 second.
type emulatorLoop struct {
	ticker           *time.Ticker
	sms              *sms.SMS
	pause, terminate chan int
	pauseEmulation   chan int
}

// newEmulatorLoop returns a new emulatorLoop instance.
func newEmulatorLoop(displayLoop sms.DisplayLoop) *emulatorLoop {
	emulatorLoop := &emulatorLoop{
		ticker:         time.NewTicker(time.Duration(1e9 / 50)), // 50 Hz
		sms:            sms.NewSMS(displayLoop),
		pause:          make(chan int),
		terminate:      make(chan int),
		pauseEmulation: make(chan int),
	}
	if flag.Arg(0) == "" {
		return nil
	}
	emulatorLoop.sms.LoadROM(flag.Arg(0))
	return emulatorLoop
}

// Pause returns the pause channel of the loop.
// If a value is sent to this channel, the loop will be paused.
func (l *emulatorLoop) Pause() chan int {
	return l.pause
}

// Terminate returns the terminate channel of the loop.
// If a value is sent to this channel, the loop will be terminated.
func (l *emulatorLoop) Terminate() chan int {
	return l.terminate
}

// Run runs emulatorLoop.
// The loop sends a cmdRenderFrame command to the sms command channel
// each time it receives a value from the ticker.
func (l *emulatorLoop) Run() {
	for {
		select {
		case <-l.pause:
			l.ticker.Stop()
			l.pause <- 0
		case <-l.terminate:
			l.terminate <- 0
		case <-l.ticker.C:
			l.sms.Command <- sms.CmdRenderFrame{}
		case <-l.pauseEmulation:
			if l.sms.Paused {
				l.ticker.Stop()
				drainTicker(l.ticker)
			} else {
				l.ticker = time.NewTicker(time.Duration(1e9 / 50))
			}
			l.pauseEmulation <- 0
		}
	}
}

// commandLoop receives commands from the sms command channel and
// forward them to the emulatorLoop or to the displayLoop.
type commandLoop struct {
	pause, terminate chan int
	emulatorLoop     *emulatorLoop
	displayLoop      sms.DisplayLoop
	numOfSentFrames  int
	cpuProfiling     bool
}

// newCommandLoop returns a commandLoop instance.
func newCommandLoop(emulatorLoop *emulatorLoop, displayLoop sms.DisplayLoop, cpuProfiling bool) *commandLoop {
	return &commandLoop{
		emulatorLoop: emulatorLoop,
		displayLoop:  displayLoop,
		cpuProfiling: cpuProfiling,
		pause:        make(chan int),
		terminate:    make(chan int),
	}
}

// Pause returns the pause channel of the loop.
// If a value is sent to this channel, the loop will be paused.
func (l *commandLoop) Pause() chan int {
	return l.pause
}

// Terminate returns the terminate channel of the loop.
// If a value is sent to this channel, the loop will be terminated.
func (l *commandLoop) Terminate() chan int {
	return l.terminate
}

// Run runs the commandLoop.
// The loop waits for commands sent to sms.command channel.
func (l *commandLoop) Run() {
	for {
		select {
		case <-l.pause:
			l.pause <- 0
		case <-l.terminate:
			l.terminate <- 0
		case _cmd := <-l.emulatorLoop.sms.Command:
			switch cmd := _cmd.(type) {

			case sms.CmdRenderFrame:
				l.displayLoop.Display() <- l.emulatorLoop.sms.RenderFrame()
				l.numOfSentFrames++
				if l.numOfSentFrames > NUM_FRAMES_FOR_PROFILING && l.cpuProfiling {
					application.Exit()
				}

			case sms.CmdLoadROM:
				l.emulatorLoop.sms.LoadROM(cmd.Filename)

			case sms.CmdJoypadEvent:
				l.emulatorLoop.sms.Joypad(cmd.Value, cmd.Event)

			case sms.CmdPauseEmulation:
				l.emulatorLoop.pauseEmulation <- 0
				<-l.emulatorLoop.pauseEmulation
				cmd.Paused <- l.emulatorLoop.sms.Paused
				if application.Verbose && l.emulatorLoop.sms.Paused {
					// instructions := z80.DisassembleN(l.emulatorLoop.sms.Memory, l.emulatorLoop.sms.Cpu.PC(), 10)
					// showDisassembled(instructions)
				}

			}
		}
	}
}

// usage shows sms executable usage.
func usage() {
	fmt.Fprintf(os.Stderr, "SMS - A Sega Master System emulator written in Go\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n\n")
	fmt.Fprintf(os.Stderr, "\tsms [options] game.sms\n\n")
	fmt.Fprintf(os.Stderr, "Options are:\n\n")
	flag.PrintDefaults()
}

func main() {
	verbose := flag.Bool("verbose", false, "verbose mode")
	debug := flag.Bool("debug", false, "debug mode")
	fullScreen := flag.Bool("fullscreen", false, "go fullscreen")
	cpuProfile := flag.String("cpuprofile", "", "write cpu profile to file")
	help := flag.Bool("help", false, "Show usage")
	flag.Usage = usage
	flag.Parse()

	if *help {
		usage()
		return
	}

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	application.Verbose = *verbose
	application.Debug = *debug

	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		log.Fatal(sdl.GetError())
	}

	screen := sms.NewSDL2xScreen(*fullScreen)
	sdlLoop := sms.NewSDLLoop(screen)
	emulatorLoop := newEmulatorLoop(sdlLoop)
	if emulatorLoop == nil {
		usage()
		return
	}
	cpuProfiling := *cpuProfile != ""
	commandLoop := newCommandLoop(emulatorLoop, sdlLoop, cpuProfiling)
	inputLoop := sms.NewInputLoop(emulatorLoop.sms)

	application.Register("Emulator loop", emulatorLoop)
	application.Register("Command loop", commandLoop)
	application.Register("SDL render loop", sdlLoop)
	application.Register("SDL input loop", inputLoop)

	exitCh := make(chan bool)
	application.Run(exitCh)
	<-exitCh
	sdl.Quit()
}
