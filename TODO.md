## TODOS

* [] add logging to file of SNMP replies
* [] calculate size of reply


## TEST

* [] polling loop


## DONE

* [x] get env vars
* [x] createFile
* [x] writeToFile
* [x] selectOidArr
* [x] create socket
* [x] send SNMP GET
* [x] get response
* [x] post response to stdOut


### FLOW

create a file to put replies in
put things in a struct for GoSNMP
connect
pick what OIDs should be gathered

In a goroutine that fires every X seconds {  
  send GET with OIDs  
  get reply back  
  log the message text to a file  
}  
use a ticker instead of that goroutine? https://gobyexample.com/tickers

(log the size in bytes of the message somehow (to the same file as the replies?))


## COMMENTS & QUESTIONS

how do I put this in runnable form on instat24?

can OIDs be called by name or do I need to get them by using the OID.index syntax?
how to call for specific interface using just OID.index? hard-code index or use ENV var?

how to run this thing in a loop and poll every X seconds

reply size in bytes calculated and logged somehow

not sure if getting the size of the reply SnmpPacket is the correct move?
should maybe be getting the size of the parts of the struct?

`SnmpPacket struct represents the entire SNMP Message or Sequence at the application layer.`
soooo it's a decoded message?  
how much overhead is added by a struct in go?
