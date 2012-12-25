package main

import (
	"flag"
	"fmt"
	"github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"
	"github.com/remogatto/application"
	"log"
	"os"
	"time"
)

// emulatorLoop sends a cmdRenderFrame command to the rendering backend
// (displayLoop) each 1/50 second.
type emulatorLoop struct {
	ticker           *time.Ticker
	sms              *SMS
	pause, terminate chan int
}

// newEmulatorLoop returns a new emulatorLoop instance.
func newEmulatorLoop(displayLoop DisplayLoop) *emulatorLoop {
	emulatorLoop := &emulatorLoop{
		ticker:    time.NewTicker(time.Duration(1e9 / 50)), // 50 Hz
		sms:       newSMS(displayLoop),
		pause:     make(chan int),
		terminate: make(chan int),
	}
	if flag.Arg(0) == "" {
		return nil
	}
	emulatorLoop.sms.loadRom(flag.Arg(0))
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
			l.sms.command <- cmdRenderFrame{}
		}
	}
}

// commandLoop receives commands from the sms command channel and
// forward them to the emulatorLoop or to the displayLoop.
type commandLoop struct {
	pause, terminate chan int
	emulatorLoop     *emulatorLoop
	displayLoop      DisplayLoop
}

// newCommandLoop returns a commandLoop instance.
func newCommandLoop(emulatorLoop *emulatorLoop, displayLoop DisplayLoop) *commandLoop {
	return &commandLoop{
		emulatorLoop: emulatorLoop,
		displayLoop:  displayLoop,
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
		case _cmd := <-l.emulatorLoop.sms.command:
			switch cmd := _cmd.(type) {

			case cmdRenderFrame:
				l.displayLoop.Display() <- l.emulatorLoop.sms.frame()

			case cmdLoadRom:
				l.emulatorLoop.sms.loadRom(cmd.fileName)

			case cmdJoypadEvent:
				l.emulatorLoop.sms.joypad(cmd.value, cmd.event)
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
	fullScreen := flag.Bool("fullscreen", false, "go fullscreen")
	help := flag.Bool("help", false, "Show usage")
	flag.Usage = usage
	flag.Parse()

	if *help {
		usage()
		return
	}

	application.Verbose = *verbose

	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		log.Fatal(sdl.GetError())
	}

	screen := newSDL2xScreen(*fullScreen)
	sdlLoop := newSDLLoop(screen)
	emulatorLoop := newEmulatorLoop(sdlLoop)
	if emulatorLoop == nil {
		usage()
		return
	}
	commandLoop := newCommandLoop(emulatorLoop, sdlLoop)
	inputLoop := newInputLoop(emulatorLoop.sms)

	application.Register("Emulator loop", emulatorLoop)
	application.Register("Command loop", commandLoop)
	application.Register("SDL render loop", sdlLoop)
	application.Register("SDL input loop", inputLoop)

	exitCh := make(chan bool)
	application.Run(exitCh)
	<-exitCh
	sdl.Quit()
}
