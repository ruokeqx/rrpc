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

day4——timeout

这里有大问题
* 一是只处理了server本地call的timeout 没有处理sendResponse的timeout(当然err log算吧)
* 二是会存在goroutine泄露的问题 因为timeout之后client一个chan阻塞；server两个chan都阻塞

解决方法就是加一个finish chan close后让goroutine return(具体见代码)

```go
// user control ctx and thus timeout
func (client *Client) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	call := client.Go(serviceMethod, args, reply, make(chan *Call, 1))
	select {
	case <-ctx.Done():
		client.removeCall(call.Seq)
		return errors.New("rpc client: call failed: " + ctx.Err().Error())
	case call := <-call.Done:
		return call.Error
	}
}
// server timeout called sent
func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()
	called := make(chan struct{})
	sent := make(chan struct{})
	go func() {
		err := req.svc.call(req.mtype, req.argv, req.replyv)
		called <- struct{}{}
		if err != nil {
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			sent <- struct{}{}
			return
		}
		server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
		sent <- struct{}{}
	}()

	if timeout == 0 {
		<-called
		<-sent
		return
	}
	select {
	case <-time.After(timeout):
		req.h.Error = fmt.Sprintf("rpc server: request handle timeout: expect within %s", timeout)
		server.sendResponse(cc, req.h, invalidRequest, sending)
	case <-called:
		<-sent
	}
}
```
