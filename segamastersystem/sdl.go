package sms

import (
	"github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"
	"github.com/remogatto/application"
	"log"
	"unsafe"
)

const (
	BPP = 4
	PITCH = 512 << 2
	BPP_PITCH = 4 + 512 << 2
)

type sdlSurface struct {
	surface *sdl.Surface
	pixels uintptr
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
	return &sdlSurface{
		surface,
		uintptr(surface.Pixels),
	}
}

// Create an SDL surface suitable for an unscaled screen
func newUnscaledSurface() *sdlSurface {
	return newSDLSurface(DISPLAY_WIDTH, DISPLAY_HEIGHT)
}

type sdlScreen interface {
	display() *sdlSurface
	border() *sdlSurface
	screen() *sdlSurface
	displayRect() *sdl.Rect
}

type sdl2xScreen struct {
	screenSurface, borderSurface, displaySurface *sdlSurface
	colorValues [32]uint32
}

func NewSDL2xScreen(fullScreen bool) *sdl2xScreen {
	sdlMode := uint32(sdl.SWSURFACE)
	if fullScreen {
		application.Logf("%s", "Activate fullscreen mode")
		sdlMode = sdl.FULLSCREEN
		sdl.ShowCursor(sdl.DISABLE)
	}
	screenSurface := &sdlSurface{sdl.SetVideoMode(SCREEN_WIDTH*2, SCREEN_HEIGHT*2, 32, sdlMode), 0}
	if screenSurface.surface == nil {
		log.Printf("%s", sdl.GetError())
		application.Exit()
		return nil
	}
	borderSurface := newSDLSurface(SCREEN_WIDTH*2, SCREEN_HEIGHT*2)
	if borderSurface.surface == nil {
		log.Printf("%s", sdl.GetError())
		application.Exit()
		return nil
	}
	displaySurface := newSDLSurface(DISPLAY_WIDTH*2, DISPLAY_HEIGHT*2)
	if displaySurface.surface == nil {
		log.Printf("%s", sdl.GetError())
		application.Exit()
		return nil
	}

	return &sdl2xScreen{
	screenSurface: screenSurface, 
	borderSurface: borderSurface, 
	displaySurface: displaySurface,
	}
}

func (screen *sdl2xScreen) Render() {
	displayRect := screen.displayRect()
	displaySurface := screen.display().surface
	borderSurface := screen.border().surface
	// copy display surface on border surface
	borderSurface.Blit(&sdl.Rect{0, 0, 0, 0}, displaySurface, nil)
	// flip surface
	screen.screen().surface.Blit(&sdl.Rect{displayRect.X, displayRect.Y, 0, 0}, borderSurface, nil)
	screen.screen().surface.Flip()
}

func (screen *sdl2xScreen) RasterizePixel(line int, pixelOffset, index byte) {
	offset := uintptr(line << 4 + int(pixelOffset) << 3)
	addr := uintptr(screen.displaySurface.pixels + offset)
	color := screen.colorValues[index]
	// Fill a 2x2 rectangle
	*(*uint32)(unsafe.Pointer(addr)) = color
	*(*uint32)(unsafe.Pointer(addr + BPP)) = color
	*(*uint32)(unsafe.Pointer(addr + PITCH)) = color
	*(*uint32)(unsafe.Pointer(addr + BPP_PITCH)) = color
}

func (screen *sdl2xScreen) UpdateBorder(index byte) {
	color := screen.colorValues[index]
	display := screen.display()
	border := screen.border()
	border.surface.FillRect(nil, color)
	displayRect := screen.displayRect()
	// copy display surface on border surface
	screen.screen().surface.Blit(nil, screen.border().surface, nil)
	// flip surface
	screen.screen().surface.Blit(&sdl.Rect{displayRect.X, displayRect.Y, 0, 0}, display.surface, nil)
	screen.screen().surface.Flip()
}

func (screen *sdl2xScreen) WritePalette(index, r, g, b byte) {
	screen.colorValues[index] = (uint32(r) << 16) | (uint32(g) << 8) | uint32(b)
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

