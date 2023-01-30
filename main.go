package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

/*
PngPacker by Michael Binder

PngPacker is a tool that scans the given file for png signatures and if given, extracts those png files.
It also can pack a given folder of png files into a pack file. While packing, it writes a PngPacker signature
so it knows when unpacking, that it is a file packed by PngPacker. The signature is an arbitrary sequence of five bytes at the beginning see: getPngPackerSignature()
Also it packs the png file names in front of the png signatures to preserve the original file names of the png files.
So the file structure of a packed file looks like as follows:
 --------------------
| PngPackerSignature |  -> the signature to identify wether it is a packed file from PngPacker
| #myImage1.png#     |  -> original file name of the first image
| png signature      |  -> png signature of the first image
| png content        |  -> png content of the first image
| #myImage2.png#     |  -> original file name of the second image
| png signature      |  -> png signature of the second image
| png content        |  -> png content of the second image
| ...                |
 --------------------

 PngPacker can not only unpack png images from its own produced pack files but any file that contains png signatures and content.
 PngPacker can identify automatically wether it is a file packed by itself or a random other file.
*/

func main() {
	var dndFile = ""
	if len(os.Args) >= 2 {
		dndFile = os.Args[1]
	}

	// dndFile = "C:\\Users\\mlb\\Documents\\DEV\\github\\go\\PngPacker\\j_output_packed"

	if dndFile == "" {
		waitExit("Please drag&drop the file onto the 'PngPacker' executable")
	}

	fmt.Println("Processing: " + dndFile + " ...")

	if _, err := os.Stat(dndFile); os.IsNotExist(err) {
		waitExit(dndFile + " not found")
	}

	writeWithPngPackerSignature := true

	if isDirectory(dndFile) {
		packPngs(dndFile, writeWithPngPackerSignature)
	} else {
		unpackPngs(dndFile)
	}
}

// --- Packing ---

func packPngs(path string, writeWithPngPackerSignature bool) {
	pngNames := collectFileNamesInDir(path)
	pngBuffers, pngNames := createPngBuffersFromPngFiles(pngNames, writeWithPngPackerSignature)

	// Go up one directory as we want to create the packfile at the same location where the user has his folder with png files
	dir := filepath.Base(path)
	updir := path[0 : len(path)-len(dir)-1] // +1 to also remove the slash
	writePngBuffersAsPackFile(updir, dir, pngBuffers, pngNames, writeWithPngPackerSignature)
}

func collectFileNamesInDir(path string) []string {
	pngNames := make([]string, 0)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		waitExit(err.Error())
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".png" {
			pngNames = append(pngNames, filepath.Join(path, file.Name()))
		}
	}
	return pngNames
}

func createPngBuffersFromPngFiles(pngFilePaths []string, writePngNames bool) ([][]byte, []string) {
	if len(pngFilePaths) == 0 {
		waitExit("No png files found in directory")
	}

	buffers := make([][]byte, 0)

	for _, path := range pngFilePaths {
		buffer := readBytes(path)
		// skip empty files
		if len(buffer) == 0 {
			continue
		}
		buffers = append(buffers, buffer)
	}

	if writePngNames {
		var filenames []string
		for _, path := range pngFilePaths {
			filenames = append(filenames, filepath.Base(path))
		}
		return buffers, filenames
	}
	return buffers, nil
}

func writePngBuffersAsPackFile(path string, outputFileName string, buffers [][]byte, pngFileNames []string, writeSignature bool) {
	if !isDirectory(path) {
		waitExit(path + " is not a valid output path")
	}

	mergedBuffer := make([]byte, 0)

	if writeSignature {
		if len(pngFileNames) == 0 {
			waitExit("Internal error: pngFileNames must not be empty if writeSignature is true")
		}
		mergedBuffer = append(mergedBuffer, getPngPackerSignature()...)
	}

	for i, buffer := range buffers {
		if writeSignature {
			mergedBuffer = append(mergedBuffer, getPngNamePrePostfix())
			mergedBuffer = append(mergedBuffer, []byte(pngFileNames[i])...)
			mergedBuffer = append(mergedBuffer, getPngNamePrePostfix())
			mergedBuffer = append(mergedBuffer, buffer...)
		} else {
			mergedBuffer = append(mergedBuffer, buffer...)
		}
	}

	fullPath := filepath.Join(path, outputFileName)
	err := os.WriteFile(fullPath+"_packed", mergedBuffer, 0644)
	if err != nil {
		waitExit(err.Error())
	}

	waitExit(strconv.Itoa(len(buffers)) + " png files packed into " + fullPath + "_packed")
}

