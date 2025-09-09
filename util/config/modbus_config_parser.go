package config

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type ModbusRegisterBitConfig struct {
	Bit  int
	Key  string
	Name string
}

type ModbusRegister struct {
	Name     string
	Key      string
	Address  uint16
	Length   uint16
	Type     int
	Function int
	Scale    float64
	Unit     string
	Bits     []ModbusRegisterBitConfig
}

const (
	MF_NONE  = 0
	MF_HOLD  = 1
	MF_INPUT = 2
	MF_COIL  = 3
)

const (
	MT_NONE  = 0
	MT_INT   = 1
	MT_FLOAT = 2
)

func ParseModbusProtocol(content string) ([]ModbusRegister, error) {
	strReader := strings.NewReader(content)
	br := bufio.NewReader(strReader)

	// Check for UTF-8 BOM (EF BB BF)
	bom, err := br.Peek(3)
	if err == nil && bytes.Equal(bom, []byte{0xEF, 0xBB, 0xBF}) {
		// Skip the BOM
		_, _ = br.Discard(3)
	}

	reader := csv.NewReader(br)
	reader.TrimLeadingSpace = true

	// Read the header
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	// Create field index mapping
	idx := make(map[string]int)
	for i, h := range headers {
		idx[strings.ToLower(strings.TrimSpace(h))] = i
	}

	var result []ModbusRegister
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(record) < 4 {
			continue
		}

		// Clean whitespace from each field
		for i := range record {
			record[i] = strings.TrimSpace(record[i])
		}

		// Skip comment lines that start with // or #
		if len(record) > 0 && (strings.HasPrefix(record[0], "//") || strings.HasPrefix(record[0], "#")) {
			continue
		}

		get := func(field string) string {
			i, ok := idx[field]
			if !ok || i >= len(record) {
				return ""
			}
			return strings.TrimSpace(record[i])
		}
		getType := func(field string) int {
			i, ok := idx[field]
			if !ok || i >= len(record) {
				return MT_NONE
			}
			str := strings.TrimSpace(record[i])
			switch str {
			case "":
				return MT_INT
			case "int":
				return MT_INT
			case "float":
				return MT_FLOAT
			default:
				fmt.Printf("Invalid type: %s\n", str)
				return MT_NONE
			}
		}
		getFunction := func(field string) int {
			i, ok := idx[field]
			if !ok || i >= len(record) {
				return MF_NONE
			}
			str := strings.TrimSpace(record[i])
			switch str {
			case "":
				return MF_HOLD
			case "coil":
				return MF_COIL
			case "input":
				return MF_INPUT
			case "hold":
				return MF_HOLD
			default:
				fmt.Printf("Invalid function: %s\n", str)
				return MT_NONE
			}
		}

		address, _ := strconv.Atoi(get("address"))
		length, _ := strconv.Atoi(get("length"))

		scale := 1.0
		if s := get("scale"); s != "" {
			scale, _ = strconv.ParseFloat(s, 64)
		}

		r := ModbusRegister{
			Name:     get("name"),
			Key:      get("key"),
			Address:  uint16(address),
			Length:   uint16(length),
			Type:     getType("type"),
			Function: getFunction("function"),
			Scale:    scale,
			Unit:     get("unit"),
		}

		// Parse bit definitions
		if bits := get("bits"); bits != "" {
			r.Bits = parseBits(bits)
		}

		result = append(result, r)
	}

	result = applyDefaults(result)
	return result, nil
}

// 解析 "0:bit_key:bit_name;1:..." 格式的位配置字段
func parseBits(bitsStr string) []ModbusRegisterBitConfig {
	var bits []ModbusRegisterBitConfig
	entries := strings.Split(bitsStr, ";")
	for _, entry := range entries {
		parts := strings.Split(strings.TrimSpace(entry), ":")
		if len(parts) != 3 {
			continue
		}
		pos, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			continue
		}
		bits = append(bits, ModbusRegisterBitConfig{
			Bit:  pos,
			Key:  strings.TrimSpace(parts[1]),
			Name: strings.TrimSpace(parts[2]),
		})
	}
	return bits
}

func applyDefaults(regs []ModbusRegister) []ModbusRegister {
	for i := range regs {
		if regs[i].Function == MF_NONE {
			regs[i].Function = MF_INPUT
		}
		if regs[i].Type == MT_NONE {
			regs[i].Type = MT_INT
		}
		if regs[i].Scale == 0 {
			regs[i].Scale = 1
		}

		if regs[i].Key == "" {
			regs[i].Key = "Key" + strconv.Itoa(i)
		}
	}
	return regs
}

func GetModbusRegisterConfig(csvName string) ([]ModbusRegister, error) {
	registers, ok := PROTOCOL_CONFIG.ModbusRegisterMap.Load(csvName)
	if !ok {
		return nil, fmt.Errorf("protocol config not found for %s", csvName)
	}
	return registers, nil
}
