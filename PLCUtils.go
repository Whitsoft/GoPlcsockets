package PLCUtils
import (
	"bytes"
	"encoding/binary"
	"math"
	"math/rand"
	"plc_h"
	"strings"
	"time"
	"fmt"
)

const MINLEN = 12
const MINCSD = 4
const CSDLEN = 12 //IF Handle + T/O + Item cnt + Type ID (Address) + Len (Address)

func Version(ver, back uint16) []byte {
	buf1 := new(bytes.Buffer)
	buf2 := new(bytes.Buffer)
	err1 := binary.Write(buf1, binary.LittleEndian, ver)
	err2 := binary.Write(buf2, binary.LittleEndian, back)
	if err1 != nil {
		fmt.Println("binary.Write failed:", err1, err2)
	}
	result := Cat2Splices(buf1.Bytes(), buf2.Bytes())
	return result
}

func Cat2Splices(S1, S2 []byte) []byte {
	result := make([]byte, len(S1)+len(S2), cap(S1)+cap(S2))
	result = append(S1, S2...)
	return result
}
func RandSession() [4]byte {
// create int32 random number
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	Rand32 := r1.Int31()
	return Uint32ToByteArray(uint32(Rand32))
	}

func RandContext() [8]byte {
	// create int64 random number
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	Rand64 := r1.Int63()
	return Uint64ToByteArray(uint64(Rand64))
}

func Uint64ToByteArray(Con uint64) [8]byte { //ByteArray0, ByteArray1,...,ByteArray7 byte) uint64 {
	var b [8]byte //:= make([]byte, 8)
	b[0] = byte(Con)
	b[1] = byte(Con >> 8)
	b[2] = byte(Con >> 16)
	b[3] = byte(Con >> 24)
	b[4] = byte(Con >> 32)
	b[5] = byte(Con >> 40)
	b[6] = byte(Con >> 48)
	b[7] = byte(Con >> 56)
	return b
}

func ByteSliceToUint64(data []byte) uint64 {
	return uint64(data[0]) | uint64(data[1])<<8 | uint64(data[2])<<16 | uint64(data[3])<<24 |
		uint64(data[4])<<32 | uint64(data[5])<<40 | uint64(data[6])<<48 | uint64(data[7])<<56
}

func Uint32ToByteArray(Con uint32) [4]byte { //ByteArray0, ByteArray1,...,ByteArray7 byte) uint64 {
	var b [4]byte
	b[0] = byte(Con)
	b[1] = byte(Con >> 8)
	b[2] = byte(Con >> 16)
	b[3] = byte(Con >> 24)
	return b
}

func ByteArrayToUint32(data [4]byte) uint32 {
	return uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
}



//func Fill_Logical_Buffer(PLCPtr plc_h.PPLC_EtherIP_info,)
func GetTNS(Value uint16) uint16 {
	return Value + 1
}



func ContextCompare(Add1, Add2 uint64) int {
	if Add1 != Add2 {
		return plc_h.NOCONTEXTMATCH
	} else {
		return 0
	}
}

func AddressCompare(Add1, Add2 []byte) int {
	var Len int
	Len = len(Add1)
	if Len != len(Add2) {
		return plc_h.NOADDRESSMATCH
	}

	for IDX := 0; IDX < Len; IDX++ {
		if Add1[IDX] != Add2[IDX] {
			return plc_h.NOADDRESSMATCH
		}
	}
	return 0
}

func ByteToUint16(N16 []byte) uint16 {
	if len(N16) != 2 {
		return 0
	}
	return uint16(byte(N16[1])*16 + byte(N16[0]))
}

func IsDigit(C byte) bool {
	if (string(C) >= "0") && (string(C) <= "9") {
		return true
	} else {
		return false
	}
}

func StrnCaseCmp(S1, S2 string, NumChars int) int {
	//S1 >, = or <   1,0,-1
	var US1, US2 string

	US1 = strings.ToUpper(S1)
	US2 = strings.ToUpper(S2)
	if (NumChars > len(S1)) || (NumChars > len(S2)) {
		return -2

		for IDX := 0; IDX < NumChars; IDX++ {
			if US1[IDX] < US2[IDX] {
				return -1
			} else if US1[IDX] > US2[IDX] {
				return 1
			}
		}
	}
	return 0
}

func FirstDigit(S string) int {
	for IDX := range S {
		if IsDigit(S[IDX]) {
			return IDX
		}
	}
	return 0
}

func Delimit(S string) int {
	I := strings.Index(S, ":")
	if I <= 0 {
		I = strings.Index(S, " ")
	}
	if I <= 0 {
		return 0
	} else {
		return I
	}
}

func Float32ToBytes(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)
	return bytes
}

func BytesToFloat32(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	return math.Float32frombits(bits)
}

func BytesToInt32(buf []byte) uint32 {
	return binary.LittleEndian.Uint32(buf)
}

func Int32ToBytes(num uint32) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, num)
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

func BytesToInt16(buf []byte) uint16 {
	return binary.LittleEndian.Uint16(buf)
}

func Int16ToBytes(num uint16) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, num)
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

func GetIntegers(Integers []uint32) []byte { //
	var result, buf []byte
	//	L := len(Integers)
	for I := range Integers {
		buf = Int32ToBytes(Integers[I])
		result = append(result, buf[:]...)
	}
	return result
}

//01 00 00 00 02 00 00 00 03 00 00 00 04 00 00 00 05 00 00 00
func PutIntegers(buf []byte) []uint32 { //
	var result []uint32
	var bytes []byte
	var I32 uint32
	L := len(buf) / 4
	for I := 0; I < L; I++ {
		bytes = buf[4*I : 4*I+4]
		I32 = BytesToInt32(bytes)
		result = append(result, I32)
	}
	return result
}

func GetWords(Integers []uint16) []byte { //
	var result, buf []byte
	//	L := len(Integers)
	for I := range Integers {
		buf = Int16ToBytes(Integers[I])
		result = append(result, buf[:]...)
	}
	return result
}

//01 00 00 00 02 00 00 00 03 00 00 00 04 00 00 00 05 00 00 00
func PutWords(buf []byte) []uint16 { //
	var result []uint16
	var bytes []byte
	var I16 uint16
	L := len(buf) / 2
	for I := 0; I < L; I++ {
		bytes = buf[2*I : 2*I+2]
		I16 = BytesToInt16(bytes)
		result = append(result, I16)
	}
	return result
}

func ByteArray2Equals(a [2]byte, b [2]byte) bool {
    for i, v := range a {
        if v != b[i] {
            return false
        }
    }
    return true
}

func ByteArray4Equals(a [4]byte, b [4]byte) bool {
    for i, v := range a {
        if v != b[i] {
            return false
        }
    }
    return true
}
