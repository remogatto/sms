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

type emulatorLoop struct {
	ticker           *time.Ticker
	sms              *SMS
	pause, terminate chan int
	displayLoop      DisplayLoop
}

func newEmulatorLoop(displayLoop DisplayLoop) *emulatorLoop {
	emulatorLoop := &emulatorLoop{
		ticker:      time.NewTicker(time.Duration(1e9 / 50)), // 50 Hz
		sms:         newSMS(displayLoop),
		displayLoop: displayLoop,
		pause:       make(chan int),
		terminate:   make(chan int),
	}
	if flag.Arg(0) == "" {
		return nil
	}
	emulatorLoop.sms.loadRom(flag.Arg(0))
	return emulatorLoop
}

func (l *emulatorLoop) Pause() chan int {
	return l.pause
}

func (l *emulatorLoop) Terminate() chan int {
	return l.terminate
}

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

type commandLoop struct {
	pause, terminate chan int
	emulatorLoop     *emulatorLoop
	displayLoop      DisplayLoop
}

func newCommandLoop(emulatorLoop *emulatorLoop, displayLoop DisplayLoop) *commandLoop {
	return &commandLoop{
		emulatorLoop: emulatorLoop,
		displayLoop:  displayLoop,
		pause:        make(chan int),
		terminate:    make(chan int),
	}
}

func (l *commandLoop) Pause() chan int {
	return l.pause
}

func (l *commandLoop) Terminate() chan int {
	return l.terminate
}

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
	application.Register("SDL backend loop", sdlLoop)
	application.Register("Input loop", inputLoop)

	exitCh := make(chan bool)
	application.Run(exitCh)
	<-exitCh
	sdl.Quit()
}
