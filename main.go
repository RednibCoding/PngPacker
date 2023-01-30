package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	var ddfile = ""
	if len(os.Args) >= 2 {
		ddfile = os.Args[1]
	}

	// ddfile = "C:\\Users\\mlb\\Documents\\DEV\\local\\go\\gothic3-TheBeginning-ImageExtractor\\i"
	if ddfile == "" {
		fmt.Println("Please drag&drop the file onto the 'g3tb-img' executable")
		waitKey()
		os.Exit(0)
	}

	fmt.Println("Processing: " + ddfile + " ...")

	data := readBytes(ddfile)

	// data := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x0a, 0x00, 0x00, 0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x0a, 0x00, 0x00}

	offsets := findPngOffsets(data)
	pngBuffers := collectPngBuffers(offsets, data)
	writePngBuffers(pngBuffers, filepath.Dir(ddfile+"\\"))
}

// func dumpBuffers(buffers [][]byte) {
// 	for _, v := range buffers {
// 		for _, d := range v {
// 			fmt.Printf(" %02X ", d)
// 		}
// 	}
// }

func readBytes(filePath string) []byte {
	buf, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err.Error())
		waitKey()
		os.Exit(0)
	}
	return buf
}

func findPngOffsets(data []byte) []int {
	pngStartPattern := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	offsets := []int{}
	dataCpy := make([]byte, len(data))
	numRemoved := 0
	copy(dataCpy, data)

	for {
		idx := bytes.Index(dataCpy, pngStartPattern)
		if idx == -1 {
			break
		}
		offsets = append(offsets, idx+numRemoved)
		dataCpy = dataCpy[idx+len(pngStartPattern):]
		numRemoved += idx + len(pngStartPattern)
	}

	fmt.Println(strconv.Itoa(len(offsets)) + " PNG images found")
	return offsets
}

func collectPngBuffers(offsets []int, data []byte) [][]byte {
	var buffers [][]byte
	last := -1

	if len(offsets) == 0 {
		fmt.Println("This file does not contain any png data")
		waitKey()
		os.Exit(0)
	}

	if len(offsets) == 1 {
		buffers = append(buffers, data[offsets[0]:])
	} else {
		for _, idx := range offsets {
			if last == -1 {
				last = idx
				continue
			}
			buffers = append(buffers, data[last:idx-1])
			last = idx
		}
		if last < len(data) {
			buffers = append(buffers, data[last:])
		}
	}

	return buffers
}

func writePngBuffers(buffers [][]byte, path string) {
	num := 1
	fullPath := ""
	if len(buffers) > 0 {
		if _, err := os.Stat(path + "_output"); os.IsNotExist(err) {
			err := os.Mkdir(path+"_output", os.ModePerm)
			if err != nil {
				fmt.Println(err.Error())
				waitKey()
				os.Exit(0)
			}
		}
	}
	for _, buf := range buffers {
		fullPath = path + "_output\\image_" + strconv.Itoa(num) + ".png"
		err := os.WriteFile(fullPath, buf, 0644)
		if err != nil {
			fmt.Println(err.Error())
			waitKey()
			os.Exit(0)
		}
		num++
	}

	fmt.Println(strconv.Itoa(len(buffers)) + " png files written to: " + filepath.Dir(fullPath))
	waitKey()
}

func waitKey() {
	var b []byte = make([]byte, 1)
	os.Stdin.Read(b)
}
