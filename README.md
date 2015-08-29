# IP Reflector

For SSHing into a server with a dynamic IP.

Useage:

    reflect_ip <API key> <Sleep time (Seconds)>
    
To build it, first `go get` the websocket library:

    go get golang.org/x/net/websocket

Then build the project:

    go build
    
Boom.

Comes with both a Darwin/ OS X and ARM Linux binary.
