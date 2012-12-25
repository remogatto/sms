package main

import (
	"github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"
	"github.com/remogatto/application"
	"log"
	"unsafe"
)

type sdlSurface struct {
	surface *sdl.Surface
}

func (s *sdlSurface) width() uint {
	return uint(s.surface.W)
}

func (s *sdlSurface) height() uint {
	return uint(s.surface.H)
}

func (s *sdlSurface) bpp() uint {
	return uint(s.surface.Format.BytesPerPixel)
}

func (s *sdlSurface) pitch() uint {
	return uint(s.surface.Pitch)
}

// Return the address of pixel at (x,y)
func (s *sdlSurface) addrXY(x, y uint) uintptr {
	pixels := uintptr(s.surface.Pixels)
	offset := uintptr(y*s.pitch() + x*s.bpp())
	return pixels + offset
}

func newSDLSurface(w, h int) *sdlSurface {
	surface := sdl.CreateRGBSurface(sdl.SWSURFACE, w, h, 32, 0, 0, 0, 0)
	if surface == nil {
		log.Printf("%s", sdl.GetError())
		application.Exit()
		return nil
	}
	return &sdlSurface{surface}
}

// Create an SDL surface suitable for an unscaled screen
func newUnscaledSurface() *sdlSurface {
	return newSDLSurface(DISPLAY_WIDTH, DISPLAY_HEIGHT)
}

type sdlScreen interface {
	renderDisplay(data *DisplayData, paletteR, paletteG, paletteB []byte) *sdlSurface
	display() *sdlSurface
	border() *sdlSurface
	screen() *sdlSurface
	displayRect() *sdl.Rect
}

type sdlUnscaledScreen struct {
	screenSurface, borderSurface, displaySurface *sdlSurface
}

func newSDLUnscaledScreen() *sdlUnscaledScreen {
	screenSurface := &sdlSurface{sdl.SetVideoMode(SCREEN_WIDTH, SCREEN_HEIGHT, 32, sdl.SWSURFACE)}
	if screenSurface.surface == nil {
		log.Printf("%s", sdl.GetError())
		application.Exit()
		return nil
	}
	borderSurface := &sdlSurface{sdl.CreateRGBSurface(sdl.SWSURFACE, SCREEN_WIDTH, SCREEN_HEIGHT, 32, 0, 0, 0, 0)}
	if borderSurface.surface == nil {
		log.Printf("%s", sdl.GetError())
		application.Exit()
		return nil
	}
	displaySurface := &sdlSurface{sdl.CreateRGBSurface(sdl.SWSURFACE, DISPLAY_WIDTH, DISPLAY_HEIGHT, 32, 0, 0, 0, 0)}
	if displaySurface.surface == nil {
		log.Printf("%s", sdl.GetError())
		application.Exit()
		return nil
	}
	return &sdlUnscaledScreen{screenSurface, borderSurface, displaySurface}
}

func (screen *sdlUnscaledScreen) renderDisplay(data *DisplayData, paletteR, paletteG, paletteB []byte) *sdlSurface {
	surface := screen.displaySurface
	surface.surface.Lock()
	for y := uint(0); y < DISPLAY_HEIGHT; y++ {
		wy := y * DISPLAY_WIDTH
		for x := uint(0); x < DISPLAY_WIDTH; x++ {
			addr := surface.addrXY(x, y)
			index := data[wy+x]
			color := rgba{paletteR[index], paletteG[index], paletteB[index], 0}
			*(*uint32)(unsafe.Pointer(addr)) = color.value32()
		}
	}
	surface.surface.Unlock()
	return surface
}

func (screen *sdlUnscaledScreen) displayRect() *sdl.Rect {
	return &sdl.Rect{BORDER_LEFT_RIGHT, BORDER_TOP_BOTTOM, DISPLAY_WIDTH, DISPLAY_HEIGHT}
}

func (screen *sdlUnscaledScreen) border() *sdlSurface {
	return screen.borderSurface
}

func (screen *sdlUnscaledScreen) screen() *sdlSurface {
	return screen.screenSurface
}

func (screen *sdlUnscaledScreen) display() *sdlSurface {
	return screen.displaySurface
}

type sdl2xScreen struct {
	screenSurface, borderSurface, displaySurface *sdlSurface
}

func newSDL2xScreen(fullScreen bool) *sdl2xScreen {
	sdlMode := uint32(sdl.SWSURFACE)
	if fullScreen {
		application.Logf("%s", "Activate fullscreen mode")
		sdlMode = sdl.FULLSCREEN
		sdl.ShowCursor(sdl.DISABLE)
	}
	screenSurface := &sdlSurface{sdl.SetVideoMode(SCREEN_WIDTH*2, SCREEN_HEIGHT*2, 32, sdlMode)}
	if screenSurface.surface == nil {
		log.Printf("%s", sdl.GetError())
		application.Exit()
		return nil
	}
	borderSurface := &sdlSurface{sdl.CreateRGBSurface(sdl.SWSURFACE, SCREEN_WIDTH*2, SCREEN_HEIGHT*2, 32, 0, 0, 0, 0)}
	if borderSurface.surface == nil {
		log.Printf("%s", sdl.GetError())
		application.Exit()
		return nil
	}
	displaySurface := &sdlSurface{sdl.CreateRGBSurface(sdl.SWSURFACE, DISPLAY_WIDTH*2, DISPLAY_HEIGHT*2, 32, 0, 0, 0, 0)}
	if displaySurface.surface == nil {
		log.Printf("%s", sdl.GetError())
		application.Exit()
		return nil
	}
	return &sdl2xScreen{screenSurface, borderSurface, displaySurface}
}

