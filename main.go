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
	prefix       = "1CG"
	skipSerial   = "1CGC0000000001"
	serialLength = 14
	offset1      = 0x0001F0C4
	offset2      = 0x0001F4C4
)

const secretKeyOffset = 0x1F5B4
const secretKeyLength = 12

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
		return fmt.Errorf("offset 0x%X is out of bounds", offset)
	}
	binary.LittleEndian.PutUint16(buf[offset:], value)
	return nil
}

func readUint16At(buf []byte, offset int) (uint16, error) {
	if offset+2 > len(buf) {
		return 0, fmt.Errorf("offset 0x%X is out of bounds", offset)
	}
	return binary.LittleEndian.Uint16(buf[offset : offset+2]), nil
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
			fmt.Printf("→ [%06X]: %s\n", i, string(sn))
			i += serialLength - 1
		}
	}

	fmt.Print("Do you want to update S/N? (Y/N): ")
	answer, _ := reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))

	if answer == "y" {
		changeSn(reader, data)
	}

	old1, _ := readUint16At(data, offset1)
	old2, _ := readUint16At(data, offset2)
	fmt.Printf("🚗 Current mileage at 0x%X: %d (%.1f km, 0x%04X)\n", offset1, old1, float64(old1)/10.0, old1)
	fmt.Printf("🚗 Current mileage at 0x%X: %d (%.1f km, 0x%04X)\n", offset2, old2, float64(old2)/10.0, old2)

	fmt.Print("Do you want to update mileage? (Y/N): ")
	answer, _ = reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))

	if answer == "y" {
		changeMileage(reader, data)
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

	// ✍️ Save everything in the end
	outFile := fileName + ".patched.bin"
	err = os.WriteFile(outFile, data, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, "❌ Error writing output file:", err)
		os.Exit(1)
	}
	fmt.Println("✅ All changes written to:", outFile)

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
	// Write uint16 to both offsets
	if err := writeUint16At(data, offset1, uint16(mileageVal)); err != nil {
		fmt.Fprintln(os.Stderr, "❌ Error writing mileage:", err)
		os.Exit(1)
	}
	if err := writeUint16At(data, offset2, uint16(mileageVal)); err != nil {
		fmt.Fprintln(os.Stderr, "❌ Error writing mileage:", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Mileage 0x%04X written to 0x%X and 0x%X\n", mileageVal, offset1, offset2)
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
