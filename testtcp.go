//* **********************************************************
// Test program to communicate with SLC500 / Micrologix PLC  *
// Using 1. Protected typed file read/write                  *
//       2. Protected typed logical read/write               *
//       3. Unprotected read/write                           *
// Also being tested are Diagnostic Status & open/close file * 
//************************************************************     
package main

import (
	"os"
	"plc_h"
	"PLCFunctions"
	"fmt"
)

const MINLEN = 12
const MINCSD = 4
const CSDLEN = 12 //IF Handle + T/O + Item cnt + Type ID (Address) + Len (Address)
var TNSValue uint16

func main() {
	//var AddItemLen, DataItemLen, ETH_IPLen uint16
	//var a []uint16

	//var answer string
	var PLCPtr plc_h.PLC_EtherIP_info
	var FData plc_h.FileData
    Floats := []float32{1245.5,6666.66}
	//0f:00:02:00:a2:04:04:86:01:02
    a := []uint16{0xff00,0x1234}
    //IG:= []byte{0x0f,0x00,0x02,0xa2,0x04,0x04,0x86,0x01,0x02}
	//PLCFunctions.FileStrToFileData("T4:1.3", &FData)
	//fmt.Println(FData)
	PLCPtr.PLCHostIP = "192.168.1.50"
	PLCPtr.PLCHostPort = 44818
	_ = PLCFunctions.Connect(&PLCPtr)

	PLCFunctions.Register_session(&PLCPtr)
	PLCPtr.PLC_EtherHdr.EIP_Command = plc_h.SendRRData

	PLCPtr.PCIP.CIPHdr.CipTimeOut = plc_h.TIMEOUT
	PLCPtr.PCIP.CIPHdr.CIPHandle = 0
	PLCPtr.PCIP.CIPHdr.ItemCnt = 2

    FData.Tag = PLCFunctions.OpenFile(&PLCPtr, plc_h.FLOAT_NO, plc_h.FLOAT_TYPE)
	FData.FileType = plc_h.FLOAT_TYPE
	FData.Offset = 0
	FData.FloatData = Floats
	PLCFunctions.TypedFilePut (&PLCPtr,&FData ,"WRITE" )
	 junk := PLCFunctions.CloseFile(&PLCPtr, FData.Tag)
	fmt.Printf("JUNK % x ",junk)
	os.Exit(0)
	PLCFunctions.FileStrToFileData("T4:1.3", &FData)
	FData.WordData = a
	//PLCFunctions.LogicalPut(&PLCPtr,&FData,"T4:1.3",2, "WRITE")
	FData.FileType = plc_h.INTEGER_TYPE
	FData.FloatData = Floats
	//PLCFunctions.LogicalPut(&PLCPtr,&FData, 2,4, 0x86, 1, 1, a, b,  c, "READ")
	fmt.Printf("Results ",FData.WordData)
	//PLCPtr.PLC_EtherHdr.CIPLen = 8 //minimum
	//_, answer = PLCFunctions.Get_Status(&PLCPtr)
//	fmt.Println(" ", answer)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

