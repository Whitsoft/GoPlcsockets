package plc_h

import (
	"fmt"
	"net"
)
var 	Blank16       = [2]byte{0x00,0x00}
var 	Blank32       = [4]byte{0x00,0x00,0x00,0x00}
var 	Blank64       = [8]byte{0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00}
var     CIPReply      = [4]byte{0xCB,0x00,0x00,0x00} 

const (
	FALSE         = 1
	TRUE          = 0
	MAXWORDS      = 24
	MAXFILE       = 120
	MAXFSET       = 64
	TYPED_LOGICAL = 1
	PROT_TYPED    = 2
	UNPROTECTED   = 3
	DIAGNOSTIC    = 3
	FOPEN         = 4
	FCLOSE        = 5
	FLSTATUS      = 6
	CONNECT_CMD   = 1
	READ_ONLY     = 1
	READ_WRITE    = 3
	PROTOVERSION  = 1
	ENCAPSHDRLEN  = 24
	CSDLEN        = 17
	
	MINLOGICAL    = 10  //bytes ahead of data
	MINTYPED      = 11
	MINCIFREAD    = 7
	MINCIFWRITE   = 6
	CMDSTSTNS     = 4
	WORDLEN       = 2
	INTLEN        = 4
	FLOATLEN      = 4

	CELL_DFLT_TIMEOUT = 5000
	NULLADDRESS       = 0x00
	CONNECTEDDATA     = 0xB1
	UNCONNECTEDDATA   = 0xB2
	STATTYPE          = 0x85
	RRADDTYPE         = 0x81
	RRDATATYPE        = 0x91 //Who knows - undocumented
	SENDRRADDRESS     = 0xA1
	SENDRRDATA        = 0xB1
	PLCCOUNT          = 1
	SUBPRE            = 1
    SUBACC            = 2
	ADDRTYPESCOUNT     = 26
	DATA_Buffer_Length = 2024
	CPH_Null           = 0

	SLC    = 3
	MICRO  = 4
	CTRUE  = 0
	CFALSE = 1

	_ENET_HEADER_LEN = 28
	_CUSTOM_LEN      = 16
	PCCC_VERSION     = 4
	PCCC_BACKLOG     = 5

	// Use one  the these values for the fnc, prot logical. Do not use any other
	// values doing so may result in unpredictable results.
	//**********************************************
	// FILE TYPES
	//**********************************************
	STATUS_TYPE  = 0x84
	BIT_TYPE     = 0x85
	TIMER_TYPE   = 0x86
	COUNTER_TYPE = 0x87
	CONTROL_TYPE = 0x88
	INTEGER_TYPE = 0x89
	FLOAT_TYPE   = 0x8A
	OUTPUT_TYPE  = 0x8B
	INPUT_TYPE   = 0x8C
	STRING_TYPE  = 0x8D
	ASCII_TYPE   = 0x8E
	BCD_TYPE     = 0x8F

	ETHERNET       = 1
	CIAddLEN      = 15
	CIDataLEN     = 94
	REGLEN         = 28
	OK             = 0
	NOTOK          = -1
	NOSESSIONMATCH = -2
	NOCONTEXTMATCH = -3
	NOADDRESSMATCH = -4
	STATUSERROR    = -5
	NOHOST         = -6
	BADADDR        = -7
	NOCONNECT      = -8
	BADCMDRET      = -9
	WINSOCKERROR   = -10
	WRITEERROR     = -11
	TCPERROR       = -12
	READERROR      = -13


	//constants for Ethernet/IP (encapsulation) header
	NOP                 = 0
	List_Targets        = 1
	List_Services       = 4
	List_Identity       = 0x63
	List_Interfaces     = 0x64
	Register_Session    = 0x65
	UnRegister_Session  = 0x66
	SendRRData          = 0x6f
	SendUnitData        = 0x70
	Indicate_Status     = 0x72
	Cancel              = 0x73
	ETHIP_Header_Length = 24
	DATA_MINLEN         = 16
	TIMEOUT             = 0x400

//*****************************************
// CIP Service Codes                      * 
//*****************************************
	FORWARD_OPEN_REQ        = 0x54
	FORWARD_OPEN_RES        = 0xD4
	FORWARD_CLOSE           = 0x4E
	CIP_GENERIC             = 0x4B
	CIP_GENERIC_REPLY       = 0xCB
	CIP_READ                = 0x4C
	CIP_READ_REPLY          = 0xCC
	CIP_READ_FRAG           = 0x52
	CIP_READ_FRAG_REPLY     = 0xD2
	CIP_WRITE               = 0x4D
	CIP_WRITE_REPLY         = 0xCD
	CIP_WRITE_FRAG          = 0x53
	CIP_WRITE_FRAG_REPLY    = 0xD3
	
    //*****************************************
	// PCCC file numbers
	//*****************************************
	BIT_NO              = 0x0003
	TIMER_NO            = 0x0004
	COUNTER_NO          = 0x0005
	CONTROL_NO          = 0x0006
	INTEGER_NO          = 0x0007
	FLOAT_NO            = 0x0008
	CIF_NO              = 0x0009
	//*****************************************
	// PCCC commands
	//*****************************************
	NULL_CMD              = 0x00
	GEN_FILE_CMD          = 0x0F
	OPEN_FILE_CMD         = 0x0F
	CLOSE_FILE_CMD        = 0x0F
	TYPE_FILE_READ_CMD    = 0x0F
	TYPE_FILE_WRITE_CMD   = 0x0F
	LOGICAL_READ_CMD      = 0x0F
	LOGICAL_WRITE_CMD     = 0x0F
	DIAG_STATUS_CMD       = 0x06
	CIF_READ_CMD          = 0x01
	CIF_WRITE_CMD         = 0x08
	
	PLC_ANSWER            = 0x4F
	PLC_CIF_RD_ANSWER     = 0x41
	PLC_CIF_WRT_ANSWER    = 0x48
	
	OPEN_FILE_FNC         = 0x81
	CLOSE_FILE_FNC        = 0x82
    TYPE_FILE_READ_FNC    = 0xA7
	TYPE_FILE_WRITE_FNC   = 0xAF
	LOGICAL_READ_FNC      = 0xA2
	LOGICAL_WRITE_FNC     = 0xAA
	DIAG_STATUS_FNC       = 0x03
	//UNPROT_READ_FNC       = none required
	//UNPROT_WRITE_FNC      = none required
	INPUT            = "I"   // Input
	OUTPUT           = "O"   // Input
	STATUS           = "S"	 // Status
    BINARY           = "B"	 // Binary
    TIMER            = "T"   // Timer
	COUNTER          = "C"   // Counter
	CONTROL          = "R"	 // Control
	INTEGER          = "N"	 // Integer
    FLOAT            = "F"	 // Float
	ASCII            = "A"	 // ASCII
	BCD              = "D"	 // BCD
	BLOCKTRANS       = "BT"  // Block Transfer
    LONGINT          = "L"	 // Long Integer
	MESSAGE          = "MG"	 // Message
	PID              = "PD"  // PID
    //			     = "SC"  // ??
	STRING           = "ST"	 // String
	PLCNAME          = "PN"	 // PLC Name
	RUNG             = "RG"  // Rung
	FORCEINTABLE     = "FI"  // Input Force Table
	FORCEOUTTABLE    = "FO"  // Output Force Table
    SECTION3         = "XA"  // Section 3 File
	SECTION4         = "XB"  // Section 4 File
	SECTION5         = "XC"  // Section 5 File
	SECTION6         = "XD"  // Section 6 File
	SECTIONFF        = "FF"  // Force File Section
)

