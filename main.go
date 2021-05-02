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
	envPollPeriod := os.Getenv("XJOB_POLLPERIOD") // need to be converted to integer?
	pollPeriod, _ := strconv.ParseUint(envPollPeriod, 10, 16)

	// just make sure they aren't screamed at for the time being
	fmt.Print(filename)

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
		Port:      uint16(port),                   // the UDP port to be used
		Version:   g.Version2c,                    // the SNMP version to use
		Community: envCommunity,                   // the SNMPv2c community string to be used
		Logger:    log.New(os.Stdout, "", 0),      // add a logger TODO this might not be wanted?
		Timeout:   time.Duration(2) * time.Second, // the timeout from the request in seconds
		MaxOids:   60,                             // max OIDs permitted in a single call, 60 "seems to be a common value that works"
	}

	// connect using the connection parameters
	// WHAT?!!! We're using UDP but still connect?
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
			snmpPoll(connectionParams, oids)
		}
	}()

}

// polls a specified target by sending a SNMP Get message and prints the reply
// input:
//	connectionparams: that one struct we put together before
//	oids: the SNMP OIDs that we want to get info about
func snmpPoll(connectionParams *g.GoSNMP, oids []string) {

	// send SNMP GET request with the specified OIDs
	result, err2 := connectionParams.Get(oids)
	if err2 != nil {
		log.Fatalf("Get() err: %v", err2)
	}

	// print the size of the struct
	// TODO do that cool calculation thing!
	fmt.Print(unsafe.Sizeof(result))

	// print the result to stdOut
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
