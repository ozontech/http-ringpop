# Ringpop sidecar for HTTP backend

Initial issue - spread out incoming HTTP requests from different clients between
different instances of application. Requests from one client should be handled 
on the same backend.

This small application solves this problem in following way:
- it works as a proxy (sidecar) in front of your HTTP backend instance
- when it receives request - it decides what backend instance should handle
this request (based on incoming IP)
- if request should be hanlded on local backend - it forwards request to it.
- if request should be hanlded on another instance - it forwards request to its ringpop sidecar.

`Ringpop sidecar for HTTP backend` could be useful if you already have HTTP backend, 
but you need to add sharding / caching / data aggregation.  
Ringpop will take ownership of the scalability and availability.

Requests forwarding based on ringpop approach with gossip under the hood.

This application is based on [Uber's Ringpop](https://eng.uber.com/intro-to-ringpop/).


```text
                          ┌───────────┐                         
                          │   Client  │                         
                          └───────────┘                         
                                │                               
         ┌──────────────────────┼──────────────────────┐        
         │                      │                      │        
         ▼                      ▼                      ▼        
┌────────────────┐     ┌────────────────┐     ┌────────────────┐
│  HTTP Ringpop  │◀───▶│  HTTP Ringpop  │◀───▶│  HTTP Ringpop  │
└────────────────┘     └────────────────┘     └────────────────┘
         │                      │                      │        
         │                      │                      │        
         ▼                      ▼                      ▼        
┌────────────────┐     ┌────────────────┐     ┌────────────────┐
│  HTTP Backend  │     │  HTTP Backend  │     │  HTTP Backend  │
│     shard 1    │     │     shard 2    │     │     shard 3    │
└────────────────┘     └────────────────┘     └────────────────┘
```

## Example

Build
```bash
GO111MODULE=on go build -o simple-backend cmd/backend-example/main.go
GO111MODULE=on go build -o ringpop cmd/ringpop/main.go
```

Run 3 HTTP backends:
```bash
./simple-backend --listen.http=:4000
./simple-backend --listen.http=:4001
./simple-backend --listen.http=:4002
```

Run ringpop on 3 nodes locally:
```bash
./ringpop --listen.http="127.0.0.1:3000" --backend.url="http://127.0.0.1:4000/" --listen.ringpop="127.0.0.1:5000" --listen.debug=":6000" --discovery.json.file=./etc/hosts.json
./ringpop --listen.http="127.0.0.1:3001" --backend.url="http://127.0.0.1:4001/" --listen.ringpop="127.0.0.1:5001" --listen.debug=":6001" --discovery.json.file=./etc/hosts.json
./ringpop --listen.http="127.0.0.1:3002" --backend.url="http://127.0.0.1:4002/" --listen.ringpop="127.0.0.1:5002" --listen.debug=":6002" --discovery.json.file=./etc/hosts.json
```

Test
```bash
curl http://localhost:3000/ -i
```

You will see smth like that (request received by one instance but handled by another):
```bash
HTTP/1.1 200 OK
Content-Length: 50
Content-Type: text/plain; charset=utf-8
X-Ringpop-Handled-By: 127.0.0.1:5002
X-Ringpop-Received-By: 127.0.0.1:5000
```

## Ringpop over Kubernetes

You can find out examples in k8s directory.

If you're running ringpop over Kubernetes note following thing: 
each instance have to know on what real IP it's running (its critical). 
For example, when instances will be discovered from DNS records, it will be something like this:
   10.27.27.42:5000
   10.27.35.133:5000
Current IP could be detected correctly only in particular cases. So, if current IP will be
detected automatically as 127.0.0.1 this node will try to join to itself.

Also, one important thing is that you should use `ClusterIP: none` for service, 
that will be used for DNS discovery, because in this case `nslookup myawesomeservice` 
will return list of A-records for all service endpoints.


## Dockerization

```bash
make build
docker build -t http-ringpop:1.0.0 .
```

Pre-built image on Dockerhub: [ozonru/http-ringpop:1.0.0](https://hub.docker.com/r/ozonru/http-ringpop)
