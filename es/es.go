package main

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

type Person struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Married bool   `json:"married"`
}

func main() {
	errorlog := log.New(os.Stdout, "APP", log.LstdFlags)
	nodeURLs := []string{"http://127.0.0.1:9200", "http://127.0.0.1:9201", "http://127.0.0.1:9202"}

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
	//time.Sleep(5 * time.Second)

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

func connect() {
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

	options := elastic.PerformRequestOptions{
		Method: "GET",
		Path:   "/_nodes/http",

		ContentType: "application/json",
		Headers: http.Header{
			"User-Agent": {"elastic/" + "7.0.32" + " (" + runtime.GOOS + "-" + runtime.GOARCH + ")"},
		},
	}

	res, err := client.PerformRequest(context.Background(), options)
	if err != nil {
		log.Print("perform err:", err)
		return
	}
	fmt.Println("res:", res)
}
