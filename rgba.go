package main

import (
	"fmt"
)

type rgba struct {
	r, g, b, a byte
}

func (color rgba) value32() uint32 {
	return (uint32(color.a) << 24) | (uint32(color.r) << 16) | (uint32(color.g) << 8) | uint32(color.b)
}

func (color rgba) String() string {
	return fmt.Sprintf("R: %d G: %d B: %d A: %d", color.r, color.g, color.b, color.a)
}
