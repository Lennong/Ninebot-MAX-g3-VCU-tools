package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/common-nighthawk/go-figure"
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

const header = "900800200D0200081502000817020008190200081B0200081D020008000000000000000000000000000000001F020008210200080000000023020008250200082702000827020008270200082702000827020008270200082702000827020008270200082702000827020008270200082702000827020008270200082702000827020008270200082702000827020008270200082702000827020008270200082702000827020008270200082702000827020008270200082702000827020008270200082702000827020008270200082702000827020008270200082702000827020008270200082702000800000000000000000000000000000000000000000000000027020008270200080000000027020008270200080000000000000000270200082702000827020008270200080000000000000000000000000000000000000000000000000000000027020008000000000000000027020008270200080000000000000000000000002702000800F002F800F03AF80AA090E8000C82448344AAF10107DA4501D100F02FF8AFF2090EBAE80F0013F0010F18BFFB1A43F001031847180C0000380C0000103A24BF78C878C1FAD8520724BF30C830C144BF04680C60704700000023002400250026103A28BF78C1FBD8520728BF30C148BF0B6070471FB51FBD10B510BD00F0AAF81146FFF7F7FF00F0ADFC00F0C8F803B4FFF7F2FF03BC00F083FA00000948804709480047FEE7FEE7FEE7FEE7FEE7FEE7FEE7FEE7FEE7FEE704480549054A064B70470000750600087101000890020020900800209004002090040020704753EA020C00F069802DE9F04B4FF00006002B1FBFB3FA83F503FA05F424FA05F65E4012BF1643B2FA82F502FA05F4C5F120051EBF22FA05FC44EA0C04203556EA044C4FEA144418BF641C4FF000084FF00009904271EB030C39D3002919BFB1FA81F701FA07F6B0FA80F700FA07F6C7F120071EBF20FA07FC46EA0C062037B6FBF4FCA7EB0507103F07F01F0BCBF120060CFA0BFB2CFA06F644BFB3460026202FA4BF5E464FF0000B5BEA060C08BF4FF0010B19EB0B09ABFB027C48EB0608C01B06FB02CC0BFB03CC71EB0C01C1E70B46024641464846BDE8F08B13B54FF000004FF00001AFF30080BDE81C4070477047704770477047754600F02BF8AE4605006946534620F00700854618B020B5FFF764FFBDE820404FF000064FF000074FF000084FF0000B21F00701AC46ACE8C009ACE8C009ACE8C009ACE8C0098D46704710B50446AFF300802046BDE81040FFF72FBF004870473002002010B5042000F0E4FC074800680749B0FBF1F007490860084600684FF47A7148430449086010BD00003800002040420F002C0000203000002001B500200099C1F309021AB1012202EB912000E0880A08BD2DE9F04705460E4600274FF00308A94600F098FB142000F01DFB3046FFF7E4FF074600240BE04FF4806101FB049000F069FB8046B8F1030F00D002E0641CBC42F1D300BF00F028FBBC4202D00020BDE8F0870120FBE710B50346002002E01C5C0C54401C9042FAD310BD2DE9F04780468946144646464FF0030A00F063FB4F46002510E03988304600F0EAFA8246BAF1030F00D009E030883988884200D004E0B61CBF1CAD1CA542ECD300BF00F0F4FAA54202D20020BDE8F0870120FBE7000010B512484068B8B172B600F023F868B900200E49486000F09DF80D4800F086F840B921211F2000F0ABF803E021211F2000F0A6F862B607E0054800F077F818B921211F2000F09CF810BD00003C0000200010000870B50024002500261A48806808B9012070BD1948006819490840B0F1005F01D00920F5E713488168154800F02DF808B10220EDE7124D104E002414E04FF4807210493046FFF77DFF4FF480720D492846FFF781FF08B90320DAE706F5807605F5807504F5807403488068A042E6D80020CEE700003C000020000001080000FE2F001000084800002070B504460D46002672B629462046FFF729FF00B9022662B6304670BD10B50C2208490948FFF749FF0648006807490968884205D00348054B0ECB0EC000F01CF810BD00003C00002000F80108A00D000870B50446206807490840B0F1005F07D16568206880F3088800BFA847012070BD0020FCE70000FE2F10B5002472B60C210648FFF7EFFE04462CB10C2204490348FFF71DFF044662B6204610BD00F801083C000020202902D10A4A107003E0212901D1094A1070074A1278074B1B789A4202DD044A127801E0034A1278034B1A72704700002800002029000020220200202248C06920F0805000F180501F49C8611F48007840F004001D49503981F850001A48C06920F080501849C8610846006820F00100401C086000BF14480068C0F340000028F9D01148406820F003000F49486000BF0D484068C0F381000028F9D10A4800680B490840084908600020486041F61070C8624FF4801008634FF41F008860C80304490860704700000010024050700040FFFFF2FE08ED00E000BF70470349496860F30711014A516070470000001002400349496860F30A21014A516070470000001002400349496860F3CD21014A51607047000000100240012807D10749496D21F030013031054A516505E00349496D21F03001014A51657047000000100240052827D2DFE800F0030A11181F00134A126861F30002114B1A601CE00F4A126861F310420D4B1A6015E00C4A126861F318620A4B1A600EE0084A126A61F30002064B1A6207E0054A526A61F30002034B5A6200E000BF00BF704700000010024010B501460020094B03EB1142126801F01F040123A3401A4001F01F040123A3409A4200D000E0012010BD00000010024030B50024002500E0641C1120FFF7E0FF012802D0B4F5405FF6D31120FFF7D8FF012801D0002500E00125284630BD000030B502460020002372B9354C646824F48034334D6C6033482C46246B24F0007404F100742C6319E02D4C646824F4803404F580342A4D6C60012A06D12A482C46646824F400346C6008E02848244C646824F4003404F50034214D6C60244CA04204D9244CA04201D2002326E0224CA04204D9224CA04201D201231EE0204CA04204D9204CA04201D2022316E01D4CA04204D91D4CA04201D203230EE01B4CA04204D91B4CA04201D2042306E0184CA04203D9184CA04200D20523094C646861F39544074D6C600D09054C646865F35E74034D6C602C46E46A63F31A64EC6230BD0010024000093D000024F40000127A0060823B00404B4C0080584F00105E5F0094357700101B7F0020BCBE00286BEE0030D73D01D8E9DC011648006820F00100401C1449086000BF12480068C0F340000028F9D00F48406820F003000D49486000BF0C484068C0F381000028F9D10948006809490840074908600020486041F61070C8624FF4801008634FF41F0088607047000000100240FFFFF2FE0349496860F30101014A5160704700000010024002484068C0F381007047000000100240002131E032280ADD194A126832235A434FF0E0235A61A0F1320290B206E0144A126842434FF0E0235A61002000224FF0E0239A611A46126942F001021A6100BF4FF0E022116901F001021AB101F48032002AF5D04FF0E022126922F001024FF0E0231A6100229A610028CBD170470000300000200149C860704700000020024070B504460D4603260A48006920F00100401C0849086125804FF4801000F030F806460448006920F0010002490861304670BD0000002002400348006920F0800080300149086170470020024003200B49C96801F0010109B100200EE00749C968C1F3800109B1012007E00449C968C1F3001109B1022000E0032070470020024000B502460323FFF7E1FF034603E0FFF7DDFF0346521E0BB9002AF8D102B90423184600BD30B5044603250D48006920F00200801C0A49086108464461006920F04000403008614804FFF7DAFF05460448006920F0020002490861284630BD0000002002400248034948600348486070472301674500200240AB89EFCD00F02AF8FFF726FC002010490872FFF725FDFFF7A7FC18E00C48007A88B100BFBFF34F8F0A48006800F4E06009490843001D07490860BFF34F8F00BF00BF00BFFDE74FF47A70FFF71DFFE5E7220200200CED00E00000FA0510B5FFF7CFFE14201949086001210846FFF7DCFD00BFFFF721FE0028FBD007210120FFF733FE01210220FFF7CFFD00BF1920FFF7FBFD0128FAD10020FFF794FD0420FFF7A5FD0420FFF798FD0120FFF7A9FD0220FFF7D8FE00BFFFF7DFFE0228FBD10020FFF79EFD00F004F810BD0000002002402DE9FF5F002400250026A146A246A3460020039002900190FFF7C6FE0090009820B1012814D0022874D115E04348406DC0F3402040B14148006BC0F3406018B13F484049086002E03F483E49086065E03E483C49086061E038484068C0F300463648C06AC00F20BB34484068C0F3834432484068C0F341750DB90F2C03D12801401C044400E0A41C26B9314860432D4908603DE029484068C0F3404020B12A4860432849086033E028486043254908602EE02248C06AC0F3031003901F48C06AC0F3082002901D48C06A00F00700019016B9DFF87CB009E018484068C0F3404010B1DFF864B001E0DFF860B00121019801FA00F0039900FB01FC0298ABFB00700146624600233846FFF791FA0D49086004E0FFE70C480B49086000BF00BF07484068C0F303190B4810F809A00548006820FA0AF003490860BDE8FF9F00100240006CDC023800002000127A000024F40000093D00AC0D0008042808D14FF0E021096941F004014FF0E022116107E04FF0E021096921F004014FF0E022116170475A50000000000000000000000000000000000000010203040607080938140008000000203C000000AC010008741400083C00002054080000C8010008"

