package plc_h

import (
	"fmt"
	"net"
)

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

	CELL_DFLT_TIMEOUT = 5000
	STATTYPE          = 0x85
	RRADDTYPE         = 0x81
	RRDATATYPE        = 0x91 //Who knows - undocumented
	PLCCOUNT          = 1

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
	FNC_STATUS  = 0x84
	FNC_BIT     = 0x85
	FNC_TIMER   = 0x86
	FNC_COUNTER = 0x87
	FNC_CONTROL = 0x88
	FNC_INTEGER = 0x89
	FNC_FLOAT   = 0x8A
	FNC_OUTPUT  = 0x8B
	FNC_INPUT   = 0x8C
	FNC_STRING  = 0x8D
	FNC_ASCII   = 0x8E
	FNC_BCD     = 0x8F

	ETHERNET       = 1
	CIPADDLEN      = 15
	CIPDATALEN     = 94
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
	// PCCC commands
	//*****************************************
	GEN_FILE_CMD          = 0x0F
	OPEN_FILE_CMD         = 0x0F
	CLOSE_FILE_CMD        = 0x0F
	PROT_FILE_READ_CMD    = 0x0F
	PROT_FILE_WRITE_CMD   = 0x0F
	LOGICAL_READ_CMD      = 0x0F
	LOGICAL_WRITE_CMD     = 0x0F
	DIAG_STATUS_CMD       = 0x06
	UNPROT_READ_CMD       = 0x01
	UNPROT_WRITE_CMD      = 0x08
	
	PLC_ANSWER            = 0x4F
	PLC_UNPROT_RD_ANSWER  = 0x41
	PLC_UNPROT_WRT_ANSWER = 0x48
	
	OPEN_FILE_FNC         = 0x81
	CLOSE_FNC_FNC         = 0x82
    PROT_FILE_READ_FNC    = 0xA7
	PROT_FILE_WRITE_FNC   = 0xAF
	LOGICAL_READ_FNC      = 0xA2
	LOGICAL_WRITE_FNC     = 0xAA
	DIAG_STATUS_FNC       = 0x03
	//UNPROT_READ_FNC       = none required
	//UNPROT_WRITE_FNC      = none required

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
type Pdata_buffer *Data_buffer
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
	PAddressHdr Address_Hdr
	ItemData    []byte
}

type Data_Hdr struct {
	CSItemType_ID uint16 //usually $91
	DataLen       uint16
	Cmd           byte
	Sts           byte
	Tns           int16
	Fnc           byte
}

//***********************************************************
// Data Item -    Part of command specific data - CIP       *
// Acronym ACPF - 2. Address Item then 3. Data Item         *
//***********************************************************
type PData_Item *Data_Item
type Data_Item struct {
	PDataHdr Data_Hdr //8 bytes
	ItemData []byte   //[CIPDATALEN]byte
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
	CIPHandle  uint32 // cardinal //zero
	CipTimeOut uint16
	ItemCnt    uint16
}

type CIP struct {
	CIPHdr   CIP_Hdr
	PAddress Address_Item
	PData    Data_Item
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
type EtherIP_Hdr struct { //24 bytes
	EIP_Command    uint16 // Such as as 0x006F SendRRData
	CIP_Len        uint16 // Length of command specific data
	Session_handle uint32
	EIP_status     uint32 //0x00000000 = success
	Context        uint64 //Sender context
	Options        uint32 // total 24 bytes
}

//***************************************************
// Convenient structure for internal use            *
// Not part of communications structures            *
//  type IP []byte   from gonet                     *
//***************************************************
type //Keep this data for individual PLC connections
PPLC_EtherIP_info *PLC_EtherIP_info
type PLC_EtherIP_info struct {
	PLC_EtherHdr EtherIP_Hdr
	PCIP         CIP
	Connection   *net.TCPConn
	PLCHostIP    string
	PLCHostPort  uint16
	Error        int
	Tag          byte
	FType        byte
	Connected    byte //1 = connected
}

//************************************
// structure to get reply from PLC   *
//************************************
type PPCCCReply *PCCCReply
type PCCCReply struct {
	Length   uint16
	Error    int
	CIPError int
	Status   byte
	Answer   []byte //was 31 DDW
}

//************************************
// structure to get reply from PLC   *
// for unprotected file access       *
//************************************
type PPCCCReplyUn *PCCCReplyUn
type PCCCReplyUn struct {
	Answer [128]byte //was 31 DDW
}

type Custom_connect struct {
	Version int16
	Backlog int16
	Junk    [12]byte
}

type PFileData *FileData
type FileData struct {
	Section    int
	Ffile      byte
	Element    byte
	Subelement byte
	Ftype      byte
	Typelen    int
	Bit        byte
	Length     byte
	Floatdata  byte
	Data       [24]byte
}

func main() {
	fmt.Println("hello")
}
