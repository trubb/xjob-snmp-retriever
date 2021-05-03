package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"unsafe"

	g "github.com/gosnmp/gosnmp"
)

func main() {

	// create a file to log replies to
	filename := createFile()

	// how often to poll the target, in seconds
	envPollPeriod := os.Getenv("XJOB_POLLPERIOD")
	pollPeriod, _ := strconv.ParseUint(envPollPeriod, 10, 16)

	////////////////////////////////////////////////////
	// begin gosnmp example code stuff
	////////////////////////////////////////////////////

	// get Target, Port, and Community from environment vars to avoid leaking info via source code
	envTarget := os.Getenv("XJOB_SNMP_TARGET")
	envPort := os.Getenv("XJOB_SNMP_PORT")
	envCommunity := os.Getenv("XJOB_SNMP_COMMUNITY")
	if len(envTarget) <= 0 {
		log.Fatal("environment variable not set: XJOB_SNMP_TARGET")
	}
	if len(envPort) <= 0 {
		log.Fatal("environment variable not set: XJOB_SNMP_PORT")
	}
	if len(envCommunity) <= 0 {
		log.Fatal("environment variable not set: XJOB_SNMP_COMMUNITY")
	}
	port, _ := strconv.ParseUint(envPort, 10, 16)

	// put together a struct containing connection parameters
	connectionParams := &g.GoSNMP{
		Target:    envTarget,                      // the network node that we want to reply
		Transport: "udp",                          // the transport protocol to use
		Port:      uint16(port),                   // the UDP port to be used
		Version:   g.Version2c,                    // the SNMP version to use
		Community: envCommunity,                   // the SNMPv2c community string to be used
		Logger:    log.New(os.Stdout, "", 0),      // add a logger
		Timeout:   time.Duration(2) * time.Second, // the timeout from the request in seconds
		MaxOids:   60,                             // max OIDs permitted in a single call, 60 "seems to be a common value that works"
	}

	// create a socket to utilize
	err := connectionParams.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %v", err)
	}
	defer connectionParams.Conn.Close()

	// latency logging
	var sent time.Time
	connectionParams.OnSent = func(x *g.GoSNMP) {
		sent = time.Now()
	}
	connectionParams.OnRecv = func(x *g.GoSNMP) {
		log.Println("Query latency in seconds:", time.Since(sent).Seconds())
	}

	// retrieve list of OIDs
	oids := selectOidArr("numbers")

	////////////////////////////////////////////////////
	// this part below is the one we want to do repeatedly
	////////////////////////////////////////////////////

	// goroutine for doing a thing every pollPeriod seconds
	go func() {
		for range time.Tick(time.Second * time.Duration(pollPeriod)) {
			//			snmpPoll(connectionParams, oids, filename)

			// send SNMP GET request with the specified OIDs
			result, err2 := connectionParams.Get(oids)
			if err2 != nil {
				log.Fatalf("Get() err: %v", err2)
			}

			// craft a SnmpPacket to check its size

			fmt.Print(connectionParams.Version)

			var pdus []g.SnmpPDU
			for _, oid := range oids {
				pdus = append(pdus, g.SnmpPDU{oid, g.Null, nil})
			}

			// need & before? just wanna make a snmp packet type struct
			request := g.SnmpPacket{
				Version:            0x1,                                 // Version2c SnmpVersion = 0x1
				Community:          envCommunity,                        // community from env variable
				MsgFlags:           1,                                   // <- v3, need to figure out how to ignore gracefully
				SecurityModel:      1,                                   // <- v3, need to figure out how to ignore gracefully
				SecurityParameters: connectionParams.SecurityParameters, // an INTERFACE, need to figure out how to ignore gracefully
				ContextEngineID:    connectionParams.ContextEngineID,    // <- v3, need to figure out how to ignore gracefully
				ContextName:        connectionParams.ContextName,        // <- v3, need to figure out how to ignore gracefully
				Error:              0,                                   // SNMPError 0 *should* be the lowest enum i.e. NoError
				ErrorIndex:         0,                                   // uint8, 0 default
				PDUType:            0xa0,                                // GetRequest PDUType = 0xa0
				NonRepeaters:       0,                                   // uint8 0 default
				MaxRepetitions:     (777 & 0x7FFFFFFF),                  // placeholder maxreps
				Variables:          pdus,                                // got them from above, hopefully they work...
			}

			// get the byte array form of the packet so we can check the size of it
			requestSize, err3 := request.MarshalMsg()
			if err3 != nil {
				log.Fatal(err)
			}

			resultSize, err := result.MarshalMsg()
			if err != nil {
				log.Fatal(err)
			}

			fmt.Print(resultSize)
			fmt.Print(unsafe.Sizeof(result))

			writeToFile(
				filename,
				"size of request: "+string(requestSize)+
					", size of response: "+string(resultSize)+
					", total size: "+string((requestSize+resultSize)))

			// print the content of result to stdOut
			for i, variable := range result.Variables {
				fmt.Printf("%d: oid: %s ", i, variable.Name)

				switch variable.Type {
				case g.OctetString:
					fmt.Printf("string: %s\n", string(variable.Value.([]byte)))
				default:
					fmt.Printf("number: %d\n", g.ToBigInt(variable.Value))
				}
			}

		}
	}()

}

