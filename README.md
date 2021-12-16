# vephar

[vephar](https://en.wikipedia.org/wiki/List_of_demons_in_the_Ars_Goetia) is a minimal K/V store with an API and a dashboard using:

- [hashicorp-raft](https://github.com/hashicorp/raft) for cluster coordination.
- [badger](https://github.com/dgraph-io/badger) for Raft logs and data storage.
- [raft-badger](https://github.com/markthethomas/raft-badger) with slight modifications for current `badger` compatibility.

## Usage

3 node cluster:

Nodes in the cluster identify with the following format:

    IPV4_ADDRESS:RAFT_TCP_PORT:HTTP_PORT

Start node 1:

    ./vephar -peerId=127.0.0.1:9090:8080 -data=./data0 \
      -join=127.0.0.1:9091:8081,127.0.0.1:9092:8082

Start node 2:

    ./vephar -peerId=127.0.0.1:9091:8081 -data=./data1 \
      -join=127.0.0.1:9090:8080,127.0.0.1:9092:8082
    
Start node 3:

    ./vephar -peerId=127.0.0.1:9092:8082 -data=./data2 \
      -join=127.0.0.1:9090:8080,127.0.0.1:9091:8081,127.0.0.1:9092:8082

Now set a key and a value:

```
curl -o - 'http://127.0.0.1:8081/kv/set?key=Hello&value=World'
{"Data":"Hello","Error":""}
```

And rertrieve it:

```
curl -o - 'http://127.0.0.1:8081/kv/get?key=Hello'
World
```

This also works for binary data values:

```
curl --request POST 'http://127.0.0.1:8080/kv/set?key=Hello' \
     --header 'Content-Type: text/plain' \
     --data-binary '@LICENSE'
{"Data":"Hello","Error":""}
```

Now delete the key:

```
curl -o - 'http://127.0.0.1:8081/kv/del?key=Hello'
{"Data":"Hello","Error":""}

curl -o - 'http://127.0.0.1:8081/kv/get?key=Hello'
{"Data":null,"Error":"key not found: [Hello]"}
```

The keys and values defined on server `8081` are also available on servers `8080` and `8082`.

These operations can also be done with the integrated web UI:

http://127.0.0.1:8081/ui

The environment variables `VPR_TRACE` and `VPR_DEBUG` can be used to log a node's execution state.
The variable values are not read, and the program only checks if they have been defined in the environment.

## Similar projects

- [go-smr](https://github.com/jabolina/go-smr)
- [dragonboat-example](https://github.com/lni/dragonboat-example)
- [raftdb](https://github.com/hslam/raftdb)
- [m2](https://github.com/qichengzx/m2)
- [raftdb](https://github.com/hanj4096/raftdb)

# Disclaimer

> This project is not production ready and still requires security and code correctness audits.
> You use this software at your own risk. Vaccove Crana, LLC., its affiliates and subsidiaries
> waive any and all liability for any damages caused to you by your usage of this software.
