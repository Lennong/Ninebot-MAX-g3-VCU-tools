package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	prefix          = "1CG"
	skipSerial      = "1CGC0000000001"
	serialLength    = 14
	speedOffset1    = 0x0001F0C4
	speedOffset2    = 0x0001F4C4
	secretKeyOffset = 0x1F5B4
	secretKeyLength = 12
)

var speedOffsets = []int{
	0x1F08D,
	0x1F091,
	0x1F48D,
	0x1F491,
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	var fileName string

	err := os.Remove("MEMORY_G3.bin.patched.bin")
	if err == nil {
		fmt.Println("✅ File deleted successfully.")
	}

	defaultFile, err := findFirstBinFile()
	if defaultFile != "" {
		fmt.Printf("Enter filename [default: %s]: ", defaultFile)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		fileName = defaultFile
		if input != "" {
			fileName = input
		}
	} else {
		fmt.Print("Enter filename: ")
		fileName, _ = reader.ReadString('\n')
		fileName = strings.TrimSpace(fileName)
	}

	if fileName == "" {
		fmt.Fprintln(os.Stderr, "❌ File not defined:", err)
		os.Exit(1)
	}

	data, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "❌ Error reading file:", err)
		os.Exit(1)
	}

	fmt.Println("📄 Found serial numbers:")
	for i := 0; i <= len(data)-serialLength; i++ {
		if bytes.Equal(data[i:i+3], []byte(prefix)) {
			sn := data[i : i+serialLength]
			if bytes.Equal(sn, []byte(skipSerial)) {
				i += serialLength - 1
				continue
			}
			fmt.Printf("-> %s\n", string(sn))
			i += serialLength - 1
		}
	}

	fmt.Print("Do you want to update S/N? (Y/N): ")
	answer, _ := reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))

	if answer == "y" {
		changeSn(reader, data)
	}

	old1, _ := readUint16At(data, speedOffset1)
	old2, _ := readUint16At(data, speedOffset2)
	fmt.Printf("🚗 Current mileage A: %d (%.1f km)\n", old1, float64(old1)/10.0)
	fmt.Printf("🚗 Current mileage B: %d (%.1f km)\n", old2, float64(old2)/10.0)

	fmt.Print("Do you want to update mileage? (Y/N): ")
	answer, _ = reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))

	if answer == "y" {
		changeMileage(reader, data)
	}

	fmt.Println("🚀 Current speed values:")
	for _, offset := range speedOffsets {
		val, err := readByteAt(data, offset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to read speed value\n")
			continue
		}
		fmt.Printf("-> %d (0x%02X)\n", val, val)
	}

	fmt.Print("Do you want to update speed? (Y/N): ")
	answer, _ = reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))

	if answer == "y" {
		fmt.Print("Enter new speed (1–99): ")
		speedStr, _ := reader.ReadString('\n')
		speedStr = strings.TrimSpace(speedStr)

		speedVal, err := strconv.Atoi(speedStr)
		if err != nil || speedVal < 1 || speedVal > 99 {
			fmt.Fprintln(os.Stderr, "❌ Invalid speed value (must be 1–99)")
			os.Exit(1)
		}

		for _, offset := range speedOffsets {
			err := writeByteAt(data, offset, byte(speedVal))
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to write speed value\n")
				os.Exit(1)
			}
		}

		fmt.Printf("✅ Speed 0x%02X written to all offsets\n", speedVal)
	}

	oldKey := data[secretKeyOffset : secretKeyOffset+secretKeyLength]

	fmt.Print("🔑 Old key (hex): ")
	for _, b := range oldKey {
		fmt.Printf("%02X ", b)
	}

	fmt.Printf("\n📦 Old key (base64): %s", base64.StdEncoding.EncodeToString(oldKey))

	fmt.Print("\nDo you want to transfer secret key from another file? (Y/N): ")
	transfer, _ := reader.ReadString('\n')
	transfer = strings.ToLower(strings.TrimSpace(transfer))

	if transfer == "y" {
		transferKey(reader, data)
	}

	outFile := fileName + ".patched.bin"
	err = os.WriteFile(outFile, data, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, "❌ Error writing output file:", err)
		os.Exit(1)
	}
	fmt.Println("✅ All changes written to:", outFile)
}

func readByteAt(buf []byte, offset int) (byte, error) {
	if offset >= len(buf) {
		return 0, fmt.Errorf("offset out of bounds")
	}
	return buf[offset], nil
}

