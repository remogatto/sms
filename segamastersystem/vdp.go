package sms

var blank_count = 0
var passed = 0

type vdp struct {
	vram                         []byte
	regs                         []byte
	palette                      []byte
	paletteR, paletteG, paletteB []byte
	addr, addrState, addrLatch   uint16
	currentLine                  uint16
	status                       byte
	hBlankCounter                int
	writeRoutine                 func(*vdp, byte)
	readRoutine                  func(*vdp) byte
	displayData                  DisplayData
	displayLoop                  DisplayLoop
}

func (vdp *vdp) updateBorder() {
	borderIndex := 16 + (vdp.regs[7] & 0xf)
	vdp.displayLoop.UpdateBorder() <- borderIndex
}

func (vdp *vdp) writeAddr(val uint16) {
	if vdp.addrState == 0 {
		vdp.addrState = 1
		vdp.addrLatch = val
	} else {
		vdp.addrState = 0
		switch val >> 6 {
		case 0, 1:
			vdp.writeRoutine = writeRAM
			vdp.readRoutine = readRAM
			vdp.addr = vdp.addrLatch | ((val & 0x3f) << 8)
			break
		case 2:
			regnum := val & 0xf
			vdp.regs[regnum] = byte(vdp.addrLatch)
			switch regnum {
			case 7:
				vdp.updateBorder()
				break
			}
			break
		case 3:
			vdp.writeRoutine = writePalette
			vdp.readRoutine = readPalette
			vdp.addr = vdp.addrLatch & 0x1f
			break
		}
	}
}

func writeRAM(vdp *vdp, val byte) {
	vdp.vram[vdp.addr] = byte(val)
	vdp.addr = (vdp.addr + 1) & 0x3fff
}

func writePalette(vdp *vdp, val byte) {
	r := val & 3
	r |= r << 2
	r |= r << 4
	g := (val >> 2) & 3
	g |= g << 2
	g |= g << 4
	b := (val >> 4) & 3
	b |= b << 2
	b |= b << 4
	vdp.paletteR[vdp.addr] = byte(r)
	vdp.paletteG[vdp.addr] = byte(g)
	vdp.paletteB[vdp.addr] = byte(b)

	vdp.displayLoop.WritePalette() <- PaletteValue{vdp.addr, byte(r), byte(g), byte(b)}

	vdp.palette[vdp.addr] = val
	vdp.addr = (vdp.addr + 1) & 0x1f

	vdp.updateBorder()
}

func (vdp *vdp) writeByte(val byte) {
	vdp.addrState = 0
	vdp.writeRoutine(vdp, val)
}

func readRAM(vdp *vdp) byte {
	res := vdp.vram[vdp.addr]
	vdp.addr = (vdp.addr + 1) & 0x1f
	return res
}

func readPalette(vdp *vdp) byte {
	res := vdp.palette[vdp.addr]
	vdp.addr = (vdp.addr + 1) & 0x3fff
	return res
}

func (vdp *vdp) readByte() byte {
	vdp.addrState = 0
	return vdp.readRoutine(vdp)
}

func (vdp *vdp) readStatus() byte {
	res := vdp.status
	vdp.status &= 0x3f
	return res
}

func (vdp *vdp) findSprites(line int) [][]int {
	spriteInfo := int(vdp.regs[5]&0x7e) << 7
	active := make([][]int, 0)
	spriteHeight := 8
	if (vdp.regs[1] & 2) != 0 {
		spriteHeight = 16
	}
	for i := 0; i < 64; i++ {
		y := int(vdp.vram[spriteInfo+i])
		if y == 208 {
			break
		}
		if y >= 240 {
			y -= 256
		}
		if line >= y && line < (y+spriteHeight) {
			if len(active) == 8 {
				vdp.status |= 0x40 // Sprite overflow
				break
			}
			active = append(active, []int{int(vdp.vram[spriteInfo+128+i*2]), int(vdp.vram[spriteInfo+128+i*2+1]), y})
		}
	}
	return active
}

