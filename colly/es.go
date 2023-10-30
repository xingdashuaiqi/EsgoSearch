package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/olivere/elastic/v7"
)

var client *elastic.Client

func esMain() {
	var err error

	// 创建 Elasticsearch 客户端连接
	client, err := elastic.NewClient(elastic.SetURL("http://localhost:9200"))
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %v", err)
		return
	}

	// 创建索引（如果不存在）
	indexName := "news_index" // 替换为您的索引名称

	// 检查索引是否存在
	exists, err := client.IndexExists(indexName).Do(context.Background())
	if err != nil {
		log.Fatalf("Error checking index existence: %v", err)
		return
	}

	if !exists {
		// 如果索引不存在，创建索引并定义映射
		createIndex, err := client.CreateIndex(indexName).BodyString(`
        {
            "mappings": {
                "properties": {
                    "content": {
                        "type": "text"
                    }
                }
            }
        }
        `).Do(context.Background())
		if err != nil {
			log.Fatalf("Error creating index: %v", err)
			return
		}
		if !createIndex.Acknowledged {
			log.Fatalf("Index creation not acknowledged")
			return
		}
	}

	var data []string

	// 读取 news_data.txt 文件的内容并存储在内存中
	file, err := os.Open("news_data.txt")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		data = append(data, line)
	}

	// 遍历 data 切片并将数据存入 Elasticsearch
	bulkRequest := client.Bulk()
	for _, content := range data {
		// 创建文档请求
		doc := elastic.NewBulkIndexRequest().
			Index(indexName).
			Type("_doc"). // 仅在 Elasticsearch 7.x 使用此行
			Doc(map[string]interface{}{
				"content": content,
			})

		// 将文档请求添加到批量请求中
		bulkRequest = bulkRequest.Add(doc)
	}

	// 执行批量写入请求
	_, err = bulkRequest.Do(context.Background())
	if err != nil {
		log.Fatalf("Error indexing documents: %v", err)
		return
	}

	// 设置 HTTP 路由和处理程序
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")

		// 使用 Elasticsearch 进行搜索
		searchResult, err := client.Search().
			Index(indexName).
			Query(elastic.NewMatchQuery("content", query)).
			Do(context.Background())

		if err != nil {
			log.Printf("Error searching Elasticsearch: %v", err)
			http.Error(w, "Error searching Elasticsearch", http.StatusInternalServerError)
			return
		}

		var results []string
		for _, hit := range searchResult.Hits.Hits {
			// 解析 _source 字段
			var source map[string]interface{}
			if err := json.Unmarshal(hit.Source, &source); err == nil {
				// 获取 content 字段的值
				content, ok := source["content"].(string)
				if ok {
					results = append(results, content)
				}
			}
		}

		// 返回 JSON 格式的搜索结果
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		responseJSON, _ := json.Marshal(results)
		w.Write(responseJSON)
	})

	// 启动 HTTP 服务器
	fmt.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