func isDumpHeaderValid(filePath string, expected []byte) (bool, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	buf := make([]byte, len(expected))
	_, err = f.Read(buf)
	if err != nil {
		return false, fmt.Errorf("cannot read file: %w", err)
	}

	return bytes.Equal(buf, expected), nil
}

func main() {
	verify := flag.Bool("v", false, "Run verify mode")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)
	var fileName string
	figure.NewFigure("NINEBOT", "", true).Print()
	figure.NewFigure("MAX G3", "", true).Print()
	figure.NewFigure("VCU TOOLS", "", true).Print()

	fmt.Printf("\nNinebot MAX G3 VCU tools")
	fmt.Printf("\nTested with 1.4.8, 1.4.5 and 1.5.5 firmwares only")
	fmt.Printf("\n!!! You perform any actions at your own risk !!!")
	fmt.Printf("\n\n\n\n\n\n")

	err := os.Remove("MEMORY_G3.bin.patched.bin")
	if err == nil {
		fmt.Println("✅ File deleted successfully.")
	}

	defaultFile, err := findFirstBinFile()
	if defaultFile != "" {
		fileName, err = readFileName("Enter filename [default: "+defaultFile+"]: ", defaultFile)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "❌ Error reading filename:", err)
			os.Exit(1)
		}
	} else {
		fileName, err = readFileName("Enter filename: ", defaultFile)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "❌ Error reading filename:", err)
			os.Exit(1)
		}
	}

	if fileName == "" {
		_, _ = fmt.Fprintln(os.Stderr, "❌ File not defined:", err)
		os.Exit(1)
	}

	data, err := os.ReadFile(fileName)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "❌ Error reading file:", err)
		os.Exit(1)
	}

	//verify length
	if len(data) != 0x20000 {
		_, _ = fmt.Fprintln(os.Stderr, "❌ File corrupted")
		os.Exit(1)
	}
	fmt.Printf("✅ Len correct: %d\n", len(data))

	//verif header
	expected, err := hex.DecodeString(header[:len(header)-(len(header)%2)])
	if err != nil {
		panic("\n❌ invalid header signature. File corrupted: " + err.Error())
		os.Exit(1)
	}

	valid, err := isDumpHeaderValid(fileName, expected)
	if err != nil {
		fmt.Println("\n❌ error:", err)
		os.Exit(1)
	}
	if valid {
		fmt.Println("\n✅ VALID header signature. Dump seems to be correct")
	} else {
		fmt.Println("\n❌ invalid header signature. File corrupted")
		os.Exit(1)
	}

	fmt.Println("\nFound serial numbers:")
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

	if *verify {
		fmt.Println("\n✅ Verify done. No changes made. Press any key to exit")
		_, _ = reader.ReadString('\n')
		os.Exit(0)
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
			_, _ = fmt.Fprintf(os.Stderr, "❌ Failed to read speed value\n")
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
			_, _ = fmt.Fprintln(os.Stderr, "❌ Invalid speed value (must be 1–99)")
			os.Exit(1)
		}

		for _, offset := range speedOffsets {
			err := writeByteAt(data, offset, byte(speedVal))
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "❌ Failed to write speed value\n")
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
		transferKey(data)
	}

	outFile := fileName + ".patched.bin"
	err = os.WriteFile(outFile, data, 0644)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "❌ Error writing output file:", err)
		os.Exit(1)
	}
	fmt.Println("✅ All changes written to:", outFile)
}

