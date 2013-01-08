package main

import (
	"github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"
	"github.com/remogatto/application"
)

var keyMap = map[string]int{
	"up":    1, // Arrow keys
	"down":  2,
	"left":  4,
	"right": 8,
	"z":     16, // Z and X for fire
	"x":     32,
	"r":     1 << 12, // R for reset button
}

type inputLoop struct {
	sms              *SMS
	pause, terminate chan int
}

func newInputLoop(sms *SMS) *inputLoop {
	return &inputLoop{
		sms:       sms,
		pause:     make(chan int),
		terminate: make(chan int),
	}
}

func (l *inputLoop) Pause() chan int {
	return l.pause
}

func (l *inputLoop) Terminate() chan int {
	return l.terminate
}

func (l *inputLoop) Run() {
	for {
		select {
		case <-l.pause:
			l.pause <- 0

		case <-l.terminate:
			l.terminate <- 0

		case _event := <-sdl.Events:
			switch e := _event.(type) {
			case sdl.QuitEvent:
				application.Exit()
			case sdl.KeyboardEvent:
				keyName := sdl.GetKeyName(sdl.Key(e.Keysym.Sym))
				application.Debugf("%d: %s\n", e.Keysym.Sym, keyName)
				if e.Type == sdl.KEYDOWN {
					l.sms.command <- cmdJoypadEvent{keyMap[keyName], JOYPAD_DOWN}
				} else if e.Type == sdl.KEYUP {
					l.sms.command <- cmdJoypadEvent{keyMap[keyName], JOYPAD_UP}
				}
				if e.Type == sdl.KEYDOWN && keyName == "p" {
					paused := make(chan bool)
					l.sms.paused = !l.sms.paused
					l.sms.command <- cmdPauseEmulation{ paused }
					<-paused
				}
				if e.Type == sdl.KEYDOWN && keyName == "d" {
					l.sms.paused = true
					paused := make(chan bool)
					l.sms.command <- cmdPauseEmulation{ paused }
					<-paused
					l.sms.command <- cmdShowCurrentInstruction{}
				}
				if e.Keysym.Sym == sdl.K_ESCAPE {
					application.Exit()
				}

			}
		}
	}
}
