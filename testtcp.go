/* GetHeadInfo
 */
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"plc_h"
	"time"
)

func main() {
	var PLCPtr plc_h.PLC_EtherIP_info
	S := _Connect(&PLCPtr)
	fmt.Println("Response ", S)
	register_session(&PLCPtr)

}
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func _Connect(PLCPtr plc_h.PPLC_EtherIP_info) string { //Send a header - receive a header
	var service string
	//var EHdrBuffer []byte//284]byte
	var verBuffer []byte
	var aConn *net.TCPConn
	var aTCPAdd *net.TCPAddr
	var Cerr error

	request := make([]byte, 24)
	service = "192.168.1.50:44818"
	//  local = "192.168.1.10:49169"

	aTCPAdd, Cerr = net.ResolveTCPAddr("tcp", service)
	// bTCPAdd, Cerr = net.ResolveTCPAddr("tcp",local )

	if Cerr != nil {
		PLCPtr.Error = plc_h.NOHOST
		return "ERROR: resolve " + service
	}

	aConn, Cerr = net.DialTCP("tcp", nil, aTCPAdd)
	if Cerr != nil {
		PLCPtr.Error = plc_h.NOCONNECT
		return "ERROR: Dial " + service
	}

	PLCPtr.Connection = aConn
	Cerr = aConn.SetKeepAlive(true)

	if Cerr != nil {
		PLCPtr.Error = plc_h.TCPERROR
		return "ERROR: Dial " + service
	}

	Cerr = aConn.SetReadBuffer(plc_h.DATA_Buffer_Length)
	if Cerr != nil {
		PLCPtr.Error = plc_h.TCPERROR
		return "ERROR: Set Read Buffer" + service
	}

	EHdrBuffer := []byte{1, plc_h.CONNECT_CMD, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	endBuffer := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	//EHdrBuffer = 1 header.Mode,PCCC_length(2),Connect(4),Status(4)
	verBuffer = Version(plc_h.PCCC_VERSION, plc_h.PCCC_BACKLOG)
	buf := make([]byte, 28, 28)
	buf = Cat2Splices(EHdrBuffer, verBuffer)
	buf = Cat2Splices(buf, endBuffer)

	_, Cerr = aConn.Write(buf)
	if Cerr != nil {
		PLCPtr.Error = plc_h.WRITEERROR
		return "ERROR: Write " + service
	}

	read_len, Rerr := aConn.Read(request)

	if (Rerr != nil) || (read_len != 24) {
		PLCPtr.Error = plc_h.READERROR
		PLCPtr.Connected = 0
		return "ERROR: Not Open " + service
	} else {
		PLCPtr.Error = plc_h.OK
		PLCPtr.Connected = 1
		return "Opened: " + service
	}
}

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

func RandContext() uint64 {
	// create int64 random number
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	Rand64 := r1.Int63()
	return uint64(Rand64)
}

func ContextToByteSlice(Con uint64) [8]byte { //ByteArray0, ByteArray1,...,ByteArray7 byte) uint64 {
	var b [8]byte
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

func ByteSliceToContext(data [8]byte) uint64 {
	return uint64(data[0]) | uint64(data[1])<<8 | uint64(data[2])<<16 | uint64(data[3])<<24 |
		uint64(data[4])<<32 | uint64(data[5])<<40 | uint64(data[6])<<48 | uint64(data[7])<<56
}

func SessionToByteSlice(Con uint32) [4]byte { //ByteArray0, ByteArray1,...,ByteArray7 byte) uint64 {
	var b [4]byte
	b[0] = byte(Con)
	b[1] = byte(Con >> 8)
	b[2] = byte(Con >> 16)
	b[3] = byte(Con >> 24)
	return b
}

func ByteSliceToSession(data [4]byte) uint32 {
	return uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
}

//*************************************************
// Get a session handle from the PLC
//*************************************************
func register_session(PLCPtr plc_h.PPLC_EtherIP_info) string { //, aConn net.TCPConn error
	var dataBuff plc_h.Data_buffer //success = 0, recvBuff
	var aConn *net.TCPConn
	var result string
	var session [4]byte
	request := make([]byte, 24)

	//var res int
	result = "NOTOK"
	aConn = PLCPtr.Connection

	//  if aConn = nil {
	//	return result
	//  }
	PLCPtr.PLC_EtherHdr.EIP_Command = plc_h.Register_Session
	PLCPtr.PLC_EtherHdr.CIP_Len = 4
	PLCPtr.PLC_EtherHdr.Session_handle = 0
	PLCPtr.PLC_EtherHdr.EIP_status = 0
	PLCPtr.PLC_EtherHdr.Context = 54346 //RandContext() //plcutils.RandContext
	PLCPtr.PLC_EtherHdr.Options = 0

	//fill_header(comm, head, debug);
	dataBuff.Data, _ = IPInfoToByteArray(PLCPtr, plc_h.REGLEN) //plcutils.IPInfoToByteArray(PLCPtr)

	//  for I:=0;I<28;I++ {
	//      fmt.Printf("%x ",dataBuff.Data[I])
	//	}
	dataBuff.Data[24] = plc_h.PROTOVERSION
	dataBuff.Data[plc_h.ETHIP_Header_Length] = 1 //* Protocol Version Number */
	dataBuff.Overall_len = plc_h.ETHIP_Header_Length + 4
	_, err := aConn.Write([]byte(dataBuff.Data))
	if err != nil {
		return result
	} else {
		result = "OK"
	}
	//kerr := aConn.SetKeepAlive(true)
	time.Sleep(100 * time.Millisecond)
	read_len, Rerr := aConn.Read(request)
	//res, ber := bufio.NewReader(aConn).Read(request)
	if (Rerr != nil) || (read_len != 24) {
		PLCPtr.Error = plc_h.READERROR
		return "Read Error: " + PLCPtr.PLCHostIP
	}
	copy(session[:], request[4:8])
	PLCPtr.PLC_EtherHdr.Session_handle = ByteSliceToSession(session)
	return "OK"
}

//**************************************************************
// Given a pointer to PLC_Ether_info
// Assemble all of the struc info into a single byte buffer
//**************************************************************
func IPInfoToByteArray(PLCPtr plc_h.PPLC_EtherIP_info, byteLen int) ([]byte, int) {
	var RBuf []byte
	var IDX, ALen, DLen uint16
	var HDR plc_h.EtherIP_Hdr
	var CIPHdr plc_h.CIP_Hdr
	var AddHdr plc_h.Address_Hdr
	var DataHdr plc_h.Data_Hdr

	var AddBuf, DataBuf []byte

	HDRBuf := new(bytes.Buffer)
	CIPHdrBuf := new(bytes.Buffer)
	AddHdrBuf := new(bytes.Buffer)
	DataHdrBuf := new(bytes.Buffer)

	HDR.EIP_Command = PLCPtr.PLC_EtherHdr.EIP_Command
	HDR.CIP_Len = PLCPtr.PLC_EtherHdr.CIP_Len
	HDR.Session_handle = PLCPtr.PLC_EtherHdr.Session_handle
	HDR.Context = PLCPtr.PLC_EtherHdr.Context
	HDR.Options = PLCPtr.PLC_EtherHdr.Options

	CIPHdr.CIPHandle = PLCPtr.PCIP.CIPHdr.CIPHandle
	CIPHdr.CipTimeOut = PLCPtr.PCIP.CIPHdr.CipTimeOut
	CIPHdr.ItemCnt = PLCPtr.PCIP.CIPHdr.ItemCnt

	AddHdr.CSItemType_ID = PLCPtr.PCIP.PAddress.PAddressHdr.CSItemType_ID
	AddHdr.DataLen = PLCPtr.PCIP.PAddress.PAddressHdr.DataLen
	DataHdr.CSItemType_ID = PLCPtr.PCIP.PData.PDataHdr.CSItemType_ID
	DataHdr.DataLen = PLCPtr.PCIP.PData.PDataHdr.DataLen

	ALen = AddHdr.DataLen
	DLen = DataHdr.DataLen

	for IDX = 0; IDX < ALen; IDX++ {
		AddBuf[IDX] = PLCPtr.PCIP.PAddress.ItemData[IDX] //get address buffer data
	}
	for IDX = 0; IDX < DLen; IDX++ {
		DataBuf[IDX] = PLCPtr.PCIP.PData.ItemData[IDX] //get data buffer data
	}

	err := binary.Write(HDRBuf, binary.LittleEndian, HDR)
	if err != nil {
		fmt.Println("binary.Write EtherIP Header:", err)
	}
	err = binary.Write(CIPHdrBuf, binary.LittleEndian, CIPHdr)
	if err != nil {
		fmt.Println("binary.Write EtherIP CIP Header:", err)
	}
	err = binary.Write(AddHdrBuf, binary.LittleEndian, AddHdr)
	if err != nil {
		fmt.Println("binary.Write EtherIP CIP Address header:", err)
	}
	err = binary.Write(DataHdrBuf, binary.LittleEndian, DataHdr)
	if err != nil {
		fmt.Println("binary.Write EtherIP CIP Data header:", err)
	}
	//Append 6 buffers int*H) Bytes
	RBuf = append(HDRBuf.Bytes(), CIPHdrBuf.Bytes()...)
	RBuf = append(RBuf, AddHdrBuf.Bytes()...)
	RBuf = append(RBuf, AddBuf...)
	RBuf = append(RBuf, DataHdrBuf.Bytes()...)
	RBuf = append(RBuf, DataBuf...)
	resBuf := make([]byte, byteLen, byteLen)
	copy(resBuf[:], RBuf) //return a slice of size specified by byteLen parameter
	return resBuf, len(resBuf)
}