func (screen *sdl2xScreen) renderDisplay(data *DisplayData, paletteR, paletteG, paletteB []byte) *sdlSurface {
	surface := screen.displaySurface
	bpp := uintptr(surface.bpp())
	pitch := uintptr(surface.pitch())
	surface.surface.Lock()
	for y := uint(0); y < DISPLAY_HEIGHT; y++ {
		wy := y * DISPLAY_WIDTH
		for x := uint(0); x < DISPLAY_WIDTH; x++ {
			addr := surface.addrXY(2*x, 2*y)
			index := data[wy+x]
			color := rgba{paletteR[index], paletteG[index], paletteB[index], 0}.value32()
			// Fill a 2x2 rectangle
			*(*uint32)(unsafe.Pointer(addr)) = color
			*(*uint32)(unsafe.Pointer(addr + bpp)) = color
			*(*uint32)(unsafe.Pointer(addr + pitch)) = color
			*(*uint32)(unsafe.Pointer(addr + pitch + bpp)) = color
		}
	}
	surface.surface.Unlock()
	return surface
}

func (screen *sdl2xScreen) border() *sdlSurface {
	return screen.borderSurface
}

func (screen *sdl2xScreen) screen() *sdlSurface {
	return screen.screenSurface
}

func (screen *sdl2xScreen) display() *sdlSurface {
	return screen.displaySurface
}

func (screen *sdl2xScreen) displayRect() *sdl.Rect {
	return &sdl.Rect{BORDER_LEFT_RIGHT * 2, BORDER_TOP_BOTTOM * 2, DISPLAY_WIDTH * 2, DISPLAY_HEIGHT * 2}
}

type sdlLoop struct {
	displayData                  chan *DisplayData
	paletteValue                 chan PaletteValue
	updateBorder                 chan byte
	pause, terminate             chan int
	paletteR, paletteG, paletteB [32]byte
	screen                       sdlScreen
}

func newSDLLoop(screen sdlScreen) *sdlLoop {
	return &sdlLoop{
		screen:       screen,
		displayData:  make(chan *DisplayData),
		paletteValue: make(chan PaletteValue),
		updateBorder: make(chan byte),
		pause:        make(chan int),
		terminate:    make(chan int),
	}
}

func (l *sdlLoop) Pause() chan int {
	return l.pause
}

func (l *sdlLoop) Terminate() chan int {
	return l.terminate
}

func (l *sdlLoop) Display() chan<- *DisplayData {
	return l.displayData
}

func (l *sdlLoop) WritePalette() chan<- PaletteValue {
	return l.paletteValue
}

func (l *sdlLoop) UpdateBorder() chan<- byte {
	return l.updateBorder
}

func (l *sdlLoop) Run() {
	for {
		select {
		case <-l.pause:
			l.pause <- 0

		case <-l.terminate:
			l.terminate <- 0

		case data := <-l.displayData:
			l.render(data)

		case value := <-l.paletteValue:
			println(value.index, value.r, value.g, value.b)
			l.paletteR[value.index] = value.r
			l.paletteG[value.index] = value.g
			l.paletteB[value.index] = value.b

		case value := <-l.updateBorder:
			l.renderBorder(value)

		}
	}
}

func (l *sdlLoop) render(data *DisplayData) {
	displayRect := l.screen.displayRect()
	// render surface
	displaySurface := l.screen.renderDisplay(data, l.paletteR[:], l.paletteG[:], l.paletteB[:])
	// copy display surface on border surface
	l.screen.border().surface.Blit(&sdl.Rect{0, 0, 0, 0}, displaySurface.surface, nil)
	// flip surface
	l.screen.screen().surface.Blit(&sdl.Rect{displayRect.X, displayRect.Y, 0, 0}, l.screen.border().surface, nil)
	l.screen.screen().surface.Flip()
}

func (l *sdlLoop) renderBorder(index byte) {
	color := rgba{l.paletteR[index], l.paletteG[index], l.paletteB[index], 0}.value32()
	display := l.screen.display()
	border := l.screen.border()
	border.surface.FillRect(nil, color)
	displayRect := l.screen.displayRect()
	// copy display surface on border surface
	l.screen.screen().surface.Blit(nil, l.screen.border().surface, nil)
	// flip surface
	l.screen.screen().surface.Blit(&sdl.Rect{displayRect.X, displayRect.Y, 0, 0}, display.surface, nil)
	l.screen.screen().surface.Flip()
}
