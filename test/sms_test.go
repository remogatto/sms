package z80

import (
	"github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"
	sms "github.com/remogatto/sms/segamastersystem"
	"log"
	"testing"
)

func BenchmarkRendering(b *testing.B) {
	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		log.Fatal(sdl.GetError())
	}

	screen := sms.NewSDL2xScreen(false)

	displayLoop := sms.NewSDLLoop(screen)
	go displayLoop.Run()

	sms := sms.NewSMS(displayLoop)

	sms.LoadROM("../roms/blockhead.sms")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		displayLoop.Display() <- sms.RenderFrame()
	}
}
