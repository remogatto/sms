package main

import (
	"github.com/remogatto/egl/examples/cube/cubelib"
	"runtime"
	"time"
)

const FRAMES_PER_SECOND = 30

type eglLoop struct {
	ticker           *time.Ticker
	displayData                  chan *DisplayData
	paletteValue                 chan PaletteValue
	updateBorder                 chan byte
	pause, terminate             chan int
	paletteR, paletteG, paletteB [32]byte
}

func newEGLLoop() *eglLoop {
	return &eglLoop{
	ticker:    time.NewTicker(time.Duration(1e9 / int(FRAMES_PER_SECOND))),
		displayData:  make(chan *DisplayData),
		paletteValue: make(chan PaletteValue),
		updateBorder: make(chan byte),
		pause:        make(chan int),
		terminate:    make(chan int),
	}
}

func (l *eglLoop) Pause() chan int {
	return l.pause
}

func (l *eglLoop) Terminate() chan int {
	return l.terminate
}

func (l *eglLoop) Display() chan<- *DisplayData {
	return l.displayData
}

func (l *eglLoop) WritePalette() chan<- PaletteValue {
	return l.paletteValue
}

func (l *eglLoop) UpdateBorder() chan<- byte {
	return l.updateBorder
}

func (l *eglLoop) Run() {
	runtime.LockOSThread()
	cubelib.Initialize()

	// Create the 3D world
	world := cubelib.NewWorld()
	world.SetCamera(0.0, 0.0, 4)

	cube := cubelib.NewCube()

	world.Attach(cube)
	angle := float32(0.0)

	for {
		select {
		case <-l.pause:
			l.pause <- 0

		case <-l.terminate:
			l.terminate <- 0

		case data := <-l.displayData:
			cube.AttachTextureFromBuffer(data[:], DISPLAY_WIDTH, DISPLAY_HEIGHT)

		case value := <-l.paletteValue:
			l.paletteR[value.index] = value.r
			l.paletteG[value.index] = value.g
			l.paletteB[value.index] = value.b

		case <-l.updateBorder:
	
		case <-l.ticker.C:
			angle += 0.05
			cube.RotateY(angle)
			world.Draw()
			cubelib.Swap()

		}
	}
}

func (l *eglLoop) render(data *DisplayData) {
}