// polls a specified target by sending a SNMP Get message and prints the reply
// input:
//	connectionparams: that one struct we put together before
//	oids: the SNMP OIDs that we want to get info about

// moved most of the code up above to avoid having to bombard this function with variables..
func snmpPoll(connectionParams *g.GoSNMP, oids []string, filename string) {

	// send SNMP GET request with the specified OIDs
	result, err2 := connectionParams.Get(oids)
	if err2 != nil {
		log.Fatalf("Get() err: %v", err2)
	}

	packetOut := x.mkSnmpPacket(GetRequest, pdus, 0, 0)

	resultSize, err := result.MarshalMsg()
	if err != nil {
		log.Fatal(err)
	}

	writeToFile(
		filename,
		"size of request: "+string(requestSize)+
			", size of response: "+string(resultSize)+
			", total size: "+string((requestSize+resultSize)))

	fmt.Print(resultSize)
	fmt.Print(unsafe.Sizeof(result))

	// print the content of result to stdOut
	for i, variable := range result.Variables {
		fmt.Printf("%d: oid: %s ", i, variable.Name)

		switch variable.Type {
		case g.OctetString:
			fmt.Printf("string: %s\n", string(variable.Value.([]byte)))
		default:
			fmt.Printf("number: %d\n", g.ToBigInt(variable.Value))
		}
	}

}

// Creates a file that is unique to each run
// Input: none
// Output: a (hopefully) unique filename in string format
func createFile() string {

	filename := "xjob_snmp_replies_" + time.Now().Format("2006-01-02T15:04:05Z07:00") + ".txt" // RFC3339 format

	// remove any old duplicate
	err := os.Remove(filename)
	if err != nil {
		log.Println("Did not delete any old file with same name")
	} else {
		log.Println("Deleted previous dump file")
	}

	// create a new file with the specified filename
	_, err = os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	return filename
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

// Pick what OIDs we should use
// Highly unclear if we even need this but it's nice to have hidden down here
// Input:
//	oidFormat: a string stating what OID "format" to use
func selectOidArr(oidFormat string) []string {

	// OIDs in dot separated form
	oids := []string{
		"1.3.6.1.2.1.31.1.1.1.1",  // ifName
		"1.3.6.1.2.1.2.2.1.14",    // ifInErrors
		"1.3.6.1.2.1.2.2.1.20",    // ifOutErrors
		"1.3.6.1.2.1.2.2.1.15",    // IfInUnknownProtos
		"1.3.6.1.2.1.2.2.1.19",    // ifOutDiscards
		"1.3.6.1.2.1.2.2.1.13",    // ifInDiscards
		"1.3.6.1.2.1.31.1.1.1.6",  // ifHCInOctets
		"1.3.6.1.2.1.31.1.1.1.10", // ifHCOutOctets
		"1.3.6.1.2.1.31.1.1.1.9",  // ifHCInBroadcastPkts
		"1.3.6.1.2.1.31.1.1.1.13", // ifHCOutBroadcastPkts
		"1.3.6.1.2.1.31.1.1.1.8",  // ifHCInMulticastPkts
		"1.3.6.1.2.1.31.1.1.1.12", // ifHCOutMulticastPkts
		"1.3.6.1.2.1.31.1.1.1.7",  // ifHCInUcastPkts
		"1.3.6.1.2.1.31.1.1.1.11", // ifHCOutUcastPkts
	}

	// OIDs in named form
	oidNames := []string{
		"ifName",               // 1.3.6.1.2.1.31.1.1.1.1
		"ifInErrors",           // 1.3.6.1.2.1.2.2.1.14
		"ifOutErrors",          // 1.3.6.1.2.1.2.2.1.20
		"IfInUnknownProtos",    // 1.3.6.1.2.1.2.2.1.15
		"ifOutDiscards",        // 1.3.6.1.2.1.2.2.1.19
		"ifInDiscards",         // 1.3.6.1.2.1.2.2.1.13
		"ifHCInOctets",         // 1.3.6.1.2.1.31.1.1.1.6
		"ifHCOutOctets",        // 1.3.6.1.2.1.31.1.1.1.10
		"ifHCInBroadcastPkts",  // 1.3.6.1.2.1.31.1.1.1.9
		"ifHCOutBroadcastPkts", // 1.3.6.1.2.1.31.1.1.1.13
		"ifHCInMulticastPkts",  // 1.3.6.1.2.1.31.1.1.1.8
		"ifHCOutMulticastPkts", // 1.3.6.1.2.1.31.1.1.1.12
		"ifHCInUcastPkts",      // 1.3.6.1.2.1.31.1.1.1.7
		"ifHCOutUcastPkts",     // 1.3.6.1.2.1.31.1.1.1.11
	}

	if oidFormat == "numbers" {
		return oids
	} else if oidFormat == "named" {
		return oidNames
	} else {
		return nil
	}
}
