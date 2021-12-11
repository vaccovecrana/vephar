# go-smr

Simple application that replicates a key-value store using the Raft protocol.
This application uses the [Hashicorp implementation](https://github.com/hashicorp/raft) of the Raft protocol.

#### Usage:

First execute:

```bash
go build
```

This will create a local executable with name `smr`. Then, with 3 terminal tabs open, execute on each one:

    ./smr -addr=127.0.0.1:8080

Will create an HTTP server on port 8080 and will create a Raft node on port 9090.

    ./smr -addr=127.0.0.1:8081 -raft=127.0.0.1:9091
    
Will create an HTTP server on port 8081 and will create a Raft node on port 9091.

    ./smr -addr=127.0.0.1:8082 -raft=127.0.0.1:9092
    
Will create an HTTP server on port 8082 and will create a Raft node on port 9092.

Those are the default configuration, that can be found on the file `config.json`. Now with the servers available,
is possible to execute requests, like:

    curl 'http://localhost:8081/set?key=Fuck&value=Police'
   
Sending a request to the HTTP server on port 8081 to associate the key `Fuck` with value `Police`. Now to retrieve
a value just execute:

    curl 'localhost:8080/get?key=Fuck'

The value defined on server 8081 is now available on server 8080.
