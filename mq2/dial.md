## 分析Dial
- 方法func (d *Dialer) DialContext(ctx context.Context, network, address string) (Conn, error) {}
```go
deadline := d.deadline(ctx, time.Now())
	if !deadline.IsZero() {
		if d, ok := ctx.Deadline(); !ok || deadline.Before(d) {
			subCtx, cancel := context.WithDeadline(ctx, deadline)
			defer cancel()
			ctx = subCtx
		}
	}
```
解析：使用 Dialer (d) 的 deadline 方法计算了上下文 (ctx) 的截止时间。如果计算得到的截止时间不为零，并且在原始上下文的截止时间之前（如果存在的话），则创建一个新的子上下文 (subCtx)，带有调整后的截止时间，并使用 defer cancel() 延迟执行取消操作。最后，将原始上下文替换为这个新的子上下文。这是为了确保连接尊重提供的上下文的截止时间。  
作用：这样的处理是为了确保连接的操作尊重上下文的截止时间。如果连接的操作需要一些时间，而上下文的截止时间在此之前到期，新的子上下文会在预定的时间内取消连接的操作，以避免连接在上下文的要求之后继续执行。这种方式确保了连接操作的安全性和可控性。

```go
	if oldCancel := d.Cancel; oldCancel != nil {
		subCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		go func() {
			select {
			case <-oldCancel:
				cancel()
			case <-subCtx.Done():
			}
		}()
		ctx = subCtx
	}
```
解析：如果存在 Cancel 字段，它创建了一个新的子上下文 (subCtx) 并与原始上下文 (ctx) 一同取消。然后启动了一个 goroutine，在旧的 Cancel 触发或子上下文完成时取消新的 subCtx。最后，将上下文替换为新的子上下文。

```go
//阴影拷贝
// Shadow the nettrace (if any) during resolve so Connect events don't fire for DNS lookups.
resolveCtx := ctx
	if trace, _ := ctx.Value(nettrace.TraceKey{}).(*nettrace.Trace); trace != nil {
		shadow := *trace
		shadow.ConnectStart = nil
		shadow.ConnectDone = nil
		resolveCtx = context.WithValue(resolveCtx, nettrace.TraceKey{}, &shadow)
	}
```
解析：“阴影拷贝”指的是通过复制某个结构的值，创建一个新的结构副本。

作用：使用阴影拷贝的方式可以确保修改操作不影响原始对象，从而提高并发程序的稳定性和可维护性。在这个特定的上下文中，ConnectStart 和 ConnectDone 字段被设置为 nil，可能是因为在某些情况下，不希望在解析过程中触发与连接事件相关的跟踪。如果直接在原始对象上修改，可能会导致对其他部分代码的影响，而不是只在当前上下文中使用这些字段。
```go
	addrs, err := d.resolver().resolveAddrList(resolveCtx, "dial", network, address, d.LocalAddr)
	if err != nil {
		return nil, &OpError{Op: "dial", Net: network, Source: nil, Addr: nil, Err: err}
	}
```
解析：通过resolver解析出可能存在的多个可用地址,对DNS解析采用轮询

作用：网络通信中，有时候一个域名可能对应多个IP地址,或者在某些情况下，一个服务可能有多个可用的网络地址。因此，resolveAddrList 函数负责解析指定的网络地址，可能会返回多个可用的地址。
```go
	sd := &sysDialer{
		Dialer:  *d,
		network: network,
		address: address,
	}

	var primaries, fallbacks addrList
	if d.dualStack() && network == "tcp" {
		primaries, fallbacks = addrs.partition(isIPv4)
	} else {
		primaries = addrs
	}

	return sd.dialParallel(ctx, primaries, fallbacks)
```
解析：创建系统拨号器，检查 Dialer (d) 是否启用了 dual stack 并且网络类型是否为 "tcp"。根据条件将解析得到的地址列表分为主地址 (primaries) 和备用地址 (fallbacks)。
调用 sd.dialParallel 方法，该方法实现了并行拨号，尝试在多个地址上进行并发连接。返回通过 dialParallel 方法得到的连接和可能的错误。如果发生错误，会返回一个带有错误信息的 OpError。

作用：dualStack ：双栈，在网络编程中，一些应用可能需要同时支持 IPv4 和 IPv6。dualStack 特性允许程序在支持 IPv6 的网络环境中，同时使用 IPv4 和 IPv6 地址。通过时间差判断是否需要支持两者