// --- Unpacking ---

func unpackPngs(filePath string) {
	data := readBytes(filePath)
	var fileNames []string
	offsets := findPngOffsets(data)
	isPngPackerPack := isPngPackerPackFile(data)
	if isPngPackerPack {
		fileNames = findPngFileNamesInPngPackerPack(data, offsets)
	}
	pngBuffers := collectPngBuffers(offsets, data)
	writePngBuffers(pngBuffers, fileNames, filepath.Dir(filePath+"\\"), isPngPackerPack)
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
		waitExit("This file does not contain any png data")
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

func writePngBuffers(buffers [][]byte, fileNames []string, path string, hasFilenamesStoredInPack bool) {
	if hasFilenamesStoredInPack && len(fileNames) == 0 {
		waitExit("Internal error: expected to have file names stored in pack file but fileNames is empty")
	}

	fullPath := ""
	if len(buffers) > 0 {
		if _, err := os.Stat(path + "_output"); os.IsNotExist(err) {
			err := os.Mkdir(path+"_output", os.ModePerm)
			if err != nil {
				waitExit(err.Error())
			}
		}
	}
	fileName := ""
	for i, buf := range buffers {
		if hasFilenamesStoredInPack {
			fileName = fileNames[i]
			fullPath = filepath.Join(path+"_output", fileName)
		} else {
			fileName := "image_"
			leadingZeros := len(strconv.Itoa(len(buffers)))
			format := "%0" + strconv.Itoa(leadingZeros) + "d"
			fullPath = filepath.Join(path+"_output", fileName+fmt.Sprintf(format, i)+".png")
			// fullPath = path + "_output\\image_" + strconv.Itoa(num) + ".png"
			// fullPath = filepath.Join(path+"_output", fileName+strconv.Itoa(i)+".png")
		}
		err := os.WriteFile(fullPath, buf, 0644)
		if err != nil {
			waitExit(err.Error())
		}
	}
	waitExit(strconv.Itoa(len(buffers)) + " png files written to: " + filepath.Dir(fullPath))
}

func findPngFileNamesInPngPackerPack(data []byte, offsets []int) []string {
	pngFileNames := make([]string, 0)
	for _, offset := range offsets {
		if data[offset-1] != getPngNamePrePostfix() {
			waitExit("Internal error[" + fmt.Sprintf("0x%02X", offset) + "]: while scanning png file name in packed file, expected name prefix (" + fmt.Sprintf("0x%02X", getPngNamePrePostfix()) + ") in file, got: " + fmt.Sprintf("0x%02X", data[offset]))
		}
		// skip the name postfix (we are going from right to left)
		nameOffset := offset - 2
		nameBuffer := make([]byte, 0)
		for {
			if data[nameOffset] != getPngNamePrePostfix() {
				nameBuffer = append(nameBuffer, data[nameOffset])
				nameOffset--
			} else {
				reverseSlice(nameBuffer)
				pngFileNames = append(pngFileNames, string(nameBuffer))
				break
			}
		}
	}

	return pngFileNames
}

// --- Helpers ---

func readBytes(filePath string) []byte {
	buf, err := os.ReadFile(filePath)
	if err != nil {
		waitExit(err.Error())
	}
	return buf
}

func isDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		waitExit(err.Error())
	}

	return fileInfo.IsDir()
}

func waitExit(msg ...string) {
	if len(msg) > 0 {
		if msg[0] != "" {
			fmt.Println(msg)
		}
	}
	var b []byte = make([]byte, 1)
	os.Stdin.Read(b)
	os.Exit(0)
}

func getPngPackerSignature() []byte {
	// Just an arbitrary signature so when unpacking it, we know it is a packed file that has been packed by PngPacker
	return []byte{0x23, 0x2F, 0x23, 0x2F, 0x23}
}

func getPngNamePrePostfix() byte {
	return byte(0x23)
}

func isPngPackerPackFile(data []byte) bool {
	idx := bytes.Index(data, getPngPackerSignature())
	return idx == 0
}

func reverseSlice[T comparable](s []T) {
	sort.SliceStable(s, func(i, j int) bool {
		return i > j
	})
}

// func dumpBuffers(buffers [][]byte) {
// 	for _, v := range buffers {
// 		for _, d := range v {
// 			fmt.Printf("0x%02X ", d)
// 		}
// 	}
// }
