## TODOS

add logging to file of replies

can OIDs be called by name or do I need to get them by using the OID.index syntax?
how to call for specific interface using just OID.index? hard-code index?

how to run this thing in a loop and poll every X seconds

reply size in bytes calculated and logged somehow


### Flow
create a file to put replies in
put things in a struct for GoSNMP
connect
pick what OIDs should be gathered
send GET with OIDs
get reply back
log the message text to a file
log the size in bytes of the message somehow (to the same file as the replies?)

## COMMENTS 

attempted writing a simple gosnmp program based on the examples and documentation
gone quite ok so far, need to structure it better and actually test it but i'd prefer to get a green light before putting code on a machine that isnt my own
how do I even put this in runnable form on instat24?
telemetry would be really nice to receive help with because this "small" program turned out to take a lot of time