func (vdp *vdp) rasterizeBackground(lineAddr int, pixelOffset byte, tileData int, tileDef int) {
	tileVal0 := vdp.vram[tileDef]
	tileVal1 := vdp.vram[tileDef+1]
	tileVal2 := vdp.vram[tileDef+2]
	tileVal3 := vdp.vram[tileDef+3]
	paletteOffset := byte(0)
	if (tileData & (1 << 11)) != 0 {
		paletteOffset = 16
	}
	if (tileData & (1 << 9)) != 0 {
		for i := 0; i < 8; i++ {
			index := (tileVal0 & 1) | ((tileVal1 & 1) << 1) | ((tileVal2 & 1) << 2) | ((tileVal3 & 1) << 3)
			index += paletteOffset
			if index != 0 {
				vdp.displayData[lineAddr+int(pixelOffset)] = index
			}
			pixelOffset++
			tileVal0 >>= 1
			tileVal1 >>= 1
			tileVal2 >>= 1
			tileVal3 >>= 1
		}
	} else {
		for i := 0; i < 8; i++ {
			index := ((tileVal0 & 128) >> 7) | ((tileVal1 & 128) >> 6) | ((tileVal2 & 128) >> 5) | ((tileVal3 & 128) >> 4)
			index += paletteOffset
			if index != 0 {
				vdp.displayData[lineAddr+int(pixelOffset)] = index
			}
			pixelOffset++
			tileVal0 <<= 1
			tileVal1 <<= 1
			tileVal2 <<= 1
			tileVal3 <<= 1
		}
	}
}

func (vdp *vdp) clearBackground(lineAddr int, pixelOffset byte) {
	for k := 0; k < 8; k++ {
		vdp.displayData[lineAddr+int(pixelOffset)] = 0
		pixelOffset++
	}
}

func (vdp *vdp) rasterizeLine(line int) {
	lineAddr := line << 8

	if (vdp.regs[1] & 64) == 0 {
		for i := 0; i < 256; i++ {
			vdp.displayData[lineAddr+i] = 0
		}
		return
	}

	effectiveLine := line + int(vdp.regs[9])

	if effectiveLine >= 224 {
		effectiveLine -= 224
	}
	sprites := vdp.findSprites(line)
	spritesLen := len(sprites)
	spriteBase := 0
	if (vdp.regs[6] & 4) != 0 {
		spriteBase = 0x2000
	}
	pixelOffset := vdp.regs[8] // * 4
	nameAddr := ((int(vdp.regs[2]) << 10) & 0x3800) + (effectiveLine>>3)<<6
	yMod := effectiveLine & 7
	borderIndex := 16 + (vdp.regs[7] & 0xf)

	for i := 0; i < 32; i++ {
		tileData := int(vdp.vram[nameAddr+i<<1]) | (int(vdp.vram[nameAddr+i<<1+1]) << 8)
		tileNum := int(tileData) & 511
		tileDef := 32 * tileNum
		if (tileData & (1 << 10)) != 0 {
			tileDef += 28 - (yMod << 2)
		} else {
			tileDef += (yMod << 2)
		}
		vdp.clearBackground(lineAddr, pixelOffset)
		// TODO: static top two rows, and static left-hand rows.
		if (tileData & (1 << 12)) == 0 {
			vdp.rasterizeBackground(lineAddr, pixelOffset, tileData, tileDef)
		}
		savedOffset := pixelOffset
		xPos := (i*8 + int(vdp.regs[8])) & 0xff
		// TODO: sprite X-8 shift
		for j := 0; j < 8; j++ {
			writtenTo := false
			for k := 0; k < spritesLen; k++ {
				sprite := sprites[k]
				offset := xPos - sprite[0]
				if offset < 0 || offset >= 8 {
					continue
				}
				spriteLine := int(line) - int(sprite[2])
				spriteAddr := int(spriteBase) + int(sprite[1])<<5 + spriteLine<<2
				effectiveBit := 7 - offset
				sprVal0 := vdp.vram[spriteAddr]
				sprVal1 := vdp.vram[spriteAddr+1]
				sprVal2 := vdp.vram[spriteAddr+2]
				sprVal3 := vdp.vram[spriteAddr+3]
				index := ((sprVal0 >> uint(effectiveBit)) & 1) | (((sprVal1 >> uint(effectiveBit)) & 1) << 1) | (((sprVal2 >> uint(effectiveBit)) & 1) << 2) | (((sprVal3 >> uint(effectiveBit)) & 1) << 3)
				if index == 0 {
					continue
				}
				if writtenTo {
					// We have a collision!.
					vdp.status |= 0x20
					break
				}
				vdp.displayData[lineAddr+int(pixelOffset)] = 16 + index
				writtenTo = true
			}
			xPos++
			pixelOffset++
		}
		if (tileData & (1 << 12)) != 0 {
			vdp.rasterizeBackground(lineAddr, savedOffset, tileData, tileDef)
		}
	}

	if (vdp.regs[0] & (1 << 5)) != 0 {
		// Blank out left hand column.
		for i := 0; i < 8; i++ {
			vdp.displayData[lineAddr+i] = borderIndex
		}
	}
}

