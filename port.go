package main

import (
	"github.com/remogatto/application"
)

type Ports struct {
	sms *SMS
}

func NewPorts() *Ports {
	return &Ports{}
}

func (p *Ports) init(sms *SMS) {
	p.sms = sms
}

func (p *Ports) ReadPort(address uint16) byte {
	return p.ReadPortInternal(address, true)
}

func (p *Ports) ReadPortInternal(address uint16, contend bool) byte {
	switch byte(address) {
	case 0x7e, 0x7f:
		return byte(p.sms.vdp.getLine())
	case 0xdc, 0xc0:
		return byte(p.sms.joystick)
	case 0xdd, 0xc1:
		return byte(p.sms.joystick >> 8)
	case 0xbe:
		return p.sms.vdp.readByte()
	case 0xbd, 0xbf:
		return p.sms.vdp.readStatus()
	case 0xde, 0xdf:
		return 0 // Unknown use
	case 0xf2:
		return 0 // YM2413
	}
	return 0
}

func (p *Ports) WritePort(address uint16, b byte) {
	p.WritePortInternal(address, b, true)
}

func (p *Ports) WritePortInternal(address uint16, b byte, contend bool) {
	switch byte(address) {
	case 0x3f:
		// Nationalisation, pretend we're British.
		natbit := ((b >> 5) & 1)
		if (b & 1) == 0 {
			natbit = 1
		}
		p.sms.joystick = (p.sms.joystick & ^(1 << 6)) | int(natbit<<6)
		natbit = ((b >> 7) & 1)
		if (b & 4) == 0 {
			natbit = 1
		}
		p.sms.joystick = (p.sms.joystick & ^(1 << 7)) | int(natbit<<7)
		break
	case 0x7e, 0x7f:
		//	soundChip.poke(val);
		break
	case 0xbd, 0xbf:
		p.sms.vdp.writeAddr(uint16(b))
		break
	case 0xbe:
		p.sms.vdp.writeByte(b)
		break
	case 0xde, 0xdf:
		break // Unknown use
	case 0xf0, 0xf1, 0xf2:
		break // YM2413 sound support: TODO
		// default:
		// 	console.log('IO port ' + hexbyte(addr) + ' = ' + val);
		// 	break;
	default:
		application.Logf("Write to IO port %x\n", address)
	}
}

func (p *Ports) ContendPortPreio(address uint16)  {}
func (p *Ports) ContendPortPostio(address uint16) {}
