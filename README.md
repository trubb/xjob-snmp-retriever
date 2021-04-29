# xjob-snmp-retriever
A simple SNMP utility using GoSNMP created for data gathering as part of a thesis project.

Run the program with environment variables provided:
`XJOB_POLLPERIOD=<seconds> XJOB_SNMP_TARGET=<target ip> XJOB_SNMP_PORT=<target port> XJOB_SNMP_COMMUNITY=<community> go run main.go`
Such as:
`XJOB_POLLPERIOD=60 XJOB_SNMP_TARGET=127.0.0.1 XJOB_SNMP_PORT=25051 XJOB_SNMP_COMMUNITY=Public go run main.go`