func (vdp *vdp) hblank() byte {
	needIrq := byte(0)
	firstDisplayLine := 3 + 13 + 54
	pastEndDisplayLine := firstDisplayLine + 192
	endOfFrame := pastEndDisplayLine + 48 + 3
	if int(vdp.currentLine) >= firstDisplayLine && int(vdp.currentLine) < pastEndDisplayLine {
		vdp.rasterizeLine(int(vdp.currentLine) - firstDisplayLine)
		vdp.hBlankCounter--
		if vdp.hBlankCounter < 0 {
			vdp.hBlankCounter = int(vdp.regs[10])
			vdp.status = vdp.status & 127
			if (vdp.regs[0] & 16) != 0 {
				needIrq |= 1
			}
		}
	}
	vdp.currentLine++
	if int(vdp.currentLine) == endOfFrame {
		vdp.currentLine = 0
		vdp.status |= 128
		if (vdp.regs[1] & 32) != 0 {
			needIrq |= 2
		}
	}
	return needIrq
}

func newVDP(displayLoop DisplayLoop) *vdp {
	vdp := &vdp{
		vram:         make([]byte, 0x4000),
		palette:      make([]byte, 32),
		paletteR:     make([]byte, 32),
		paletteG:     make([]byte, 32),
		paletteB:     make([]byte, 32),
		regs:         make([]byte, 16),
		writeRoutine: func(vdp *vdp, b byte) {},
		readRoutine:  func(vdp *vdp) byte { return 0 },
		displayLoop:  displayLoop,
	}
	vdp.reset()
	return vdp
}

func (vdp *vdp) reset() {
	for i := 0x0000; i < 0x4000; i++ {
		vdp.vram[i] = 0
	}
	for i := 0; i < 32; i++ {
		vdp.paletteR[i], vdp.paletteG[i], vdp.paletteB[i], vdp.palette[i] = 0, 0, 0, 0
	}
	for i := 0; i < 16; i++ {
		vdp.regs[i] = 0
	}
	for i := 2; i <= 5; i++ {
		vdp.regs[i] = 0xff
	}
	vdp.regs[6] = 0xfb
	vdp.regs[10] = 0xff
	vdp.currentLine, vdp.status, vdp.hBlankCounter = 0, 0, 0
}

func (vdp *vdp) getLine() uint16 {
	return (vdp.currentLine - 64) & 0xff
}

func (vdp *vdp) dumpSprites() {
	spriteInfo := (vdp.regs[5] & 0x7e) << 7
	for i := byte(0); i < 64; i++ {
		y := vdp.vram[spriteInfo+i]
		x := vdp.vram[spriteInfo+128+i*2]
		t := vdp.vram[spriteInfo+128+i*2+1]
		println(i, " x: ", x, " y: ", y, " t: ", t)
	}
}
