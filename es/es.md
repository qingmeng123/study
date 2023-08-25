# olivere/elastic/v7

## 含有重试和健康检查
- olivere/elastic/v7客户端默认未设定重试,需要配置retrier和retryStatusCodes
```go
//默认无
retrier:                   noRetries, // no retries by default
retryStatusCodes:          nil,       // no automatic retries for specific HTTP status codes
//配置retrier
client, err := elastic.NewClient(elastic.SetURL("http://127.0.0.1:9200"), elastic.SetSniff(false),
elastic.SetRetrier(
elastic.NewBackoffRetrier(elastic.NewSimpleBackoff(10000)), // 设置重试策略，框架自带几种
),
elastic.SetRetryStatusCodes(504),
)
```

- 他会定期进行健康检查，client包中healthcheck方法。
healthcheck 对集群中的所有节点进行健康检查。根据节点状态，它将连接标记为死亡、将其设置为活动等。
如果健康检查被禁用并且强制为 false，则这是无操作。超时指定等待 Elasticsearch 响应的时间


- 如果所有节点都被标记为死亡,则会在内部处理时重新标为alive(关闭sniffer时)，返回ErrNoClient，后续处理ErrNoClient
```go
// next returns the next available connection, or ErrNoClient.
func (c *Client) next() (*conn, error) {
	// We do round-robin here.
	// TODO(oe) This should be a pluggable strategy, like the Selector in the official clients.
	c.connsMu.Lock()
	defer c.connsMu.Unlock()

	i := 0
	numConns := len(c.conns)
	for {
		i++
		if i > numConns {
			break // we visited all conns: they all seem to be dead
		}
		c.cindex++
		if c.cindex >= numConns {
			c.cindex = 0
		}
		conn := c.conns[c.cindex]
		if !conn.IsDead() {
			return conn, nil
		}
	}

	//我们这里遇到了死锁：所有节点都被标记为死亡。
	//如果禁用嗅探，连接将永远不会再次被标记为活动状态。
	//因此，如果嗅探被禁用，我们将它们标记为活着。然后，将在下一次调用 PerformRequest 时获取它们。
	if !c.snifferEnabled {
		c.errorf("elastic: all %d nodes marked as dead; resurrecting them to prevent deadlock", len(c.conns))
		for _, conn := range c.conns {
			conn.MarkAsAlive()
		}
	}

	// We tried hard, but there is no node available
	return nil, errors.Wrap(ErrNoClient, "no available connection")
}
```

- 发起请求时，客户端会先判断连接是否正常，若ErrNoClient，则根据设定的Retrier和retryStatusCodes尝试重试
```go
// client包中，PerformRequest方法，找不到节点时，调用重试
    for {
        pathWithParams := opt.Path
		if len(opt.Params) > 0 {
			pathWithParams += "?" + opt.Params.Encode()
		}

		// Get a connection
		conn, err = c.next()
		if errors.Cause(err) == ErrNoClient {
			n++
			if !retried {
				// Force a healtcheck as all connections seem to be dead.
				c.healthcheck(ctx, timeout, false)
				if healthcheckEnabled {
					retried = true
					continue
				}
			}
			wait, ok, rerr := retrier.Retry(ctx, n, nil, nil, err)
			if rerr != nil {
				return nil, rerr
			}
			if !ok {
				return nil, err
			}
			retried = true
			time.Sleep(wait)
			continue // try again
		}
		if err != nil {
			c.errorf("elastic: cannot get connection from pool")
			return nil, err
		}

		req, err = NewRequest(opt.Method, conn.URL()+pathWithParams)
		if err != nil {
			c.errorf("elastic: cannot create request for %s %s: %v", strings.ToUpper(opt.Method), conn.URL()+pathWithParams, err)
			return nil, err
		}  
    }
```

## 重连测试
### 单节点
```go
type Person struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Married bool   `json:"married"`
}

func main() {
	errorlog := log.New(os.Stdout, "APP", log.LstdFlags)
	nodeURLs := []string{"http://127.0.0.1:9200"}

	client, err := elastic.NewClient(elastic.SetURL(nodeURLs...),
		elastic.SetSniff(false),
		elastic.SetHealthcheckTimeout(2*time.Second),
		elastic.SetRetrier(
			elastic.NewBackoffRetrier(NewMyBackoff(time.Second)), // 设置重试策略
		),
		elastic.SetRetryStatusCodes(504),
		elastic.SetErrorLog(errorlog),
	)

	if err != nil {
		// Handle error
		log.Fatal("connect err:", err)
		return
	}

	fmt.Println("connect to es success")

	//等待5s,关闭es服务器，查看重试结果后重启服务器
	time.Sleep(5 * time.Second)

	//请求，测试重试
	p1 := Person{Name: "lmh", Age: 18, Married: false}
	put1, err := client.Index().
		Index("user").
		BodyJson(p1).
		Do(context.Background())
	if err != nil {
		// Handle error
		panic(err)
	}
	fmt.Printf("Indexed user %s to index %s, type %s\n", put1.Id, put1.Index, put1.Type)

}

// 自定义回退策略，固定时长无限制重试，中间会有健康检查日志说明，但单节点会重新标记为alive后继续重试
type MyBackoff struct {
	interval time.Duration
}

func NewMyBackoff(interval time.Duration) *MyBackoff {
	return &MyBackoff{interval: interval}
}

// Next implements BackoffFunc for MyBackoff.
func (b *MyBackoff) Next(retry int) (time.Duration, bool) {
	log.Println("try ", retry)
	return b.interval, true
}

//输出结果
/*
connect to es success
//关闭es
2023/08/09 14:16:53 try  1
2023/08/09 14:16:56 try  2

2023/08/09 14:17:41 try  17
2023/08/09 14:17:44 try  18
2023/08/09 14:17:47 try  19
//定期的健康检查
APP2023/08/09 14:17:48 elastic: http://127.0.0.1:9200 is dead
APP2023/08/09 14:17:49 elastic: all 1 nodes marked as dead; resurrecting them to prevent deadlock
2023/08/09 14:18:28 try  35
2023/08/09 14:18:29 try  36
2023/08/09 14:18:30 try  37
//开启es
2023/08/09 14:18:31 try  38
2023/08/09 14:18:32 try  39
Indexed user DLXy2IkB9fxddsoja9g2 to index user, type _doc*/

```


