package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"unsafe"

	// name the import
	snmp "github.com/gosnmp/gosnmp"
)

func main() {

	// setup environment variables
	fileName, pollPeriod, envTarget, envCommunity, port, oids := setup()

	// put together a struct containing connection parameters
	connectionParams := &snmp.GoSNMP{
		Target:    envTarget,                      // the network node that we want to reply
		Transport: "udp",                          // the transport protocol to use
		Port:      uint16(port),                   // the UDP port to be used
		Version:   snmp.Version2c,                 // the SNMP version to use
		Community: envCommunity,                   // the SNMPv2c community string to be used
		Logger:    log.New(os.Stdout, "", 0),      // add a logger
		Timeout:   time.Duration(2) * time.Second, // the timeout from the request in seconds
		MaxOids:   60,                             // max OIDs permitted in a single call, default 60
	}

	// create a socket to utilize
	err := connectionParams.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %v", err)
	}
	defer connectionParams.Conn.Close()

	// latency logging
	var sent time.Time
	connectionParams.OnSent = func(x *snmp.GoSNMP) {
		sent = time.Now()
	}
	connectionParams.OnRecv = func(x *snmp.GoSNMP) {
		log.Println("Query latency in seconds:", time.Since(sent).Seconds())
	}

	// loop instead of goroutine for doing a thing every pollPeriod seconds
	// because the routine didn't work as expected and time is short
	for range time.Tick(time.Second * time.Duration(pollPeriod)) {
		sizeOfRequest := snmpMockRequest(connectionParams, oids)
		sizeOfResponse := snmpPoll(connectionParams, oids)

		// write these as just the integers so we can do CSV format?
		sizes := fmt.Sprintf("%d, %d, %d",
			sizeOfRequest, sizeOfResponse, (sizeOfRequest + sizeOfResponse))
		writeToFile(fileName, sizes)
	}
}

// get the environment variables and turn them into usable variables
// Output:
//	fileName:			the name of the file to write to
//	pollPeriod: 	how often to poll the target, in seconds
//	envTarget:		ip address of target
//	envPort:			the udp port to use
//	envCommunity:	the SNMPv2c community to use
func setup() (string, uint64, string, string, uint64, []string) {
	fileName := createFile()
	writeToFile(fileName, "size of request, size of response, total size")
	envPollPeriod := os.Getenv("XJOB_POLLPERIOD")
	envTarget := os.Getenv("XJOB_SNMP_TARGET")
	envCommunity := os.Getenv("XJOB_SNMP_COMMUNITY")
	envPort := os.Getenv("XJOB_SNMP_PORT")

	if len(envPollPeriod) <= 0 {
		log.Fatal("environment variable not set: XJOB_SNMP_TARGET")
	}
	if len(envTarget) <= 0 {
		log.Fatal("environment variable not set: XJOB_SNMP_TARGET")
	}
	if len(envPort) <= 0 {
		log.Fatal("environment variable not set: XJOB_SNMP_PORT")
	}
	if len(envCommunity) <= 0 {
		log.Fatal("environment variable not set: XJOB_SNMP_COMMUNITY")
	}

	pollPeriod, _ := strconv.ParseUint(envPollPeriod, 10, 16)
	port, _ := strconv.ParseUint(envPort, 10, 16)

	oids := []string{
		/*
			".1.3.6.1.4.1.2021.10.1.3.1", // TEST linux cpu 1 min load
			".1.3.6.1.2.1.1.3.0",         // TEST linux system uptime
			".1.3.6.1.2.1.2.2.1.2.2",     // TEST linux ifName for if@index .2
		*/
		// interface at index .59 added to end of all OIDs
		"1.3.6.1.2.1.31.1.1.1.1.59",  // ifName
		"1.3.6.1.2.1.2.2.1.14.59",    // ifInErrors
		"1.3.6.1.2.1.2.2.1.20.59",    // ifOutErrors
		"1.3.6.1.2.1.2.2.1.15.59",    // IfInUnknownProtos
		"1.3.6.1.2.1.2.2.1.19.59",    // ifOutDiscards
		"1.3.6.1.2.1.2.2.1.13.59",    // ifInDiscards
		"1.3.6.1.2.1.31.1.1.1.6.59",  // ifHCInOctets
		"1.3.6.1.2.1.31.1.1.1.10.59", // ifHCOutOctets
		"1.3.6.1.2.1.31.1.1.1.9.59",  // ifHCInBroadcastPkts
		"1.3.6.1.2.1.31.1.1.1.13.59", // ifHCOutBroadcastPkts
		"1.3.6.1.2.1.31.1.1.1.8.59",  // ifHCInMulticastPkts
		"1.3.6.1.2.1.31.1.1.1.12.59", // ifHCOutMulticastPkts
		"1.3.6.1.2.1.31.1.1.1.7.59",  // ifHCInUcastPkts
		"1.3.6.1.2.1.31.1.1.1.11.59", // ifHCOutUcastPkts
	}

	return fileName, pollPeriod, envTarget, envCommunity, port, oids
}

