package sms

import (
	"github.com/scottferg/Go-SDL/sdl"
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

func NewInputLoop(sms *SMS) *inputLoop {
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
					l.sms.Command <- CmdJoypadEvent{keyMap[keyName], JOYPAD_DOWN}
				} else if e.Type == sdl.KEYUP {
					l.sms.Command <- CmdJoypadEvent{keyMap[keyName], JOYPAD_UP}
				}
				if e.Type == sdl.KEYDOWN && keyName == "p" {
					paused := make(chan bool)
					l.sms.Paused = !l.sms.Paused
					l.sms.Command <- CmdPauseEmulation{paused}
					<-paused
				}
				if e.Type == sdl.KEYDOWN && keyName == "d" {
					l.sms.Paused = true
					paused := make(chan bool)
					l.sms.Command <- CmdPauseEmulation{paused}
					<-paused
					l.sms.Command <- CmdShowCurrentInstruction{}
				}
				if e.Keysym.Sym == sdl.K_ESCAPE {
					application.Exit()
				}

			}
		}
	}
}
