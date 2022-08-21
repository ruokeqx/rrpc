day1——server

```go
(server *Server) Accept
->for{
    conn, err := lis.Accept()
    go server.ServeConn(conn)
        ->server.serveCodec(f(conn))
            ->for{
                req, err := server.readRequest(cc)
                go server.handleRequest(cc, req, sending, wg)
            }
}
```

day2——client

```go
Dial
->NewClient
    ->newClientCodec
        ->go client.receive()
            ->client.cc.ReadBody(call.Reply)
// 同步发送
(client *Client).Call
// 异步发送
->(client *Client).Go
    ->(client *Client).send
```

day3——service register
结构体映射为服务(使用反射)

```go
type service struct {
	name   string
	typ    reflect.Type
	rcvr   reflect.Value
	method map[string]*methodType
}
type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	numCalls  uint64
}
newService(&foo)
->s.registerMethods()
(s *service) call(m *methodType, argv, replyv reflect.Value)
```