//***************************************************
//* just a buffer of bytes                          *
//* and a count of those bytes that are significant *
//***************************************************
type PSimpleBuf *SimpleBuf
type SimpleBuf struct {
	Cnt  int
	Data [MAXFILE * 2]byte //[0..MAXFILE*2-1]
}

type FloatRecord struct {
	HiWord uint16
	LoWord uint16
}

type PLCFile [MAXWORDS]uint16
type PLCFloat [MAXWORDS]FloatRecord
type PLCTimer [3]uint16
type PLCCounter [3]uint16
type File [MAXFILE]uint16

type PServices *_services
type _services struct {
	S_type  uint16
	Length  uint16
	Version uint16
	Flags   uint16
	Name    [16]byte
}

type Float_Buffer [33]byte

//******************************************
// A large buffer of bytes                 *
//******************************************
type PData_buffer *Data_buffer
type Data_buffer struct {
	Data        []byte
	Length      uint16
	Overall_len uint16
}

type Address_Hdr struct {
	CSItemType_ID uint16 //usually $91
	DataLen       uint16
}

//***********************************************************
// Address Item - Part of command specific data - CIP       *
// Acronym ACPF - 1. Item count then 2. Address Item        *
//***********************************************************
type Address_Item struct {
	AddressHdr Address_Hdr
	ItemData    []byte
}

type Data_Hdr struct {
	CSItemType_ID uint16 //usually $91
	DataLen       uint16
//	Cmd           byte
//	Sts           byte
//	Tns           int16
//	Fnc           byte
}

//***********************************************************
// Data Item -    Part of command specific data - CIP       *
// Acronym ACPF - 2. Address Item then 3. Data Item         *
//***********************************************************
type PData_Item *Data_Item
type Data_Item struct {
	DataHdr Data_Hdr //4 bytes
	ItemData []byte   //[CIDataLEN]byte
	//  fnc byte
	// FileNo byte
	// FileType byte
	// Cmdsize byte //sans data  size
	//  fset uint16
	//  Elem byte
	//  SubElem byte
	// Addr uint16      //unique to unprotected file read N9 or N7 SLC & Micro?
	// tag uint16       //unique to protected typed file read/write
	// data [0..63]  byte

}

