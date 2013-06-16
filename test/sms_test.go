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

	displayLoop := smslib.NewSDLLoop(screen)
	go displayLoop.Run()

	sms := smslib.NewSMS(displayLoop)

	sms.LoadROM("../roms/blockhead.sms")
	
	numOfGeneratedFrames := 100
	generatedFrames := make([]smslib.DisplayData, numOfGeneratedFrames)

	for i := 0; i < numOfGeneratedFrames; i++ {
		generatedFrames = append(generatedFrames, *sms.RenderFrame())
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, frame := range generatedFrames {
			displayLoop.Display() <- &frame
		}
	}
}

func BenchmarkCPU(b *testing.B) {
	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		log.Fatal(sdl.GetError())
	}

	screen := smslib.NewSDL2xScreen(false)

	displayLoop := smslib.NewSDLLoop(screen)
	go displayLoop.Run()

	sms := smslib.NewSMS(displayLoop)

	sms.LoadROM("../roms/blockhead.sms")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sms.RenderFrame()
	}
}