// craft a mock SnmpPacket to check its size
// Input:
//	connectionparams: that one struct we put together before
//	oids:							the SNMP OIDs that we want to get info about
// Output:
//	sizeOfRequest:		the size in bytes of the mock SNMP request
func snmpMockRequest(connectionParams *snmp.GoSNMP, oids []string) int {
	var pdus []snmp.SnmpPDU
	for _, oid := range oids {
		pdus = append(pdus, snmp.SnmpPDU{oid, snmp.Null, nil})
	}

	// need & before? just wanna make a snmp packet type struct
	mockRequest := snmp.SnmpPacket{
		Version:            0x1,                                 // Version2c SnmpVersion = 0x1
		Community:          connectionParams.Community,          // community from env variable
		MsgFlags:           1,                                   // <- v3
		SecurityModel:      1,                                   // <- v3
		SecurityParameters: connectionParams.SecurityParameters, // <- v3
		ContextEngineID:    connectionParams.ContextEngineID,    // <- v3
		ContextName:        connectionParams.ContextName,        // <- v3
		Error:              0,                                   // SNMPError 0 *should* be the lowest enum i.e. NoError
		ErrorIndex:         0,                                   // uint8, 0 default
		PDUType:            0xa0,                                // GetRequest PDUType = 0xa0
		NonRepeaters:       0,                                   // uint8 0 default
		MaxRepetitions:     (777 & 0x7FFFFFFF),                  // placeholder maxreps
		Variables:          pdus,                                // as specified from above
	}

	// get the byte array form of the packet so we can check the size of it
	requestSize, err := mockRequest.MarshalMsg()
	if err != nil {
		log.Fatal(err)
	}
	sizeOfRequest := len(requestSize)

	return sizeOfRequest
}

// polls a specified target by sending a SNMP Get message and prints the reply
// Input:
//	connectionparams: that one struct we put together before
//	oids: 						the SNMP OIDs that we want to get info about
//	envCommunity: 		the SNMPv2c community string to use
// Output:
//	sizeOfResponse:		the size in bytes of the SNMP response
func snmpPoll(connectionParams *snmp.GoSNMP, oids []string) int {

	// send SNMP GET request with the specified OIDs
	snmpGetResponse, err2 := connectionParams.Get(oids)
	if err2 != nil {
		log.Fatalf("Get() err: %v", err2)
	}

	responseSize, err := snmpGetResponse.MarshalMsg()
	if err != nil {
		log.Fatal(err)
	}
	sizeOfResponse := len(responseSize)

	fmt.Print(unsafe.Sizeof(snmpGetResponse), "\n") // just to see if we can get anything here

	// print the content of result to stdOut
	for i, variable := range snmpGetResponse.Variables {
		fmt.Printf("%d: oid: %s ", i, variable.Name)

		switch variable.Type {
		case snmp.OctetString:
			fmt.Printf("string: %s\n", string(variable.Value.([]byte)))
		default:
			fmt.Printf("number: %d\n", snmp.ToBigInt(variable.Value))
		}
	}

	return sizeOfResponse
}

// Creates a file that is unique to each run
// Output:
//	fileName:	a (hopefully) unique filename in string format
func createFile() string {

	fileName := "xjob_snmp_replies_" + time.Now().Format("2006-01-02T15:04:05Z07:00") + ".txt" // RFC3339 format

	// remove any old duplicate
	err := os.Remove(fileName)
	if err != nil {
		log.Println("Did not delete any old file with same name")
	} else {
		log.Println("Deleted previous dump file")
	}

	// create a new file with the specified filename
	_, err = os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}

	return fileName
}

// Writes the provided input to a file
// Input:
// 	target:	a file to write to
//	input:	a string that shall be written to the file
func writeToFile(target string, input string) error {

	file, err := os.OpenFile(target, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if _, err := file.WriteString(input + "\n"); err != nil {
		log.Fatal(err)
	}
	if err := file.Close(); err != nil {
		log.Fatal(err)
	}
	return nil
}