type CIP_Hdr struct {
	CIPHandle  [4]byte // cardinal //zero
	CipTimeOut uint16
	ItemCnt    uint16
}

type CIP struct {
	CIPHdr   CIP_Hdr
	Address Address_Item
	Data    Data_Item
}

type Ethernet_header struct { //284 bytes
	Mode        byte
	Submode     byte
	Pccc_length uint16
	Connection  uint32
	Status      uint32
	Custom      [26]byte
	Df1_data1   [246]byte
}

//***********************************************************
// Ethernet/IP Encapsulation header - same for all commands *
// Start of Ethernet/IP Industrial protocol                 *
//***********************************************************
type PEtherIP_Hdr *EtherIP_Hdr
type EtherIP_Hdr struct {     //24 bytes
	EIP_Command    uint16     // Such as as 0x006F SendRRData
	CIP_Len        uint16     // Length of command specific data
	Session_handle [4]byte
	EIP_status     [4]byte    //0x00000000 = success
	Context        [8]byte    //Sender context
	Options        [4]byte    // total 24 bytes
}

//***************************************************
// Convenient structure for internal use            *
// Not part of communications structures            *
//  type IP []byte   from gonet                     *
//***************************************************
type //Keep this data for individual PLC connections
PPLC_EtherIP_info *PLC_EtherIP_info
type PLC_EtherIP_info struct {
	PLC_EtherHdr  EtherIP_Hdr
	PCIP          CIP
	Connection    *net.TCPConn
	PLCHostIP     string
	PLCHostPort   uint16
	Error         int
	Tag           byte
	FType         byte
	Connected     byte //1 = connected
	SeqCount      uint16
	ConnectSN     uint16
	CPConnectID   [4]byte  //  O->T 
	PLCConnectID  [4]byte  //  T->O
}

//************************************
// structure to get reply from PLC   *
//************************************
type PPCCCReply *PCCCReply
type PCCCReply struct {
	Cmd            uint16
	Length         uint16
	Session_handle [4]byte
	Context        [8]byte
	Status         [4]byte    //0x00000000 = success
	Error          uint16
	Answer         []byte 
}

//************************************
// structure to get reply from PLC   *
// for unprotected file access       *
//************************************
type PPCCCReplyUn *PCCCReplyUn
type PCCCReplyUn struct {
	Size   byte
	Answer []byte //was 31 DDW
}

type Custom_connect struct {
	Version int16
	Backlog int16
	Junk    [12]byte
}


//************************************************************
// Data structure to carry all relevant information for      *
// Logical read/write, CIF read/write, Typed file read/write *
// Use this data to populate
//************************************************************
type PFileData *FileData
type FileData struct {
	Section    int          // For extended PLC file memory
	Length     byte         // Length of buffer
	PutCmd     byte         // Command to PLC
	GetCmd     byte         // Command from PLC
	TNS        uint16
	TNSGet     uint16
	Status     byte         // Read
	EXStatus   byte         // Read only when Status 0xF0
	Function   byte         // Write
	Offset     uint16       // Typed file Read/Write
	FileNo     byte         // Logical Read/Write
	FileType   byte         // Typed file Read/Write
	Element    byte         // Logical Read/Write
	SubElement byte         // Logical Read/Write
	Bit        byte
	Addr       uint16       // CIF Read/Write
	Tag        uint16       // Typed file Read/Write
	FloatData  []float32    // Converted float32 data
	ByteData   []byte
	WordData   []uint16     // Converted word data
	TypeLen    int          // Length of elements contained in Data
	Size       byte         // Read/Write - size in bytes of data
	String     string
	Data       []byte       // Raw data element values in byte form - just float, word or byte for now
	FileDesc   string
}

type PPLCMessage  *PLCMessage
type PLCMessage  struct {
     SHandle    []byte //array[0..3] of byte;
     CMD        byte
     ConID      []byte //array[0..3] of byte;
     funct      byte
     Size       byte
     FileNo     byte
     FileType   byte
     Element    byte
     SubElement byte
     Data       []byte //: array[0..63] of byte;
     FileDesc   string
}
//***********************************************************
// PLCService plus CSD - will contain the DataItem.ItemData *
// Coming from
type PPLCService  *PLCService
type PLCService  struct {
	 Seq_Count     uint16
	 Service       byte        //Start of Common Industrial Protocol
	 Path_Size     byte
	 Req_Path      [4]byte
	 Cmd_Spc_Data  FileData  //CSD
}

type CSD struct { //Command specific data
     SHandle    []byte    //array[0..3] of byte;
	 Unknown    []byte    // 7 bytes - usually 07 01 00 4C 6D 4F 40
     CMD        byte      // Such as 0F
	 Status     byte
	 TNS        uint16
     Funct      byte
     Size       byte
     FileNo     byte
     FileType   byte
     Element    byte
     SubElement byte
     Data       []byte //: array[0..63] of byte;
     FileDesc   string
	
}
func main() {
	fmt.Println("hello")
}
