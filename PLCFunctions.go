package PLCFunctions
import (
	"fmt"
	"plc_h"
	"strconv"
	"strings"
	"PLCUtils"
	"time"
	"net"
	"bytes"
	"encoding/binary"
)

const MINLEN = 12
const MINCSD = 4
const CSDLEN = 12 //IF Handle + T/O + Item cnt + Type ID (Address) + Len (Address)
var TNSValue uint16
//func Fill_Logical_Buffer(PLCPtr plc_h.PPLC_EtherIP_info,)
func GetTNS() uint16 {
	TNSValue ++
	return TNSValue 
}

func Connect(PLCPtr plc_h.PPLC_EtherIP_info) string { //S} a header - receive a header
	var service string
	//var EHdrBuffer []byte//284]byte
	var verBuffer []byte
	var aConn *net.TCPConn
	var aTCPAdd *net.TCPAddr
	var Cerr error

	request := make([]byte, plc_h.ENCAPSHDRLEN)
	service = PLCPtr.PLCHostIP + ":" + strconv.Itoa(44818)

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
	verBuffer = PLCUtils.Version(plc_h.PCCC_VERSION, plc_h.PCCC_BACKLOG)
	buf := make([]byte, 28, 28)
	buf = PLCUtils.Cat2Splices(EHdrBuffer, verBuffer)
	buf = PLCUtils.Cat2Splices(buf, endBuffer)

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


//**************************************************************
// Input a string, fill in a FileData structure by side effect *
// T4:1/ACC example                                            *
// Fill in Section, FileNo, FileType, Element and SubElement   *
//**************************************************************
func StrToFileData(FileAddr string, FileData plc_h.PFileData) {
	var I, Slash, Dot int
	//var Bit string
	var _FType, _File, _Elem, Sub, TmpF string
	var FileNum byte
	
	FileData.Data =  make([]byte,0)
	FileData.Section = -1
	FileData.FileNo = 0
	FileData.Element = 0
	FileData.SubElement = 0
	//FileData.Floatdata = plc_h.FALSE

	//------------------------- SLC 5/05 Encoding ----------------------
	FileData.FileNo = 0
	FileData.Element = 0
	FileData.SubElement = 0
	FileData.Section = 0
	FileData.Bit = 0

	TmpF = strings.ToUpper(FileAddr)
	I = PLCUtils.FirstDigit(TmpF)
	_FType = TmpF[0:I] //Everything before first digit
	if I == 0 {
		return
	}
	TmpF = TmpF[I:]   //digits onward
	I = PLCUtils.Delimit(TmpF) //Look for a delimiter
	if I == 0 {
		return
	}
	_File = TmpF[0:I] //from first digit up to delimiter
	JI, _ := strconv.Atoi(_File)
	FileNum = byte(JI)
	TmpF = TmpF[I+1:] //Everything after delimiter
	I = PLCUtils.FirstDigit(TmpF)
	if I < 0 {
		TmpF = ""
	}

	Slash = strings.Index(TmpF, "/")
	Dot = strings.Index(TmpF, ".")
								
	if (Dot > 0) && (Slash > Dot) {
		SL := TmpF[0:Dot]
		SM := TmpF[Dot+1 : Slash]
		SR := TmpF[Slash+1:]
		TmpF = SL + "/" + SM + "." + SR
		Slash = Dot
	}
	if Slash > 0 {
		_Elem = TmpF[0:Slash]
	} else {
		_Elem = TmpF
	}
	Slash = strings.Index(TmpF, "/")
	Dot = strings.Index(TmpF, ".")
	if Dot >= 0 {
		//Bit:=TmpF[Dot:]
		Sub = TmpF[Slash+1 : Dot-Slash]
	} else if Slash > 0 {
		Sub = TmpF[Slash+1:]
	}

	if _Elem != "" {
		JI, _ = strconv.Atoi(_Elem)
		FileData.Element = byte(JI)
	}
	if Sub == "ACC" {
		FileData.SubElement = 2
	} else if Sub == "PRE" {
		FileData.SubElement = 1
	} else if Sub == "LEN" {
		FileData.SubElement = 1
	} else if Sub == "POS" {
		FileData.SubElement = 2
	} else {
		FileData.SubElement = 0
	}
	FileData.TypeLen = 2 //default
	FileData.FileNo = FileNum

	if _FType == plc_h.OUTPUT {
		FileData.FileType = plc_h.OUTPUT_TYPE
		FileData.FileNo = 0 //zero
	} else if _FType == plc_h.INPUT {
		FileData.FileType = plc_h.INPUT_TYPE
		FileData.FileNo = 1 //one
	} else if _FType == plc_h.STATUS {
		FileData.FileNo = 0
		FileData.FileType = plc_h.STATUS_TYPE
	} else if _FType == plc_h.STRING {
		FileData.FileNo = 0
		FileData.FileType = plc_h.STRING_TYPE
		//FileData.TypeLen = 0x54
	} else if _FType == plc_h.BINARY {
		FileData.FileType = plc_h.BIT_TYPE
	} else if _FType == plc_h.TIMER {
		FileData.FileType = plc_h.TIMER_TYPE
	} else if _FType ==  plc_h.COUNTER {
		FileData.FileType = plc_h.COUNTER_TYPE
	} else if _FType ==  plc_h.CONTROL {
		FileData.FileType = plc_h.CONTROL_TYPE
	} else if _FType ==  plc_h.INTEGER {
		FileData.FileType = plc_h.INTEGER_TYPE
	} else if _FType ==  plc_h.FLOAT {
		FileData.FileType = plc_h.FLOAT_TYPE
		//FileData.Floatdata = plc_h.TRUE
		FileData.TypeLen = 4
	} else if _FType == plc_h.ASCII {
		FileData.FileType = plc_h.ASCII_TYPE
		FileData.Bit = FileData.SubElement
		FileData.SubElement = 0
		FileData.TypeLen = len(_File)
	} else if _FType == plc_h.BCD {
		FileData.FileType = plc_h.BCD_TYPE
	} else if _FType == "P" {
		FileData.Section = 1
		FileData.FileNo = 7
		FileData.Element = 0
	}
return
}


//***********************************************
// Convert a string into a FileData structure   *
//***********************************************
func FileStrToFileData(FileAddr string, FileData plc_h.PFileData) {
	var x int
	var prefix, suffix string
	//var  tempFileData [3]string

	FileData.Section = -1
	FileData.FileNo = 0
	FileData.Element = 0
	FileData.SubElement = 0
	//FileData.Floatdata = plc_h.FALSE
	// tempFileData        = ""

	//------------------------- SLC 5/05 Encoding ----------------------
	FileData.FileNo = 0
	FileData.Element = 0
	FileData.SubElement = 0
	FileData.Section = 0
	suffix = ""
	prefix = string(FileAddr[0])

	for x = 1; x <= len(FileAddr); x++ {
		if isDigit(FileAddr[x]) {
			suffix = suffix + string(FileAddr[x])
		} else {
			break
		}
	}

	I, ERR := strconv.Atoi(suffix)
	fmt.Println(ERR, suffix)

	FileData.FileNo = byte(I)

	if prefix == "O" {
		FileData.FileType = plc_h.OUTPUT_TYPE
		FileData.TypeLen = 2
	} else if prefix == "I" {
		FileData.FileType = plc_h.INPUT_TYPE
		FileData.TypeLen = 2
	} else if prefix == "S" {
		FileData.FileType = plc_h.STATUS_TYPE
		FileData.TypeLen = 2
	} else if prefix == "B" {
		//inc(x);
		FileData.FileType = plc_h.BIT_TYPE
		FileData.TypeLen = 2
	} else if prefix == "T" {
		FileData.FileType = plc_h.TIMER_TYPE
		FileData.TypeLen = 2
	} else if prefix == "C" {
		FileData.FileType = plc_h.COUNTER_TYPE
		FileData.TypeLen = 2
	} else if prefix == "R" {
		FileData.FileType = plc_h.CONTROL_TYPE
		FileData.TypeLen = 2
	} else if prefix == "N" {
		FileData.FileType = plc_h.INTEGER_TYPE
		FileData.TypeLen = 2
	} else if prefix == "F" {
		FileData.FileType = plc_h.FLOAT_TYPE
		//FileData.Floatdata = plc_h.TRUE
		FileData.TypeLen = 4
	} else if prefix == "A" {
		//  inc(x);
		FileData.FileType = plc_h.ASCII_TYPE
		FileData.TypeLen = 1
	} else if prefix == "D" {
		FileData.FileType = plc_h.BCD_TYPE
		FileData.TypeLen = 2
	} else if prefix == "P" { //special case to read program FileData from PLC.
		FileData.Section = 1
		FileData.FileNo = 7
		FileData.Element = 0
	}

	//fmt.Println("Section ", FileData.Section)
	//fmt.Println("Element ", FileData.Element)
	//fmt.Println("Sub Element ", FileData.SubElement)
	////fmt.Printf("FType %x", FileData.FileType)
	//fmt.Println(" ")
	//fmt.Println("Type Len ", FileData.TypeLen)
	//fmt.Println("Bit ", FileData.Bit)
	//fmt.Println("Length ", FileData.Length)

	return
}

func isDigit(C byte) bool {
	if (string(C) >= "0") && (string(C) <= "9") {
		return true
	} else {
		return false
	}
}

//**************************************************************
// Given a single byte buffer                                  *
// transfer the buffer data into a single PCCCReply struct     *
//**************************************************************
func ByteArrayToReply(PLCReply plc_h.PPCCCReply, ByteBuf []byte) bool {
	//var	AItem    plc_h.Address_Item
	//var	DItem    plc_h.Data_Item
	var IDX int
	var DataLen, AddLen uint16
	var Command uint16
	var status uint32
	//var CSDLen uint16

	//Get lengths of address data and data data buffers
	AddLen = PLCUtils.ByteToUint16(ByteBuf[34:36])
	DataLen = PLCUtils.ByteToUint16(ByteBuf[38+AddLen : 38+AddLen+2])
	Command = PLCUtils.ByteToUint16(ByteBuf[0:2])
	//CSDLen  = uint16(PLCUtils.BytesToInt16(ByteBuf[2:4]))
	if len(ByteBuf) < plc_h.ENCAPSHDRLEN + plc_h.CSDLEN { 
 	  PLCReply.Error  = plc_h.FALSE             // no data exists
	  return false	
	}
				
	IDX = plc_h.ENCAPSHDRLEN  + plc_h.CSDLEN
	DataBuf := ByteBuf[IDX:IDX + int(DataLen)]

	PLCReply.Cmd     = Command
	PLCReply.Length  = DataLen
	copy(PLCReply.Session_handle[:], ByteBuf[4:8])  
	copy(PLCReply.Status[:],ByteBuf[8:12])        
	copy(PLCReply.Context[:],ByteBuf[12:20])
	PLCReply.Answer  = DataBuf
	PLCReply.Error   = 0
	status = PLCUtils.ByteArrayToUint32(PLCReply.Status)
	if status != 0 {
		return false
	} else {
		return true
	}

}

func ParseStatus(PCCC plc_h.PPCCCReply) (string) {
	var SerRev byte
	var Series = ""
	var Revision = ""
	var Name = ""
    var status uint32
	status = PLCUtils.ByteArrayToUint32(PCCC.Status)
	if (PCCC.Error != 0) || (status != 0) {
		return " Status request failed."
	}
	SerRev = PCCC.Answer[8]
	Series = "Series = "+strconv.Itoa(int(1 + (SerRev&240)>>4))
	Revision = "Rev = "+strconv.Itoa(int(SerRev&15) + 64)
	Name = "Name = "+string(PCCC.Answer[9:16])
	return Name+", "+Series+", "+Revision
}

//*************************************************
// Get a diagnostic status from the PLC
//*************************************************
func Get_Status(PLCPtr plc_h.PPLC_EtherIP_info) (string, string) {
	var dataBuff plc_h.Data_buffer //success = 0, recvBuff
	var aConn *net.TCPConn
    var PLCReply plc_h.PCCCReply
	var s string
	
	request := make([]byte, 58+plc_h.ENCAPSHDRLEN)
	aConn = PLCPtr.Connection
	PLCPtr.PLC_EtherHdr.EIP_status = plc_h.Blank32
	Fill_CS_Address(PLCPtr, plc_h.STATTYPE, plc_h.DIAG_STATUS_CMD)
	Fill_CS_DataHdr(PLCPtr, plc_h.DIAG_STATUS_CMD, plc_h.DIAG_STATUS_FNC)
	dataBuff.Data, _ = IPInfoToByteArray(PLCPtr, 34+plc_h.ENCAPSHDRLEN) //plcutils.IPInfoToByteArray(PLCPtr)
	dataBuff.Overall_len = plc_h.ETHIP_Header_Length + 4
	_, err := aConn.Write([]byte(dataBuff.Data))
	if err != nil {
		return "NOTOK", ""
	}
	time.Sleep(100 * time.Millisecond)
	_, Rerr := aConn.Read(request)
	if (Rerr != nil) {
		PLCPtr.Error = plc_h.READERROR
		return "Read Error: " + PLCPtr.PLCHostIP, ""
	}
	_ = ByteArrayToReply(&PLCReply, request) 
	s = ParseStatus(&PLCReply)
	return "OK", s
}

//******************************************************************************************
// Given a pointer to PLC_Ether_info                                                       *
// Assemble all of the struct info into a single byte buffer                               *
// byteLen is number of bytes past function Cmd + Sts+ TNS + Fnc is common to all commands *
//******************************************************************************************
func IPInfoToByteArray(PLCPtr plc_h.PPLC_EtherIP_info, byteLen int) ([]byte, int) {
	var RBuf []byte
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

	AddHdr.CSItemType_ID = PLCPtr.PCIP.Address.AddressHdr.CSItemType_ID
	AddHdr.DataLen = PLCPtr.PCIP.Address.AddressHdr.DataLen
	DataHdr.CSItemType_ID = PLCPtr.PCIP.Data.DataHdr.CSItemType_ID
	DataHdr.DataLen = PLCPtr.PCIP.Data.DataHdr.DataLen
	//	DataHdr.Cmd = PLCPtr.PCIP.Data.DataHdr.Cmd
	//	DataHdr.Sts = PLCPtr.PCIP.Data.DataHdr.Sts
	//	DataHdr.Tns = PLCPtr.PCIP.Data.DataHdr.Tns
	//	DataHdr.Fnc = PLCPtr.PCIP.Data.DataHdr.Fnc

	AddBuf = append(AddBuf, PLCPtr.PCIP.Address.ItemData[:]...)
	DataBuf = append(DataBuf, PLCPtr.PCIP.Data.ItemData[:]...)
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
	//App} 6 buffers int*H) Bytes
	RBuf = append(HDRBuf.Bytes(), CIPHdrBuf.Bytes()...)
	RBuf = append(RBuf, AddHdrBuf.Bytes()...)
	RBuf = append(RBuf, AddBuf...)
	RBuf = append(RBuf, DataHdrBuf.Bytes()...)
	RBuf = append(RBuf, DataBuf...)

	resBuf := make([]byte, byteLen, byteLen)
	copy(resBuf[:], RBuf) //return a slice of size specified by byteLen parameter
	return resBuf, len(resBuf)
}

//*************************************************
// Replyt to Register_Session from PLC            *
//*************************************************
func ReplyRegSession(PLCPtr plc_h.PPLC_EtherIP_info) string { //, aConn net.TCPConn error
	var dataBuff plc_h.Data_buffer //success = 0, recvBuff
	var aConn *net.TCPConn
//	var Cont [8]byte
	//session := make([]byte, 4)
	//var res int
	aConn = PLCPtr.Connection

	//  if aConn = nil {
	//	return result
	//  }

	PLCPtr.PLC_EtherHdr.EIP_Command = plc_h.Register_Session
	PLCPtr.PLC_EtherHdr.CIP_Len = 4
	PLCPtr.PLC_EtherHdr.Session_handle = PLCUtils.RandSession()
	PLCPtr.PLC_EtherHdr.EIP_status = plc_h.Blank32
	//Cont = PLCUtils.Uint64ToByteSlice(PLCUtils.RandContext())
	PLCPtr.PLC_EtherHdr.Context = PLCUtils.RandContext()
	PLCPtr.PLC_EtherHdr.Options = plc_h.Blank32
	//fill_header(comm, head, debug);
	dataBuff.Data, _ = IPInfoToByteArray(PLCPtr, plc_h.REGLEN) //plcutils.IPInfoToByteArray(PLCPtr)
	dataBuff.Data[24] = plc_h.PROTOVERSION
	dataBuff.Data[plc_h.ETHIP_Header_Length] = 1 //* Protocol Version Number */
	dataBuff.Overall_len = plc_h.ETHIP_Header_Length + 4
	_, err := aConn.Write([]byte(dataBuff.Data))
	if err != nil {
		return "NOTOK"
	} else {
		return "OK"
	}
}	
	
//*************************************************
// Get a session handle from the PLC
//*************************************************
func Register_session(PLCPtr plc_h.PPLC_EtherIP_info) string { //, aConn net.TCPConn error
	var dataBuff plc_h.Data_buffer //success = 0, recvBuff
	var aConn *net.TCPConn
	var result string
	request := make([]byte, 28)
	var session  [4]byte
	//var res int
	result = "NOTOK"
	aConn = PLCPtr.Connection

	//  if aConn = nil {
	//	return result
	//  }
	PLCPtr.PLC_EtherHdr.EIP_Command = plc_h.Register_Session
	PLCPtr.PLC_EtherHdr.CIP_Len = 4
	PLCPtr.PLC_EtherHdr.Session_handle = plc_h.Blank32
	PLCPtr.PLC_EtherHdr.EIP_status = plc_h.Blank32
	PLCPtr.PLC_EtherHdr.Context = PLCUtils.RandContext() //RandContext() //plcutils.RandContext
	PLCPtr.PLC_EtherHdr.Options = plc_h.Blank32
	//fill_header(comm, head, debug);
	dataBuff.Data, _ = IPInfoToByteArray(PLCPtr, plc_h.REGLEN) //plcutils.IPInfoToByteArray(PLCPtr)
	dataBuff.Data[24] = plc_h.PROTOVERSION
	dataBuff.Data[plc_h.ETHIP_Header_Length] = 1 //* Protocol Version Number */
	dataBuff.Overall_len = plc_h.ETHIP_Header_Length + 4
	fmt.Println()
	fmt.Printf("Data % x ",dataBuff.Data)
	_, err := aConn.Write([]byte(dataBuff.Data))
	if err != nil {
		return result
	} else {
		result = "OK"
	}
	//kerr := aConn.SetKeepAlive(true)
	time.Sleep(100 * time.Millisecond)
	read_len, Rerr := aConn.Read(request)

	if (Rerr != nil) || (read_len != 28) {
		PLCPtr.Error = plc_h.READERROR
		return "Read Error: " + PLCPtr.PLCHostIP
	}
	copy(session[:], request[4:8])
	PLCPtr.PLC_EtherHdr.Session_handle = session
	return "OK"
}

func UnRegister_session(PLCPtr plc_h.PPLC_EtherIP_info) string { //, aConn net.TCPConn error
 var dataBuff plc_h.Data_buffer //success = 0, recvBuff
	var aConn *net.TCPConn
	var result string
	request := make([]byte, 28)
	//var res int
	result = "NOTOK"
	aConn = PLCPtr.Connection

	//  if aConn = nil {
	//	return result
	//  }
	PLCPtr.PLC_EtherHdr.EIP_Command = plc_h.UnRegister_Session
	PLCPtr.PLC_EtherHdr.CIP_Len = 4
	PLCPtr.PLC_EtherHdr.EIP_status = plc_h.Blank32
	PLCPtr.PLC_EtherHdr.Context = PLCUtils.RandContext() //RandContext() //plcutils.RandContext
	PLCPtr.PLC_EtherHdr.Options = plc_h.Blank32
	//fill_header(comm, head, debug);
	dataBuff.Data, _ = IPInfoToByteArray(PLCPtr, plc_h.REGLEN) //plcutils.IPInfoToByteArray(PLCPtr)
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

	if (Rerr != nil) || (read_len != 28) {
		PLCPtr.Error = plc_h.READERROR
		return "Read Error: " + PLCPtr.PLCHostIP
	}
	return "OK"
	}

//**********************************************************************
// Command specific data - address portion                             *
// Put 2 byte command + length of address portion + plc ip address     *
// ip address as xxx.xxx.xxx.xxx or fewer bytes                        *
// ip address may be padded with trailing zero for an even no of bytes *
// PLCPtr - sent as an address - return data len, data len + min len   *
//**********************************************************************
func Fill_CS_Address(PLCPtr plc_h.PPLC_EtherIP_info, CSAddress_Type uint16, Cmd byte) {
	var IPlen uint16
	var IPAddByte []byte
    var AddData = []byte{1} 
	PLCPtr.PCIP.CIPHdr.CipTimeOut = 1024 //Added check it out
	PLCPtr.PCIP.CIPHdr.ItemCnt  = 2
    PLCPtr.PCIP.Address.AddressHdr.CSItemType_ID = CSAddress_Type
    PLCPtr.PCIP.Address.AddressHdr.DataLen = 1
	
	PLCPtr.PCIP.Address.ItemData = AddData
	if Cmd == plc_h.DIAG_STATUS_CMD {
		IPAddByte = []byte(PLCPtr.PLCHostIP) //PLC IP Address
		PLCPtr.PCIP.Address.ItemData = append(PLCPtr.PCIP.Address.ItemData, IPAddByte[:]...)
		IPlen = uint16(len(IPAddByte)) //length of PLC IP Address in chars (bytes)
		if IPlen%2 == 0 {              //pad len to odd number
			IPlen++
			S := []byte{0}
			PLCPtr.PCIP.Address.ItemData = append(PLCPtr.PCIP.Address.ItemData, S[0]) //pad with 0
		} else { //CM!= DIAG_STATUS_CMD //just set data = to 1
			IPlen = 1
			S := []byte{1}
			PLCPtr.PCIP.Address.ItemData = append(PLCPtr.PCIP.Address.ItemData, S[0]) // or insert 1
		}

		PLCPtr.PLC_EtherHdr.CIP_Len = MINLEN + IPlen
		PLCPtr.PCIP.Address.AddressHdr.DataLen = IPlen //Length of PLC IP Address
	} else if CSAddress_Type == plc_h.NULLADDRESS {
		PLCPtr.PLC_EtherHdr.CIP_Len = MINLEN 
		PLCPtr.PCIP.Address.ItemData = make([]byte,0)
		PLCPtr.PCIP.Address.AddressHdr.DataLen = 0 //Length of PLC IP Address
	} else {
		PLCPtr.PLC_EtherHdr.CIP_Len = MINLEN + 1
		PLCPtr.PCIP.Address.AddressHdr.DataLen = 1 //Length of PLC IP Address
	}
}

//**********************************************************************
// Command specific data - data portion                                *
// Put 2 byte command + length of data portion + data header           *
// + common portion of ItemData - Cmd, Status, TNS, Fnc                * 
//**********************************************************************
func Fill_CS_DataHdr(PLCPtr plc_h.PPLC_EtherIP_info,  Cmd_type, Fnc byte) {
	const DMINLEN uint16 = 5 //size of CSItemType_ID + size of dataLen (Data_Item)
	var data []byte
	// Fill out header

	switch Cmd_type {
	case plc_h.GEN_FILE_CMD:
		PLCPtr.PCIP.Data.DataHdr.CSItemType_ID = plc_h.RRDATATYPE //word  - don't count this
	case plc_h.CIF_READ_CMD:
		PLCPtr.PCIP.Data.DataHdr.CSItemType_ID = plc_h.RRDATATYPE //word  - don't count this
	case plc_h.CIF_WRITE_CMD:
		PLCPtr.PCIP.Data.DataHdr.CSItemType_ID = plc_h.RRDATATYPE //word  - don't count this
	case plc_h.DIAG_STATUS_CMD:
		{
			PLCPtr.PCIP.Data.DataHdr.CSItemType_ID = plc_h.STATTYPE
			//	PLCPtr.PCIP.Data.DataHdr.Cmd = plc_h.FLSTATUS
		}
	}
	//len = size of sts+ size of cmd + size of TNS
	PLCPtr.PCIP.Data.DataHdr.DataLen = DMINLEN //word  don't count total 1 word + 2 bytes = 4 bytes
	data = append(data, Cmd_type)
	data = append(data, 0)
	Junk := PLCUtils.Int16ToBytes(GetTNS())
	data = append(data,Junk[:]...)
	data = append(data,Fnc)
	PLCPtr.PLC_EtherHdr.CIP_Len += 9    //CIP bytes up to & including data length
	PLCPtr.PCIP.Data.ItemData = data
	fmt.Printf ("DATA % x ",PLCPtr.PCIP.Data.DataHdr)
	fmt.Println()
	fmt.Printf ("DATA % x ",PLCPtr.PCIP.Data.ItemData)
	return
}

//**************************************************************
// Given a single byte buffer
// transfer the buffer data into a single PLC_Ether_info struct
//**************************************************************
func ByteArrayToIPInfo(PLCPtr plc_h.PPLC_EtherIP_info, ByteBuf []byte) bool {
	//var	AItem    plc_h.Address_Item
	//var	DItem    plc_h.Data_Item
	var EIPHead plc_h.EtherIP_Hdr
	var CIPHead plc_h.CIP_Hdr
	var AddHdr plc_h.Address_Hdr
	var DataHdr plc_h.Data_Hdr
	var AddLen, DataLen, DataPtr uint16
	//Get lengths of address data and data data buffers
	AddLen = PLCUtils.ByteToUint16(ByteBuf[34:36])
	DataLen = PLCUtils.ByteToUint16(ByteBuf[38+AddLen : 38+AddLen+2])

	// EIP_Command, CIP_Len, Session_Handle, EIP_Status, Context, Options - 24 bytes
	EIPHeadSlice := bytes.NewReader(ByteBuf[0:23])
	// CIP_Handle, CipTimeOut, ItemCnt - 8 bytes
	CIPHeadSlice := bytes.NewReader(ByteBuf[24:31])
	// CSItemType_ID, DataLen - 4 bytes - address header
	AddHeadSlice := bytes.NewReader(ByteBuf[32:35])
	// Address Item data - number of bytes = DataLen
	AItemSlice := ByteBuf[36 : 36+AddLen]
	DataPtr = 36 + AddLen + 4

	DItemSlice := ByteBuf[DataPtr : DataPtr+DataLen]

	// CSItemType_ID, DataLen - 4 bytes - data header
	DataHeadSlice := bytes.NewReader(ByteBuf[DataLen : DataLen+8])
	// DItemSlice := ByteBuf[AddLen : (AddLen + 8 + DataLen -1)]

	err1 := binary.Read(EIPHeadSlice, binary.LittleEndian, EIPHead) //
	err2 := binary.Read(CIPHeadSlice, binary.LittleEndian, CIPHead)
	err3 := binary.Read(AddHeadSlice, binary.LittleEndian, AddHdr)   // 4 bytes
	err4 := binary.Read(DataHeadSlice, binary.LittleEndian, DataHdr) // 8 bytes
	fmt.Printf("Lengths %x", string(DItemSlice))
	return false
	//Now we have EtherIP_Hdr data plus CIP data
	PLCPtr.PLC_EtherHdr = EIPHead
	PLCPtr.PCIP.CIPHdr = CIPHead
	PLCPtr.PCIP.Address.AddressHdr = AddHdr
	PLCPtr.PCIP.Address.ItemData = AItemSlice

	PLCPtr.PCIP.Data.DataHdr = DataHdr
	PLCPtr.PCIP.Data.ItemData = DItemSlice
	if err1 != nil {
		return false
	}
	if err2 != nil {
		return false
	}
	if err3 != nil {
		return false
	}
	if err4 != nil {
		return false
	}
	return true
}

//***********************************************
// Convert a string into a FileData structure   *
//***********************************************
func fileStrToFileData(FileAddr string, FileData plc_h.PFileData) {
	var x int
	var prefix, suffix string
	//var  tempFileData [3]string

	FileData.Section = -1
	FileData.FileNo = 0
	FileData.Element = 0
	FileData.SubElement = 0
	//FileData.Floatdata = plc_h.FALSE
	// tempFileData        = ""

	//------------------------- SLC 5/05 Encoding ----------------------
	FileData.FileNo = 0
	FileData.Element = 0
	FileData.SubElement = 0
	FileData.Section = 0
	suffix = ""
	prefix = string(FileAddr[0])

	for x = 1; x <= len(FileAddr); x++ {
		if PLCUtils.IsDigit(FileAddr[x]) {
			suffix = suffix + string(FileAddr[x])
		} else {
			break
		}
	}

	I, ERR := strconv.Atoi(suffix)
	fmt.Println(ERR, suffix)

	FileData.FileNo = byte(I)

	if prefix == "O" {
		FileData.FileType = plc_h.OUTPUT_TYPE
		FileData.TypeLen = 2
	} else if prefix == "I" {
		FileData.FileType = plc_h.INPUT_TYPE
		FileData.TypeLen = 2
	} else if prefix == "S" {
		FileData.FileType = plc_h.STATUS_TYPE
		FileData.TypeLen = 2
	} else if prefix == "B" {
		//inc(x);
		FileData.FileType = plc_h.BIT_TYPE
		FileData.TypeLen = 2
	} else if prefix == "T" {
		FileData.FileType = plc_h.TIMER_TYPE
		FileData.TypeLen = 2
	} else if prefix == "C" {
		FileData.FileType = plc_h.COUNTER_TYPE
		FileData.TypeLen = 2
	} else if prefix == "R" {
		FileData.FileType = plc_h.CONTROL_TYPE
		FileData.TypeLen = 2
	} else if prefix == "N" {
		FileData.FileType = plc_h.INTEGER_TYPE
		FileData.TypeLen = 2
	} else if prefix == "F" {
		FileData.FileType = plc_h.FLOAT_TYPE
		//FileData.Floatdata = plc_h.TRUE
		FileData.TypeLen = 4
	} else if prefix == "A" {
		//  inc(x);
		FileData.FileType = plc_h.ASCII_TYPE
		FileData.TypeLen = 1
	} else if prefix == "D" {
		FileData.FileType = plc_h.BCD_TYPE
		FileData.TypeLen = 2
	} else if prefix == "P" { //special case to read program FileData from PLC.
		FileData.Section = 1
		FileData.FileNo = 7
		FileData.Element = 0
	}

	//fmt.Println("Section ", FileData.Section)
	//fmt.Println("Element ", FileData.Element)
	//fmt.Println("Sub Element ", FileData.SubElement)
	//fmt.Printf("FType %x", FileData.FileType)
	//fmt.Println(" ")
	//fmt.Println("Type Len ", FileData.TypeLen)
	//fmt.Println("Bit ", FileData.Bit)
	//fmt.Println("Length ", FileData.Length)

	return
}

//*******************************************************************************************
//  Data returned by PLC from a Protected typed FILE read/write                             *
//  DBytes is the byte sequence returned by  the PLC                                        *
//  byte sequence is interpreted and returned in FData - a File Data Structure see plc_h.go *
//*******************************************************************************************
func TypedFileGet (FData plc_h.PFileData, DBytes []byte) int {
 var NumElements, IDX, I, front int
 var Cmd byte
 var TBytes []byte
	 Cmd = FData.PutCmd
	 if Cmd == plc_h.TYPE_FILE_READ_FNC {
		front = 4                           //CMD,STS,TNS(2)
		FData.EXStatus = 0
	} else if Cmd == plc_h.TYPE_FILE_WRITE_FNC {
		front = 5                           //CMD,STS,TNS(2),EXT STS   
		FData.EXStatus = DBytes[4]                    
	} else {
	  return plc_h.FALSE
	}	
	
	FData.Length = byte(len(DBytes))

    TBytes         = DBytes[front:]	
	FData.Data     = TBytes
	FData.Size     = byte(len(TBytes))
	FData.GetCmd   = DBytes[0]
	FData.Status   = DBytes[1]
	FData.TNS      = PLCUtils.BytesToInt16(DBytes[2:3])
    if FData.FileType == plc_h.FLOAT_TYPE {
	   NumElements      = int(FData.Size / plc_h.FLOATLEN)
	   FData.ByteData   = make([]byte,0)
	   FData.WordData   = make([]uint16, 0)
	   FData.FloatData  = make([]float32, NumElements)	
	   for I = 0; I < NumElements; I++ {
		   IDX   = I*plc_h.FLOATLEN	
		   Junk := PLCUtils.BytesToFloat32(TBytes[IDX:IDX+plc_h.FLOATLEN])
		   FData.FloatData = append(FData.FloatData,Junk ) 
	    } // for	
    } else if (FData.FileType != plc_h.STATUS_TYPE) && (FData.FileType != plc_h.STRING_TYPE){
		NumElements      = int(FData.Size / plc_h.WORDLEN)
		FData.ByteData   = make([]byte,0)
		FData.FloatData  = make([]float32, 0)
	    FData.WordData   = make([]uint16, NumElements)
		for I = 0; I < NumElements; I++ {
		   IDX   = I*plc_h.WORDLEN	
		   Junk := PLCUtils.BytesToInt16(TBytes[IDX:IDX+plc_h.WORDLEN])
		   FData.WordData = append(FData.WordData,Junk ) 
	    } //for
	} else {  //Interpret data as bytes ( for now - how about strings?)
	    NumElements      = int(FData.Size)             
	    FData.WordData   = make([]uint16, 0)
	    FData.FloatData  = make([]float32,0)
		FData.ByteData   = make([]byte,NumElements)
		FData.ByteData   = TBytes
		FData.Data       = TBytes  
	}//else	
	return plc_h.TRUE
} // function

//****************************************************
// Write the Typed File command - receive the answer * 
//****************************************************
func TypedFile(PLCPtr plc_h.PPLC_EtherIP_info, FData plc_h.PFileData) (string, string, error) {
	var aConn *net.TCPConn
	var dataBuff plc_h.Data_buffer //success = 0, recvBuff
    var PLCReply plc_h.PCCCReply
	var s string
	//PLCPtr.PCIP.Data.ItemData = FData.Data 
	//PLCPtr.PCIP.Data.DataHdr.DataLen = uint16(FData.Size+11)               // 11 = Number of bytes in Typed file read/write sans data
    PLCPtr.PLC_EtherHdr.CIP_Len = uint16(len(PLCPtr.PCIP.Data.ItemData))+plc_h.CSDLEN
	aConn = PLCPtr.Connection
	fmt.Println("Sizes ",FData.Size,PLCPtr.PCIP.Data.DataHdr.DataLen)
	request := make([]byte, plc_h.ENCAPSHDRLEN + plc_h.CSDLEN + FData.Size + 4)
    enCapSize :=plc_h.ENCAPSHDRLEN + plc_h.CSDLEN + int(PLCPtr.PCIP.Data.DataHdr.DataLen)
	dataBuff.Data, _ = IPInfoToByteArray(PLCPtr, enCapSize) //plcutils.IPInfoToByteArray(PLCPtr)
			fmt.Printf("Encaps % v ",enCapSize)
	dataBuff.Overall_len = plc_h.ETHIP_Header_Length + 4
	_, err := aConn.Write([]byte(dataBuff.Data))
	if err != nil {
		return "NOTOK", "", err
	}
	time.Sleep(100 * time.Millisecond)
	_, Rerr := aConn.Read(request)
	if (Rerr != nil) {
		PLCPtr.Error = plc_h.READERROR
		return "Read Error: " + PLCPtr.PLCHostIP, "", Rerr
	}
	_ = ByteArrayToReply(&PLCReply, request) 
	FData.Data = PLCReply.Answer[4:]
 
	return "OK", s, nil
}

//************************************************************************************************************************
//  Protected typed FILE read/write   - from/to an open file                                                             *
//  Elements denotes the number of elements, i.e. Floats, 16 bit Words, 8 bit bytes                                      *
//  Element data to write will be in a FileData structure, FloatData, WordData or ByteData                               *
//  RW = "READ" denotes a read operation, else Write                                                                     *
//************************************************************************************************************************
func TypedFilePut (PLCPtr plc_h.PPLC_EtherIP_info,FData plc_h.PFileData,Elements int, RW string) {
  const MINBYTES = 11
  var NumElements int
  var TmpData []byte

	if RW == "READ" {
	   FData.ByteData   = make([]byte,0)
	   FData.WordData   = make([]uint16, 0)
	   FData.FloatData  = make([]float32, 0)	
	   FData.PutCmd = plc_h.TYPE_FILE_READ_CMD 
	   FData.Function = plc_h.TYPE_FILE_READ_FNC
	   if FData.FileType == plc_h.FLOAT_TYPE {
		   FData.Size = byte(Elements * plc_h.FLOATLEN)
	   } else if (FData.FileType != plc_h.STATUS_TYPE) && (FData.FileType != plc_h.STRING_TYPE)	{
		   FData.Size = byte(Elements * plc_h.WORDLEN)
	   } else {
	     FData.Size = byte(NumElements)
	   } 
	} else {    // Set size of data in bytes
	    if (FData.FileType == plc_h.FLOAT_TYPE)	{
			FData.ByteData   = make([]byte,0)
	        FData.WordData   = make([]uint16, 0)
			NumElements = len(FData.FloatData) 
			FData.Size = byte(NumElements * plc_h.FLOATLEN)
			for I := 0; I < Elements; I ++ {
		      Junk := PLCUtils.Float32ToBytes(FData.FloatData[I])
			  TmpData = append(TmpData,Junk[:]...)
		      }
	    } else if (FData.FileType != plc_h.STATUS_TYPE) && (FData.FileType != plc_h.STRING_TYPE)	{
			NumElements = len(FData.WordData) / plc_h.WORDLEN
		    FData.Size = byte(NumElements * plc_h.WORDLEN)
			for I := 0; I < Elements; I ++ {
		      Junk := PLCUtils.Int16ToBytes(FData.WordData[I * plc_h.WORDLEN])
		      TmpData = append(TmpData,Junk[:]...)
		      }
	    } else {
			FData.WordData   = make([]uint16, 0)
	        FData.FloatData  = make([]float32,0)
			TmpData          = FData.ByteData
		    FData.Size       = byte(len(FData.ByteData))
		} 
	    FData.PutCmd   = plc_h.TYPE_FILE_WRITE_CMD
	    FData.Function = plc_h.TYPE_FILE_WRITE_FNC 
	}
	FData.Status   = 0
	FData.TNS      = GetTNS()
	FData.Data = append(FData.Data,FData.PutCmd)         //Cmd, Status, TNS(2), Fnc, Size, Tag(2), Offset(2), Type
	FData.Data = append(FData.Data,FData.Status)
	Junk := PLCUtils.Int16ToBytes(FData.TNS)
	FData.Data = append(FData.Data,Junk[:]...)
	FData.Data = append(FData.Data,FData.Function)
	FData.Data = append(FData.Data,FData.Size)           //Size in bytes for data buffer which (contains only element values in byte form)
	Junk  = PLCUtils.Int16ToBytes(FData.Tag)
	FData.Data = append(FData.Data,Junk[:]...)
	Junk  = PLCUtils.Int16ToBytes(FData.Offset)
	FData.Data = append(FData.Data,Junk[:]...)
	FData.Data = append(FData.Data,FData.FileType)
	FData.Data = append(FData.Data,TmpData[:]...)
	if RW == "READ" {
	  PLCPtr.PCIP.Data.DataHdr.DataLen = uint16(MINBYTES)  
	} else {
	  PLCPtr.PCIP.Data.DataHdr.DataLen = uint16(FData.Size+MINBYTES)               // 11 = Number of bytes in Typed file read/write sans data
	}
	PLCPtr.PCIP.Data.ItemData = FData.Data 
    _,_,_ = PutData(PLCPtr,FData, RW)
}


//************************************************************************************************************************
//  Protected typed LOGICAL read/write                                                                                   *
//  FType ex: 86 for Timer, FNum ex: 4 for T4, ElemNum ex: 2 for T4:2, SubNum ex:PRE for T4:1/PRE                        *
//  Element data to write will be in a FileData structure, FloatData, WprdData or ByteData                               *
//  RW = "READ" denotes a read operation, else Write                                                                     *
//  Alter a file data structure for DataItem data                                                                        *
//  A buffer combining Encapsulation Header + Command Specific Data forms a buffer sent to PLC via socket                *
//************************************************************************************************************************
func LogicalPut(PLCPtr plc_h.PPLC_EtherIP_info,FData plc_h.PFileData, Element string, Elements int, RW string) {
	var NumElements int
	StrToFileData(Element, FData)
	if RW == "READ" {                                  // Number of elements requested determines size
	   FData.PutCmd   = plc_h.LOGICAL_READ_CMD 
	   FData.Function = plc_h.LOGICAL_READ_FNC
	   if FData.FileType == plc_h.FLOAT_TYPE {
		   FData.Size = byte(Elements * plc_h.FLOATLEN)
	   } else if (FData.FileType != plc_h.STATUS_TYPE) && (FData.FileType != plc_h.STRING_TYPE)	{
		   FData.Size = byte(Elements * plc_h.WORDLEN)
	   } else {
	     FData.Size = byte(Elements)
	   } 
	} else {                                           //Given data determines size 
	   FData.PutCmd   = plc_h.LOGICAL_WRITE_CMD
	   FData.Function = plc_h.LOGICAL_WRITE_FNC
	   if FData.FileType == plc_h.FLOAT_TYPE {
		   NumElements = len(FData.FloatData) 
		   FData.Size = byte(NumElements * plc_h.FLOATLEN)
	   } else if (FData.FileType != plc_h.STATUS_TYPE) && (FData.FileType != plc_h.STRING_TYPE)	{
		   NumElements = len(FData.WordData)
		   FData.Size = byte(NumElements * plc_h.WORDLEN)
	   } else {
	     NumElements =	len(FData.ByteData)
	     FData.Size = byte(NumElements)
	   } 
	} //if "READ"

	FData.Status     = 0  
	FData.TNS        = GetTNS() 
  //	StrToFileData(Element, FData)
  // FData.FileNo     = FNum                  //Filled in by FileStrToFileData(Element, &FData)
	//FData.FileType   = FType
	//FData.Element    = ElemNum
	//FData.SubElement = SubNum
				
	
	//FData.WordData   = make([]uint16,0)
	
	FData.Data = append(FData.Data,FData.PutCmd)
	FData.Data = append(FData.Data,FData.Status)
	Junk      := PLCUtils.Int16ToBytes(FData.TNS)
	FData.Data = append(FData.Data,Junk[:]...)
	FData.Data = append(FData.Data,FData.Function)
	FData.Data = append(FData.Data,FData.Size)
	FData.Data = append(FData.Data,FData.FileNo)
	FData.Data = append(FData.Data,FData.FileType)
	FData.Data = append(FData.Data,FData.Element)
	FData.Data = append(FData.Data,FData.SubElement)

	if RW == "READ" {
		FData.ByteData  = make([]byte,0)
	    FData.WordData   = make([]uint16, 0)
	    FData.FloatData = make([]float32, 0)	
	} else {
    	if FData.FileType == plc_h.FLOAT_TYPE {
	   FData.WordData   = make([]uint16, 0)
	   FData.ByteData  = make([]byte,0)
	   FData.Length    = plc_h.MINLOGICAL + FData.Size
		   for I:=0; I < NumElements; I++ {
		      Junk := PLCUtils.Float32ToBytes(FData.FloatData[I])
		      FData.Data = append(FData.Data,Junk[:]...)
		      }
    } else if (FData.FileType != plc_h.STATUS_TYPE) && (FData.FileType != plc_h.STRING_TYPE)	{
	    FData.FloatData   = make([]float32, 0)
		FData.ByteData    = make([]byte,0)
	    FData.Length      = plc_h.MINLOGICAL + FData.Size 
		    for I:=0; I < NumElements; I++ {
		       Junk := PLCUtils.Int16ToBytes(FData.WordData[I])
		       FData.Data = append(FData.Data,Junk[:]...)
		       }
	} else {
	        FData.FloatData  = make([]float32,0)
	        FData.WordData   = make([]uint16, 0)
		    FData.Length     = plc_h.MINLOGICAL + FData.Size
			FData.Data       = FData.ByteData
	    }	
    }
   // FData.Size = byte(len(Sbyte))
    FData.Length      = plc_h.MINLOGICAL + FData.Size
	PLCPtr.PLC_EtherHdr.EIP_status = plc_h.Blank32
	Fill_CS_Address(PLCPtr, plc_h.RRADDTYPE, plc_h.GEN_FILE_CMD)
	Fill_CS_DataHdr(PLCPtr, FData.PutCmd ,FData.Function)
	PLCPtr.PCIP.Data.ItemData = FData.Data
	PLCPtr.PCIP.Data.DataHdr.DataLen = uint16(len(FData.Data))

	_,_,_ = PutData(PLCPtr,FData, RW)
}

//******************************************************************
//  Unprotected read/write  CIF file - Common Interface File       *
//  RW = "READ" denotes read else it is a write                    *
//  Elements is the no. of elements not the number of bytes        *
//******************************************************************
func CIFPut(PLCPtr plc_h.PPLC_EtherIP_info, FData plc_h.PFileData, Elements byte, Addr uint16, RW string) {
	var NumElements int
	var Size byte

	PLCPtr.PLC_EtherHdr.EIP_Command = plc_h.SENDRRDATA
	Size = Elements * plc_h.WORDLEN
	if RW == "READ" {
	   FData.Size   = Elements * plc_h.WORDLEN
	   FData.PutCmd = plc_h.CIF_READ_CMD
	   FData.Length = plc_h.MINCIFREAD// + Size
	} else {
		NumElements = len(FData.WordData)
		FData.Size  = byte(NumElements * plc_h.WORDLEN)
	    FData.PutCmd = plc_h.CIF_WRITE_CMD
	    FData.Length = plc_h.MINCIFWRITE + Size
	}	// READ
	
	FData.FileType = plc_h.INTEGER_TYPE
    FData.Status = 0
	FData.TNS    = GetTNS()
	FData.Addr   = Addr 
	FData.ByteData  = make([]byte,0)
	FData.FloatData = make([]float32, 0)
	FData.Data      = make([]byte,6)                       //Minimum length of data for CIF read/write Cmd, Sts, TNS(2), ADDR(2)
	FData.Data[0]   = FData.PutCmd
	FData.Data[1]   = FData.Status
	Junk           := PLCUtils.Int16ToBytes(FData.TNS)
	FData.Data[2] = Junk[0]
	FData.Data[3] = Junk[1]
	Junk           = PLCUtils.Int16ToBytes(Addr)
	FData.Data[4] = Junk[0]
	FData.Data[5] = Junk[1]
	if RW == "READ" {
	   FData.Data   = append(FData.Data,FData.Size) 
	} else {
		for I:=0; I < int(Elements); I++ {
		       Junk = PLCUtils.Int16ToBytes(FData.WordData[I])
		       FData.Data = append(FData.Data,Junk[:]...)
		       }
	} 
	PLCPtr.PLC_EtherHdr.EIP_status = plc_h.Blank32
	Fill_CS_Address(PLCPtr, plc_h.RRADDTYPE, plc_h.GEN_FILE_CMD)
	Fill_CS_DataHdr(PLCPtr, FData.PutCmd ,FData.Function)
	PLCPtr.PCIP.Data.DataHdr.DataLen = uint16(len(FData.Data))
	PLCPtr.PCIP.Data.ItemData = FData.Data
	_,_,_ = PutData(PLCPtr,FData, RW)
}

//*******************************************************************************************
//  Data returned by PLC from a Typed File R/W, Logical File R/W or CIF R/W                 *
//  DBytes is the byte sequence returned by the PLC                                         *
//  PFileData will already contain FType ex: 86 for Timer, FNum ex: 4 for T4,               *
//  ElemNum ex: 2 for T4:2, SubNum ex:PRE for T4:1/PRE                                      *
//  byte sequence is interpreted and returned in FData - a File Data Structure see plc_h.go *
//*******************************************************************************************
func GetData(FData plc_h.PFileData) {
 var NumDataBytes, NumElements, IDX, I int
	 NumDataBytes = len(FData.Data)
	 if FData.FileType == plc_h.FLOAT_TYPE {
		NumElements      = NumDataBytes / plc_h.FLOATLEN
		FData.ByteData   = make([]byte,0)
	    FData.WordData   = make([]uint16, 0)
	    FData.FloatData  = make([]float32, NumElements)	
	    for I = 0; I < NumElements; I++ {
		   IDX = I*plc_h.FLOATLEN	
		   FData.FloatData[I] = PLCUtils.BytesToFloat32(FData.Data[IDX:IDX+plc_h.FLOATLEN])
	    } // for	
    } else if FData.FileType != plc_h.STATUS_TYPE ||FData.FileType != plc_h.STRING_TYPE{
		NumElements      = NumDataBytes / plc_h.WORDLEN
		FData.ByteData   = make([]byte,0)
		FData.FloatData  = make([]float32, 0)
	    FData.WordData   = make([]uint16, NumElements)
		for  I = 0; I < NumElements; I++ {
		   IDX = I*plc_h.WORDLEN	
		   FData.WordData[I]= PLCUtils.BytesToInt16(FData.Data[IDX:IDX+plc_h.WORDLEN])
	    } //for
	} else {  //Interpret data as bytes ( for now - how about strings?)
	    NumElements = int(FData.Size)
	    FData.WordData   = make([]uint16, 0)
	    FData.FloatData  = make([]float32,0)
		FData.ByteData   =  make([]byte,NumElements)
		FData.ByteData   = FData.Data
	}//else	
} // function

//*************************************************
// Write the  command - receive the answer        * 
//*************************************************
func PutData(PLCPtr plc_h.PPLC_EtherIP_info, FData plc_h.PFileData, RW string) (string, string, error) {
	var aConn *net.TCPConn
	var dataBuff plc_h.Data_buffer //success = 0, recvBuff
    var PLCReply plc_h.PCCCReply
	var request []byte
	var s string

    PLCPtr.PLC_EtherHdr.CIP_Len = uint16(len(PLCPtr.PCIP.Data.ItemData))+plc_h.CSDLEN
	aConn = PLCPtr.Connection
	if RW != "READ" {
	   request = make([]byte, plc_h.ENCAPSHDRLEN + plc_h.CSDLEN + 4)                              //4 is minimum reply size
	} else {
	   request = make([]byte, plc_h.ENCAPSHDRLEN + plc_h.CSDLEN + FData.Size + 4)              //Allow space for reply data  
	}
    enCapSize :=plc_h.ENCAPSHDRLEN + plc_h.CSDLEN + int(PLCPtr.PCIP.Data.DataHdr.DataLen)

	dataBuff.Data, _ = IPInfoToByteArray(PLCPtr, enCapSize) //plcutils.IPInfoToByteArray(PLCPtr)
	dataBuff.Overall_len = plc_h.ETHIP_Header_Length + 4

	_, err := aConn.Write([]byte(dataBuff.Data))
	if err != nil {
		return "NOTOK", "", err
	}
	time.Sleep(100 * time.Millisecond)
	_, Rerr := aConn.Read(request)
	if (Rerr != nil) {
		PLCPtr.Error = plc_h.READERROR
		return "Read Error: " + PLCPtr.PLCHostIP, "", Rerr
	}
	_ = ByteArrayToReply(&PLCReply, request)
	FData.Data = PLCReply.Answer[4:]
	GetData(FData)	
	//s = ParseStatus(&PLCReply)
 
	return "OK", s, nil
}

func OpenFile(PLCPtr plc_h.PPLC_EtherIP_info, FileNo uint16, FileType byte) uint16 { //Tag
    var dataBuff plc_h.Data_buffer //success = 0, recvBuff
	var aConn *net.TCPConn
    var PLCReply plc_h.PCCCReply
//	var s string
	
	request := make([]byte, 6 + plc_h.CSDLEN + plc_h.ENCAPSHDRLEN) // 6 bytes Cmd, Sts, TNS(2 bytes), Tag(2 bytes)
	aConn = PLCPtr.Connection
	PLCPtr.PLC_EtherHdr.EIP_status = plc_h.Blank32
	PLCPtr.PCIP.Data.DataHdr.DataLen = 9   //Data length
	Fill_CS_Address(PLCPtr, plc_h.RRADDTYPE , plc_h.OPEN_FILE_CMD)
	Fill_CS_DataHdr(PLCPtr, plc_h.OPEN_FILE_CMD, plc_h.OPEN_FILE_FNC)                   // Includes Cmd, Status, TNS
	PLCPtr.PLC_EtherHdr.CIP_Len += 4
	PLCPtr.PCIP.Data.DataHdr.DataLen += 4
	dataBuff.Data, _ = IPInfoToByteArray(PLCPtr, 5 + plc_h.CSDLEN + plc_h.ENCAPSHDRLEN) // Cmd, Status, TNS(2) & Fnc

	dataBuff.Data = append(dataBuff.Data,plc_h.READ_WRITE)
	junk := PLCUtils.Int16ToBytes(FileNo)
	dataBuff.Data = append(dataBuff.Data,junk[:]...)
	dataBuff.Data = append(dataBuff.Data,FileType)

	dataBuff.Overall_len = plc_h.ETHIP_Header_Length + 9
   	_, err := aConn.Write([]byte(dataBuff.Data))
	if err != nil {
		return 0 //"NOTOK", ""
	}

    time.Sleep(100 * time.Millisecond)
	_, Rerr := aConn.Read(request)
	if (Rerr != nil) {
		PLCPtr.Error = plc_h.READERROR
		return 0// "Read Error: " + PLCPtr.PLCHostIP, ""
	}
	_ = ByteArrayToReply(&PLCReply, request) 
	if len(PLCReply.Answer) < 6  { //Error
	   return 0
	} else if PLCReply.Answer[1] != 0 {
	   return 0	
	}
	return PLCUtils.BytesToInt16(PLCReply.Answer[4:])
}
 
func CloseFile(PLCPtr plc_h.PPLC_EtherIP_info,Tag uint16) uint16 { //Tag
    var dataBuff plc_h.Data_buffer //success = 0, recvBuff
	var aConn *net.TCPConn
	
	request := make([]byte, 4 + plc_h.CSDLEN + plc_h.ENCAPSHDRLEN) // 4 bytes Cmd, Sts, Tag(2)
	aConn = PLCPtr.Connection
	PLCPtr.PLC_EtherHdr.EIP_status = plc_h.Blank32
	PLCPtr.PCIP.Data.DataHdr.DataLen = 7                           // Data length
	Fill_CS_Address(PLCPtr, plc_h.RRADDTYPE , plc_h.CLOSE_FILE_CMD)
	Fill_CS_DataHdr(PLCPtr, plc_h.CLOSE_FILE_CMD, plc_h.CLOSE_FILE_FNC)                   // Includes Cmd, Status, TNS
	PLCPtr.PLC_EtherHdr.CIP_Len += 2                                                      // Just 2 byte tag after Fnc
	PLCPtr.PCIP.Data.DataHdr.DataLen += 2
	dataBuff.Data, _ = IPInfoToByteArray(PLCPtr, 5 + plc_h.CSDLEN + plc_h.ENCAPSHDRLEN)   // Cmd, Status, TNS(2) & Fnc
    junk := PLCUtils.Int16ToBytes(Tag)
	dataBuff.Data = append(dataBuff.Data,junk[:]...)

	dataBuff.Overall_len = plc_h.ETHIP_Header_Length + 7
   	_, err := aConn.Write([]byte(dataBuff.Data))
	
	if err != nil {
		return 0 //"NOTOK", ""
	}

    time.Sleep(100 * time.Millisecond)
	_, Rerr := aConn.Read(request)
	
	if (Rerr != nil) {
		PLCPtr.Error = plc_h.READERROR
		return 0// "Read Error: " + PLCPtr.PLCHostIP, ""
	}
	return PLCUtils.BytesToInt16(request[2:])
 }

func ForwardOpenReply(PLCPtr plc_h.PPLC_EtherIP_info,buf [128]byte) (err error) {
    AddLn  := PLCUtils.ByteToUint16(buf[34:36])
    Fill_CS_Address(PLCPtr, plc_h.NULLADDRESS, plc_h.GEN_FILE_CMD) 
	PLCPtr.PCIP.Data.DataHdr.CSItemType_ID = plc_h.UNCONNECTEDDATA
	PLCPtr.PCIP.Data.DataHdr.DataLen = 30                       //UNCONNECTEDDATA word + Length word
	OpenFwdBuf := make([]byte,0)
	
	// Encapsulation header
	OpenFwdBuf = append(OpenFwdBuf,buf[0:2]...)                 //Command
    L := PLCUtils.Int16ToBytes(46)
	OpenFwdBuf = append(OpenFwdBuf,L...)
	OpenFwdBuf = append(OpenFwdBuf,buf[4:8]...)                 //Session Handle 
	OpenFwdBuf = append(OpenFwdBuf,buf[8:12]...)                //Status
	OpenFwdBuf = append(OpenFwdBuf,buf[12:20]...)               //Context
	OpenFwdBuf = append(OpenFwdBuf,buf[20:24]...)               //Options
	              
	//Command Specific data
	OpenFwdBuf = append(OpenFwdBuf,plc_h.Blank32[:]...)          // CIP Interface Handle = 00 00 00 00
	L = PLCUtils.Int16ToBytes(PLCPtr.PCIP.CIPHdr.CipTimeOut)     // PLC Timeout   
	OpenFwdBuf = append(OpenFwdBuf,L...)                         // echo same Timeout  
	L = PLCUtils.Int16ToBytes(PLCPtr.PCIP.CIPHdr.ItemCnt)        // Item count   
	OpenFwdBuf = append(OpenFwdBuf,L...)                         // Item count - still 2
	L = PLCUtils.Int16ToBytes(plc_h.NULLADDRESS)         
	OpenFwdBuf = append(OpenFwdBuf,L...)                         // Null address = 00 00
	OpenFwdBuf = append(OpenFwdBuf,plc_h.Blank16[:]...)          // Length of Address data is zero - data is nil
	L = PLCUtils.Int16ToBytes(plc_h.UNCONNECTEDDATA)  
	OpenFwdBuf = append(OpenFwdBuf,L...)                         // Unconnected data = 0xB2 00
	L = PLCUtils.Int16ToBytes(30)
	OpenFwdBuf = append(OpenFwdBuf,L...)                         // 30 bytes coming in CIP section
		
	//CIP Message
	OpenFwdBuf = append(OpenFwdBuf,plc_h.FORWARD_OPEN_RES)  
	OpenFwdBuf = append(OpenFwdBuf,0x00)
	OpenFwdBuf = append(OpenFwdBuf,0x00)
	OpenFwdBuf = append(OpenFwdBuf,0x00)
	
	O_T := PLCUtils.RandSession()
	OpenFwdBuf = append(OpenFwdBuf, O_T[:]...)                       // O -> T connection ID
	copy(PLCPtr.PLCConnectID[:],O_T[:])                               // Store computer connection ID
	OpenFwdBuf = append(OpenFwdBuf, buf[52+AddLn:56+AddLn]...)       // T -> O connection ID
	copy(PLCPtr.PLCConnectID[:],buf[52+AddLn:56+AddLn])              // Store PLC connection ID
	OpenFwdBuf = append(OpenFwdBuf, buf[56+AddLn:58+AddLn]...)       // Connection S/N
	PLCPtr.ConnectSN = PLCUtils.ByteToUint16(buf[56+AddLn:58+AddLn]) // Store Connection S/N
	OpenFwdBuf = append(OpenFwdBuf, buf[58+AddLn:60+AddLn]...)       // Vendor ID
	OpenFwdBuf = append(OpenFwdBuf, buf[60+AddLn:64+AddLn]...)       // Originator S/N
	OpenFwdBuf = append(OpenFwdBuf, buf[68+AddLn:72+AddLn]...)       // RPI
	OpenFwdBuf = append(OpenFwdBuf, buf[68+AddLn:72+AddLn]...)       // RPI 2
	OpenFwdBuf = append(OpenFwdBuf, 0x00)
	OpenFwdBuf = append(OpenFwdBuf, 0x00)
	PLCPtr.PCIP.Data.ItemData = OpenFwdBuf
	_, err = PLCPtr.Connection.Write(OpenFwdBuf)
	return err

}

func ServiceReply(PLCPtr plc_h.PPLC_EtherIP_info) (err error) {
	ReplyBuf := make([]byte,0)
	// Encapsulation header
	jnk := PLCUtils.Int16ToBytes(PLCPtr.PLC_EtherHdr.EIP_Command)
	ReplyBuf = append(ReplyBuf,jnk[:]...) 
	jnk = PLCUtils.Int16ToBytes(PLCPtr.PLC_EtherHdr.CIP_Len)
	ReplyBuf = append(ReplyBuf,jnk[:]...) 
	ReplyBuf = append(ReplyBuf,PLCPtr.PLC_EtherHdr.Session_handle[:]...) 
	ReplyBuf = append(ReplyBuf,PLCPtr.PLC_EtherHdr.EIP_status[:]...) 
	ReplyBuf = append(ReplyBuf,PLCPtr.PLC_EtherHdr.Context[:]...) 
	ReplyBuf = append(ReplyBuf,PLCPtr.PLC_EtherHdr.Options[:]...) 
	// Command Specific Data
	ReplyBuf = append(ReplyBuf,plc_h.Blank32[:]...)                           // Interface Handle CIP
	ReplyBuf = append(ReplyBuf,plc_h.Blank16[:]...)                           // Timeout = 0 - Don't know why
	ReplyBuf = append(ReplyBuf,PLCUtils.Int16ToBytes(0x02)[:]...)             // Item count = 2   
	jnk = PLCUtils.Int16ToBytes(plc_h.SENDRRADDRESS)                           
	ReplyBuf = append(ReplyBuf,jnk[:]...)                                     // Connected Address Item
	ReplyBuf = append(ReplyBuf,PLCUtils.Int16ToBytes(4)[:]...)                // Address data length
	ReplyBuf = append(ReplyBuf,PLCPtr.PLCConnectID[:]...) 
	jnk = PLCUtils.Int16ToBytes(plc_h.SENDRRDATA)                             // Connect ID of PLC - length 4
	ReplyBuf = append(ReplyBuf,jnk[:]...)                                     // Connected Data Item
	ReplyBuf = append(ReplyBuf,PLCUtils.Int16ToBytes(6)[:]...)                // Reply contains 6 bytes
	jnk = PLCUtils.Int16ToBytes(PLCPtr.SeqCount+1)
	ReplyBuf = append(ReplyBuf,jnk[:]...)                                     // Incremented Sequence count
	ReplyBuf = append(ReplyBuf,plc_h.CIPReply[:]...)
	_, err = PLCPtr.Connection.Write(ReplyBuf)

	return err
}

func ParseUnitData(PLCPtr plc_h.PPLC_EtherIP_info,Service plc_h.PPLCService, buf [128]byte) {
	var ConID [4]byte
    var AddLen, DataLen uint16
		
	AddLen = PLCUtils.BytesToInt16(buf[34:36])
	copy(ConID[:],buf[36:40])
	DataLen = PLCUtils.BytesToInt16(buf[38+AddLen:40+AddLen])
	Service.Seq_Count = PLCUtils.BytesToInt16(buf[40+AddLen:42+AddLen])
	PLCPtr.SeqCount  = Service.Seq_Count
	Service.Service = buf[42+AddLen]
	Service.ReqPathSize = buf[43+AddLen]
	copy(Service.ReqPath[:],buf[44+AddLen:48+AddLen])
	
	if PLCUtils.ByteArray4Equals(ConID, PLCPtr.PLCConnectID) == false {   //PLC should use my connection ID
		fmt.Println(" Connection ID does not match")
		}
	// CIP data
	DataBuf := make([]byte,DataLen)	
	copy(DataBuf[:],buf[48+AddLen:])
	Service.Cmd_Spc_Data.GetCmd     = DataBuf[7]
	Service.Cmd_Spc_Data.Status     = DataBuf[8]
	Service.Cmd_Spc_Data.TNS        = PLCUtils.BytesToInt16(DataBuf[9:11])
	Service.Cmd_Spc_Data.Function   = DataBuf[11]
	Service.Cmd_Spc_Data.Size       = DataBuf[12]
	Service.Cmd_Spc_Data.FileNo     = DataBuf[13]
	Service.Cmd_Spc_Data.FileType   = DataBuf[14]
	Service.Cmd_Spc_Data.Element    = DataBuf[15]
	Service.Cmd_Spc_Data.SubElement = DataBuf[16]
	Service.Cmd_Spc_Data.Data       = make([]byte,Service.Cmd_Spc_Data.Size)
	copy(Service.Cmd_Spc_Data.Data,DataBuf[17:])
	DecodeData(Service , Service.Cmd_Spc_Data.Data) 
    //Command Specific Data - set up for reply
	PLCPtr.PLC_EtherHdr.CIP_Len = 26
	copy(PLCPtr.PCIP.CIPHdr.CIPHandle[:],buf[24:28])                     // CSD Interface handle = CIP
	PLCPtr.PCIP.CIPHdr.CipTimeOut = 0                                    // CSD Time out = 0
	PLCPtr.PCIP.CIPHdr.ItemCnt = 2                                       // CSD Itemcount = 2
	PLCPtr.PCIP.Address.AddressHdr.CSItemType_ID = plc_h.SENDRRADDRESS   // CSD Connected Address Item
	PLCPtr.PCIP.Address.AddressHdr.DataLen       = 4                     // CSD Address Data Length
	PLCPtr.PCIP.Data.DataHdr.CSItemType_ID       = plc_h.SENDRRDATA      // CSD Connected Data Item
	PLCPtr.PCIP.Data.DataHdr.DataLen = 6                                 // CSD Data Length
	PLCPtr.PCIP.Data.ItemData = make([]byte,6)
	Seq := PLCUtils.Int16ToBytes(Service.Seq_Count)
	copy(PLCPtr.PCIP.Data.ItemData[:], Seq[:])   
	PLCPtr.PCIP.Data.ItemData = append(PLCPtr.PCIP.Data.ItemData,plc_h.CIPReply[:]...)  // CSD Data
       _ = ServiceReply(PLCPtr)
	}
	
func DecodeInteger(Service plc_h.PPLCService, buf []byte) {
 LN := len(buf)
 Service.Cmd_Spc_Data.WordData = make([]uint16,LN/4)
 for I := 0; I < LN	/2; I++ {
	Service.Cmd_Spc_Data.WordData[I] = PLCUtils.BytesToInt16(buf[I*2:I*2+2])
    }
}

func DecodeFloat(Service plc_h.PPLCService, buf []byte) {
 LN := len(buf)
 Service.Cmd_Spc_Data.FloatData = make([]float32,LN/4)
 for I := 0; I < LN	/4; I++ {
	Service.Cmd_Spc_Data.FloatData[I] = PLCUtils.BytesToFloat32(buf[I*4:I*4+4])
    }
}
	
func DecodeData(Service plc_h.PPLCService, ByteBuf []byte) {
	var FType byte
	FType = Service.Cmd_Spc_Data.FileType
	switch FType {
	  case plc_h.STATUS_TYPE: 
	      // Service.Cmd_Spc_Data.String = ByteBuf use strconv
	  case plc_h.BIT_TYPE:
	       DecodeInteger(Service, ByteBuf)
	  case plc_h.TIMER_TYPE:
	       DecodeInteger(Service, ByteBuf)
	  case plc_h.COUNTER_TYPE:
	       DecodeInteger(Service, ByteBuf)
	  case plc_h.CONTROL_TYPE:
	       DecodeInteger(Service, ByteBuf)
	  case plc_h.INTEGER_TYPE:
	       DecodeInteger(Service, ByteBuf)
	  case plc_h.FLOAT_TYPE:
	       DecodeFloat(Service, ByteBuf)
	  case plc_h.OUTPUT_TYPE:
	       DecodeInteger(Service, ByteBuf)
	  case plc_h.INPUT_TYPE:
	       DecodeInteger(Service, ByteBuf)
	  case plc_h.STRING_TYPE:
	       return
	  case plc_h.ASCII_TYPE:
	       return
	  case plc_h.BCD_TYPE:
	       return
      	   }
				
}

func PrintData(Service plc_h.PPLCService) {
	var FType byte
	FType = Service.Cmd_Spc_Data.FileType
	switch FType {
	  case plc_h.STATUS_TYPE: 
	       return
	  case plc_h.BIT_TYPE:
	       PrintInteger(Service)
	  case plc_h.TIMER_TYPE:
	       PrintInteger(Service)
	  case plc_h.COUNTER_TYPE:
	       PrintInteger(Service)
	  case plc_h.CONTROL_TYPE:
	       PrintInteger(Service)
	  case plc_h.INTEGER_TYPE:
	       PrintInteger(Service)
	  case plc_h.FLOAT_TYPE:
	       PrintFloat(Service)
	  case plc_h.OUTPUT_TYPE:
	       PrintInteger(Service)
	  case plc_h.INPUT_TYPE:
	       PrintInteger(Service,)
	  case plc_h.STRING_TYPE:
	       fmt.Printf("String %v \n",Service.Cmd_Spc_Data.String)
	  case plc_h.ASCII_TYPE:
	       return
	  case plc_h.BCD_TYPE:
	       return
      	   }
	
}

func PrintFloat(Service plc_h.PPLCService) {
 LN := len(Service.Cmd_Spc_Data.FloatData)
 for I := 0; I < LN	; I++ {
	fmt.Printf("Float %v ", Service.Cmd_Spc_Data.FloatData[I])
    }
 fmt.Println()		
}

func PrintInteger(Service plc_h.PPLCService) {
 LN := len(Service.Cmd_Spc_Data.WordData)
 for I := 0; I < LN	; I++ {
	fmt.Printf("Integer %v ", Service.Cmd_Spc_Data.WordData)
    }
 fmt.Println()	
}
