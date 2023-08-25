package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"

	"os"

	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	r  map[string]interface{}
	wg sync.WaitGroup
)

func customBackoff(retries int) time.Duration {
	// 这里可以根据重试次数来计算退避时间间隔，这个示例中简单地返回 1 秒
	return time.Second
}

func main1() {

	// 创建 Elasticsearch 客户端
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
		MaxRetries: 5, //默认为3
		//EnableRetryOnTimeout: true, //默认false
		RetryBackoff: customBackoff,
		/*Transport: &http.Transport{
			ResponseHeaderTimeout: time.Second * 10,
		},*/
		// 使用标准日志记录器，设置 Debug 模式
		Logger: &customLogger{
			logger: log.New(os.Stdout, "custom-logger ", log.LstdFlags),
		},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Println("Error creating the client: ", err)
		return
	}

	// 创建 Elasticsearch Info 请求
	req := esapi.InfoRequest{}

	// 模拟连接失败
	// 关闭 Elasticsearch，或者设置错误的地址，来模拟连接失败的情况
	// 这会触发客户端的重连机制
	// ...

	log.Println("Attempting to connect...")

	res, err := req.Do(context.Background(), es) // 将 context 传递给 Do 方法
	if err != nil {
		log.Printf("Error getting response: %s", err)
		// 此处可以处理连接错误，或者超时错误
	} else {
		defer res.Body.Close()
		if res.IsError() {
			log.Printf("Error: %s", res.String())
		} else {
			// 解析响应并处理
			log.Printf("Response: %s", res.String())
		}
	}

	//index(es)

	//update
	//update(es)

}

type customLogger struct {
	logger *log.Logger
}

func (cl *customLogger) LogRoundTrip(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) error {
	if err != nil {
		cl.logger.Printf("Request failed to %s: %s", req.URL, err)
	} else {
		cl.logger.Printf("Request to %s took %s", req.URL, dur)
	}
	return nil
}

func (cl *customLogger) Printf(format string, v ...interface{}) {
	cl.logger.Printf(format, v...)
}

func (cl *customLogger) RequestBodyEnabled() bool {
	return true
}

func (cl *customLogger) ResponseBodyEnabled() bool {
	return true
}

func index(es *elasticsearch.Client) {
	// 2. Index documents concurrently
	//
	for i, title := range []string{"Test One", "Test Two"} {
		wg.Add(1)

		go func(i int, title string) {
			defer wg.Done()

			// Build the request body.
			data, err := json.Marshal(struct {
				Title string `json:"title"`
			}{Title: title})
			if err != nil {
				log.Fatalf("Error marshaling document: %s", err)
			}

			// Set up the request object.
			req := esapi.IndexRequest{
				Index:      "test",
				DocumentID: strconv.Itoa(i + 1),
				Body:       bytes.NewReader(data),
				Refresh:    "true",
			}

			// Perform the request with the client.
			res, err := req.Do(context.Background(), es)
			if err != nil {
				log.Fatalf("Error getting response: %s", err)
			}
			defer res.Body.Close()

			if res.IsError() {
				log.Printf("[%s] Error indexing document ID=%d", res.Status(), i+1)
			} else {
				// Deserialize the response into a map.
				var r map[string]interface{}
				if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
					log.Printf("Error parsing the response body: %s", err)
				} else {
					// Print the response status and indexed document version.
					log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
				}
			}
		}(i, title)
	}
	wg.Wait()

	log.Println(strings.Repeat("-", 37))
}

func update1(es *elasticsearch.Client) {
	dat := map[string]interface{}{
		"doc": map[string]interface{}{
			"title": "test 你好",
		},
	}
	data, err := json.Marshal(dat)

	if err != nil {
		log.Fatalf("Error marshaling document: %s", err)
	}

	req := esapi.UpdateRequest{
		Index:      "test",
		DocumentID: "2",
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		log.Printf("Error update document: %s", res.String())
	} else {
		// 解析响应并打印结果
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Printf("Error parsing response body: %s", err)
		} else {
			log.Printf("Document update: %s", r["result"])
		}
	}
	log.Println(strings.Repeat("-", 37))

	//delete
	res, err = es.Delete("test", "1", es.Delete.WithRefresh("true"))
	if err != nil {
		log.Fatal("delete err:", err)
	}

	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		log.Printf("Error deleting document: %s", res.String())
	} else {
		// 解析响应并打印结果
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Printf("Error parsing response body: %s", err)
		} else {
			log.Printf("Document deleted: %s", r["result"])
		}
	}
	log.Println(strings.Repeat("-", 37))

	// 3. Search for the indexed documents
	//
	// Build the request body.
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"title": "test",
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	// Perform the search request.
	res, err = es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("test"), //索引
		es.Search.WithBody(&buf),    //匹配doc中title包含test的
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	// Print the response status, number of results, and request duration.
	log.Printf(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(r["took"].(float64)),
	)
	// Print the ID and document source for each hit.
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
	}

	log.Println(strings.Repeat("=", 37))
}
