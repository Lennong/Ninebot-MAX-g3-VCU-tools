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

var speedOffsets = []int{
	0x1F08D,
	0x1F091,
	0x1F48D,
	0x1F491,
}

const (
	prefix          = "1CG"
	skipSerial      = "1CGC0000000001"
	serialLength    = 14
	speedOffset1    = 0x0001F0C4
	speedOffset2    = 0x0001F4C4
	secretKeyOffset = 0x1F5B4
	secretKeyLength = 12
)

func SetSn(data []byte, newSerial string, reader *bufio.Reader) {
	newSerial = strings.ToUpper(strings.TrimSpace(newSerial))
	if len(newSerial) != serialLength {
		_, _ = fmt.Fprintln(os.Stderr, "\n❌ Invalid serial number format")
		_, _ = reader.ReadString('\n')
		os.Exit(1)
	}

	count := 0
	for i := 0; i <= len(data)-serialLength; i++ {
		if bytes.Equal(data[i:i+3], []byte(prefix)) {
			sn := data[i : i+serialLength]
			if bytes.Equal(sn, []byte(skipSerial)) {
				i += serialLength - 1
				continue
			}
			copy(data[i:i+serialLength], newSerial)
			count++
			i += serialLength - 1
		}
	}
	if count == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "\n❌ no serials replaced:")
		_, _ = reader.ReadString('\n')
		os.Exit(1)
	}

	fmt.Printf("\n✅ Replaced %d serial number(s)\n", count)
}

func SetMileage(data []byte, mileageStr string, reader *bufio.Reader) {
	mileageVal, err := strconv.Atoi(mileageStr)
	if err != nil || mileageVal < 0 || mileageVal > 0xFFFF {
		_, _ = fmt.Fprintln(os.Stderr, "\n❌ Invalid mileage value (must be 0–65535)")
		_, _ = reader.ReadString('\n')
		os.Exit(1)
	}
	if err = writeUint16At(data, speedOffset1, uint16(mileageVal)); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "\n❌ Error writing mileage")
		_, _ = reader.ReadString('\n')
		os.Exit(1)
	}

	if err = writeUint16At(data, speedOffset2, uint16(mileageVal)); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "\n❌ Error writing mileage")
		_, _ = reader.ReadString('\n')
		os.Exit(1)
	}

	fmt.Printf("\n✅ Mileage 0x%04X written to both locations\n", mileageVal)
}

func SetSpeed(data []byte, speedStr string, reader *bufio.Reader) {
	speedVal, err := strconv.Atoi(speedStr)
	if err != nil || speedVal < 1 || speedVal > 125 {
		_, _ = fmt.Fprintln(os.Stderr, "\n❌ Invalid speed value (must be 1–99)")
		_, _ = reader.ReadString('\n')
		os.Exit(1)
	}

	for _, offset := range speedOffsets {
		err = writeByteAt(data, offset, byte(speedVal))
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "\n❌ Failed to write speed value\n")
			_, _ = reader.ReadString('\n')
			os.Exit(1)
		}
	}

	fmt.Printf("\n✅ Speed 0x%02X written to all offsets\n", speedVal)
}

func SetUidKey(data []byte, reader *bufio.Reader) {
	sourceName, err := readFileName("\nEnter source file name with original key: ", "")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "\n❌ Error reading filename:", err)
		_, _ = reader.ReadString('\n')
		os.Exit(1)
	}

	sourceData, err := os.ReadFile(sourceName)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "\n❌ Error reading source file:", err)
		_, _ = reader.ReadString('\n')
		os.Exit(1)
	}

	if len(sourceData) < secretKeyOffset+secretKeyLength {
		_, _ = fmt.Fprintln(os.Stderr, "\n❌ Source file too small for key extraction")
		_, _ = reader.ReadString('\n')
		os.Exit(1)
	}

	if len(data) < secretKeyOffset+secretKeyLength {
		_, _ = fmt.Fprintln(os.Stderr, "\n❌ Target file too small for key injection")
		_, _ = reader.ReadString('\n')
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

func writeUint16At(buf []byte, offset int, value uint16) error {
	if offset+2 > len(buf) {
		return fmt.Errorf("offset out of bounds")
	}
	binary.LittleEndian.PutUint16(buf[offset:], value)

	return nil
}
