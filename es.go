package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"github.com/rs/cors"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/olivere/elastic/v7"
)

// 初始化MySQL数据库连接
func initMySQL() (*sql.DB, error) {
	dbConfig := mysql.Config{
		User:                 "root",
		Passwd:               "123456",
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "jrbase",
		AllowNativePasswords: true,
	}
	return sql.Open("mysql", dbConfig.FormatDSN())
}

// 初始化Elasticsearch客户端连接
func initElasticsearch() (*elastic.Client, error) {
	return elastic.NewClient(elastic.SetURL("http://localhost:9200"))
}

// 将数据从MySQL存储到Elasticsearch
func indexDataToElasticsearch(db *sql.DB, client *elastic.Client, indexName string) {
	// 查询MySQL数据
	rows, err := db.Query(`
      SELECT title, content, url, date FROM news
      UNION
      SELECT title, content, url, date FROM notice
      UNION
      SELECT title, content, url, date FROM departmental
      UNION
      SELECT title, content, url, date FROM schoolenterprise
      UNION
      SELECT title, content, url, date FROM specialsubject
    `)
	if err != nil {
		log.Fatalf("Failed to query MySQL: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var title string
		var content string
		var url string
		var date string
		if err := rows.Scan(&title, &content, &url, &date); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		doc := map[string]interface{}{
			"title":   title,
			"content": content,
			"url":     url,
			"date":    date,
		}

		_, err := client.Index().
			Index(indexName).
			Id(url).
			BodyJson(doc).
			Do(context.Background())
		if err != nil {
			log.Printf("Failed to index document: %v", err)
		} else {
			log.Printf("Document indexed successfully: %s", url)
		}
	}
}

// 启动HTTP服务器
func startHTTPServer() {
	r := mux.NewRouter()
	r.HandleFunc("/search", handleSearchRequest)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://127.0.0.1:5173"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"*"},
		Debug:          true,
	})

	handler := c.Handler(r)

	http.Handle("/", handler)
	fmt.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}

// 处理搜索请求
func handleSearchRequest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	fmt.Println("Query:", query)

	client, err := initElasticsearch()
	if err != nil {
		log.Printf("Error creating Elasticsearch client: %v", err)
		http.Error(w, "Error creating Elasticsearch client", http.StatusInternalServerError)
		return
	}

	defer client.Stop()

	// Elasticsearch搜索
	searchResult, err := client.Search().
		Index("newsindex"). // 指定索引
		Query(elastic.NewMultiMatchQuery(query, "content")).
		Do(context.Background())

	if err != nil {
		log.Printf("Error searching Elasticsearch: %v", err)
		http.Error(w, "Error searching Elasticsearch", http.StatusInternalServerError)
		return
	}

	var results []map[string]interface{}
	for _, hit := range searchResult.Hits.Hits {
		// 解析 _source 字段
		var source map[string]interface{}
		if err := json.Unmarshal(hit.Source, &source); err == nil {
			results = append(results, source)
		}
	}

	// 返回 JSON 格式的搜索结果
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	responseJSON, _ := json.Marshal(results)
	w.Write(responseJSON)
}

func main() {
	db, err := initMySQL()
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer db.Close()

	client, err := initElasticsearch()
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %v", err)
		return
	}

	indexName := "newsindex"
	exists, err := client.IndexExists(indexName).Do(context.Background())
	if err != nil {
		log.Fatalf("Error checking index existence: %v", err)
		return
	}

	if !exists {
		log.Println("Index does not exist. You can create it if needed.")
	}

	indexDataToElasticsearch(db, client, indexName)
	startHTTPServer()
}
