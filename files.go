package main

import (
	"io/ioutil"
)

func readROM(path string) ([]byte, error) {
	fileData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rom = make([]byte, len(fileData))
	copy(rom, fileData)
	return rom, nil
}
