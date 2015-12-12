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
	"PLCUtils"
	"fmt"
)

const MINLEN = 12
const MINCSD = 4
const CSDLEN = 12 //IF Handle + T/O + Item cnt + Type ID (Address) + Len (Address)
var TNSValue uint16

func main() {
	var PLCPtr plc_h.PLC_EtherIP_info
	var FData plc_h.FileData

	PLCPtr.PLCHostIP = "192.168.1.50"
	PLCPtr.PLCHostPort = 44818
	_ = PLCFunctions.Connect(&PLCPtr)

	PLCFunctions.Register_session(&PLCPtr)
	PLCPtr.PLC_EtherHdr.EIP_Command = plc_h.SendRRData

	PLCPtr.PCIP.CIPHdr.CipTimeOut = plc_h.TIMEOUT
	junk := PLCUtils.Uint32ToByteArray(0)
	copy(PLCPtr.PCIP.CIPHdr.CIPHandle[:],junk[:])
	PLCPtr.PCIP.CIPHdr.ItemCnt = 2
	
    //**********************************************
    // Test Typed File read/write                  *
	// Open file to get a tag - used in read/write *
	// read/write - then close file                *
	//**********************************************	
	Floats := []float32{1245.5,6666.66}
	FData.FloatData = Floats
    FData.Tag = PLCFunctions.OpenFile(&PLCPtr, plc_h.FLOAT_NO, plc_h.FLOAT_TYPE)
	FData.FileType = plc_h.FLOAT_TYPE
	FData.Offset = 0
	FData.FloatData = Floats
	PLCFunctions.TypedFilePut (&PLCPtr,&FData, 2 ,"WRITE" )
	PLCFunctions.TypedFilePut (&PLCPtr,&FData, 2 ,"READ" )
	_ = PLCFunctions.CloseFile(&PLCPtr, FData.Tag)
	fmt.Printf("Typed File Read % v",FData.FloatData)
	fmt.Println()
	
	
	Ints := []uint16{0xff00,0x1234}
	FData.WordData = Ints
    PLCFunctions.CIFPut(&PLCPtr, &FData , 2, 0, "WRITE")   //Write 2 words (0xff00,0x1234) to N9:0
	PLCFunctions.CIFPut(&PLCPtr, &FData , 2, 0, "READ")    //Read the results
    fmt.Printf("CIF File Read % v",FData.WordData)
    fmt.Println()
    
	//*********************************************
    // Test Logical read/write                        *
	//*********************************************	
	PLCFunctions.LogicalPut(&PLCPtr,&FData,"B3:0/2",3, "READ") //T4:1/ACC
	fmt.Printf("Logical Read % v",FData.WordData)
	fmt.Println()
	PLCFunctions.UnRegister_session(&PLCPtr)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