func readFileName(promt, defaultFile string) (string, error) {
	binFiles := getBinFiles(".")
	var completerItems []readline.PrefixCompleterInterface
	for _, f := range binFiles {
		completerItems = append(completerItems, readline.PcItem(f))
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          promt,
		AutoComplete:    readline.NewPrefixCompleter(completerItems...),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		panic(err)
	}
	defer func(rl *readline.Instance) {
		_ = rl.Close()
	}(rl)

	line, err := rl.Readline()
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}
	if len(line) == 0 {
		line = defaultFile
	}

	fmt.Println("You selected:", strings.TrimSpace(line))
	return strings.TrimSpace(line), nil
}

func getBinFiles(dir string) []string {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var binFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(strings.ToLower(f.Name()), ".bin") {
			binFiles = append(binFiles, f.Name())
		}
	}
	return binFiles
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

func transferKey(data []byte) {
	sourceName, err := readFileName("Enter source file name", "")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "❌ Error reading filename:", err)
		os.Exit(1)
	}
	sourceData, err := os.ReadFile(sourceName)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "❌ Error reading source file:", err)
		os.Exit(1)
	}
	if len(sourceData) < secretKeyOffset+secretKeyLength {
		_, _ = fmt.Fprintln(os.Stderr, "❌ Source file too small for key extraction")
		os.Exit(1)
	}
	if len(data) < secretKeyOffset+secretKeyLength {
		_, _ = fmt.Fprintln(os.Stderr, "❌ Target file too small for key injection")
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
		_, _ = fmt.Fprintln(os.Stderr, "❌ Invalid mileage value (must be 0–65535)")
		os.Exit(1)
	}
	if err := writeUint16At(data, speedOffset1, uint16(mileageVal)); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "❌ Error writing mileage")
		os.Exit(1)
	}
	if err := writeUint16At(data, speedOffset2, uint16(mileageVal)); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "❌ Error writing mileage")
		os.Exit(1)
	}
	fmt.Printf("✅ Mileage 0x%04X written to both locations\n", mileageVal)
}

func changeSn(reader *bufio.Reader, data []byte) {
	fmt.Print("Enter new serial number (must be 14 characters): ")
	newSerial, _ := reader.ReadString('\n')
	newSerial = strings.TrimSpace(newSerial)
	if len(newSerial) != serialLength {
		_, _ = fmt.Fprintln(os.Stderr, "❌ Invalid serial number format")
		os.Exit(1)
	}

	count, err := patchSerials(data, newSerial)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "❌ Error patching serials:", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Replaced %d serial number(s)\n", count)
}