func writeByteAt(buf []byte, offset int, value byte) error {
	if offset >= len(buf) {
		return fmt.Errorf("offset out of bounds")
	}
	buf[offset] = value
	return nil
}

func patchSerials(content []byte, newSerial string) (int, error) {
	count := 0
	for i := 0; i <= len(content)-serialLength; i++ {
		if bytes.Equal(content[i:i+3], []byte(prefix)) {
			sn := content[i : i+serialLength]
			if bytes.Equal(sn, []byte(skipSerial)) {
				i += serialLength - 1
				continue
			}
			copy(content[i:i+serialLength], []byte(newSerial))
			count++
			i += serialLength - 1
		}
	}
	if count == 0 {
		return 0, fmt.Errorf("no serials replaced")
	}
	return count, nil
}

func findFirstBinFile() (string, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".bin") {
			return entry.Name(), nil
		}
	}
	return "", fmt.Errorf("no .bin file found in current directory")
}

func writeUint16At(buf []byte, offset int, value uint16) error {
	if offset+2 > len(buf) {
		return fmt.Errorf("offset out of bounds")
	}
	binary.LittleEndian.PutUint16(buf[offset:], value)
	return nil
}

func readUint16At(buf []byte, offset int) (uint16, error) {
	if offset+2 > len(buf) {
		return 0, fmt.Errorf("offset out of bounds")
	}
	return binary.LittleEndian.Uint16(buf[offset : offset+2]), nil
}

func transferKey(reader *bufio.Reader, data []byte) {
	fmt.Print("Enter source file name: ")
	sourceName, _ := reader.ReadString('\n')
	sourceName = strings.TrimSpace(sourceName)

	sourceData, err := os.ReadFile(sourceName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "❌ Error reading source file:", err)
		os.Exit(1)
	}
	if len(sourceData) < secretKeyOffset+secretKeyLength {
		fmt.Fprintln(os.Stderr, "❌ Source file too small for key extraction")
		os.Exit(1)
	}
	if len(data) < secretKeyOffset+secretKeyLength {
		fmt.Fprintln(os.Stderr, "❌ Target file too small for key injection")
		os.Exit(1)
	}

	newKey := sourceData[secretKeyOffset : secretKeyOffset+secretKeyLength]
	fmt.Printf("\n📦 New key (base64): %s", base64.StdEncoding.EncodeToString(newKey))
	fmt.Print("\n🔑 New key (hex): ")
	for _, b := range newKey {
		fmt.Printf("%02X ", b)
	}

	copy(data[secretKeyOffset:secretKeyOffset+secretKeyLength], newKey)
	fmt.Println("\n✅ Secret key transferred into current working data")
}

func changeMileage(reader *bufio.Reader, data []byte) {
	fmt.Print("Enter new mileage (0–65535): ")
	mileageStr, _ := reader.ReadString('\n')
	mileageStr = strings.TrimSpace(mileageStr)

	mileageVal, err := strconv.Atoi(mileageStr)
	if err != nil || mileageVal < 0 || mileageVal > 0xFFFF {
		fmt.Fprintln(os.Stderr, "❌ Invalid mileage value (must be 0–65535)")
		os.Exit(1)
	}
	if err := writeUint16At(data, speedOffset1, uint16(mileageVal)); err != nil {
		fmt.Fprintln(os.Stderr, "❌ Error writing mileage")
		os.Exit(1)
	}
	if err := writeUint16At(data, speedOffset2, uint16(mileageVal)); err != nil {
		fmt.Fprintln(os.Stderr, "❌ Error writing mileage")
		os.Exit(1)
	}
	fmt.Printf("✅ Mileage 0x%04X written to both locations\n", mileageVal)
}

func changeSn(reader *bufio.Reader, data []byte) {
	fmt.Print("Enter new serial number (must be 14 characters): ")
	newSerial, _ := reader.ReadString('\n')
	newSerial = strings.TrimSpace(newSerial)
	if len(newSerial) != serialLength {
		fmt.Fprintln(os.Stderr, "❌ Invalid serial number format")
		os.Exit(1)
	}

	count, err := patchSerials(data, newSerial)
	if err != nil {
		fmt.Fprintln(os.Stderr, "❌ Error patching serials:", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Replaced %d serial number(s)\n", count)
}
