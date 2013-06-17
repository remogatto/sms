package z80

import (
	"github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"
	smslib "github.com/remogatto/sms/segamastersystem"
	"log"
	"testing"
)

func BenchmarkRendering(b *testing.B) {
	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		log.Fatal(sdl.GetError())
	}

	screen := smslib.NewSDL2xScreen(false)

	sms := smslib.NewSMS(screen)

	sms.LoadROM("../roms/blockhead.sms")
	
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sms.Frame().Render()
	}
}

// func BenchmarkCPU(b *testing.B) {
// 	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
// 		log.Fatal(sdl.GetError())
// 	}

// 	screen := smslib.NewSDL2xScreen(false)

// 	displayLoop := smslib.NewSDLLoop(screen)
// 	go displayLoop.Run()

// 	sms := smslib.NewSMS(displayLoop)

// 	sms.LoadROM("../roms/blockhead.sms")
	
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		sms.RenderFrame()
// 	}
// }
