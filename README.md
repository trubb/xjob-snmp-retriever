# xjob-snmp-retriever

A simple SNMP utility program based on [GoSNMP](https://github.com/gosnmp/gosnmp) and using [getopt](https://github.com/pborman/getopt/) to retrieve arguments.  
Created for generating SNMP traffic and data gathering as part of a thesis project.

## Compilation

To compile it for use in a Linux environment:
```
GOOS=linux GOARCH=amd64 go build -o xjob-snmp-retriever
```

## Use

Run the resulting executable file with arguments:
```
./xjob-snmp-retriever [--community value] [--pollperiod value] [--port value] [--target value]
     --community=value
                   The snmp v2c community to use
     --pollperiod=value
                   How often to poll the target (seconds) [60]
     --port=value  
                   The udp port to use for polling [161]
     --target=value
                   The target host to poll
```

Like this:
```
./snmp-retriever --pollPeriod 5 --target 127.0.0.1 --community myCoolCommunity --port 1337
```
