package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func editCustom(reader *bufio.Reader) {
	fmt.Print("\nChoose firmware version: ")
	fmt.Print("\n1) 1.4.8")
	fmt.Print("\n2) 1.5.4")
	fmt.Print("\n3) 1.5.5 (BETA)")
	fmt.Printf("\nEnter: ")

	transfer, _ := reader.ReadString('\n')
	transfer = strings.ToLower(strings.TrimSpace(transfer))

	fileName := ""

	switch transfer {
	case "1":
		fmt.Println("You selected 1.4.8")
		fileName = "MEMORY_G3_1CGBC0000C0000_1.4.8_0.bin"
	case "2":
		fmt.Println("You selected 1.5.4")
		fileName = "MEMORY_G3_1CGCС00007C0000_1.5.4.bin"
	case "3":
		fmt.Println("You selected 1.5.5")
		fileName = "MEMORY_G3_1CGCC1234C1234_1.5.5.bin"
	default:
		{
			fmt.Println("\nInvalid selection")
			os.Exit(1)
		}
	}

	data, err := os.ReadFile("DUMPS/" + fileName)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "❌ Error reading file:", err)
		os.Exit(1)
	}

	fmt.Print("\nEnter new serial number (must be 14 characters like 1CGCC****C****): ")
	newSerial, _ := reader.ReadString('\n')
	newSerial = strings.ToUpper(strings.TrimSpace(newSerial))
	if len(newSerial) != serialLength {
		_, _ = fmt.Fprintln(os.Stderr, "\n❌ Invalid serial number format")
		_, _ = reader.ReadString('\n')
		os.Exit(1)
	}

	SetSn(data, newSerial, reader)
	fmt.Print("Enter new mileage (0–65535): ")
	mileageStr, _ := reader.ReadString('\n')
	mileageStr = strings.TrimSpace(mileageStr)

	SetMileage(data, mileageStr, reader)
	fmt.Print("Enter new speed (1–125): ")
	speedStr, _ := reader.ReadString('\n')
	speedStr = strings.TrimSpace(speedStr)

	SetSpeed(data, speedStr, reader)
	SetUidKey(data, reader)

	outFile := fileName + ".patched.bin"
	err = os.WriteFile(outFile, data, 0644)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "❌ Error writing output file:", err)
		_, _ = reader.ReadString('\n')
		os.Exit(1)
	}

	fmt.Println("✅ All changes written to:", outFile)
	_, _ = reader.ReadString('\n')
}
