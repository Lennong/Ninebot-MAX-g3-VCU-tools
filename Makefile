SRC = main.go
BUILD_DIR = build

.PHONY: all windows linux macos clean
all: windows linux macos

windows:
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=386 go build -o $(BUILD_DIR)/fix_vcu_x86.exe $(SRC)

linux:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=386 go build -o $(BUILD_DIR)/fix_vcu_x86 $(SRC)

macos:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)fix_vcu_arm64 $(SRC)

clean:
	rm -rf $(BUILD_DIR)