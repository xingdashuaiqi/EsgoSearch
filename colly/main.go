package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gocolly/colly/v2"
)

func main() {
	// 创建一个新的 Collector 对象并指定一些选项
	c := colly.NewCollector(
		colly.AllowedDomains("news.baidu.com"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Safari/537.36"),
	)

	// 创建文件来保存爬取的数据
	file, err := os.Create("news_data.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// 在发出请求之前执行的回调
	c.OnRequest(func(r *colly.Request) {
		// 设置 Request 头部信息
		r.Headers.Set("Host", "https://www.jxuspt.com/")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Accept", "*/*")
		r.Headers.Set("Origin", "http://www.baidu.com")
		r.Headers.Set("Referer", "http://www.baidu.com")
		r.Headers.Set("Accept-Encoding", "gzip,deflate")
		r.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9")
		fmt.Println("Visiting", r.URL)
	})

	// 对响应的 HTML 元素进行处理
	c.OnHTML("title", func(e *colly.HTMLElement) {
		// 处理 title 元素
		title := e.Text
		// 写入数据到文件
		_, err := file.WriteString("Title: " + title + "\n")
		if err != nil {
			fmt.Println("Error writing to file:", err)
		}

		fmt.Println("Title:", title)
	})

	// 查找所有具有 class="hotnews"属性的 <div>内的 <a> 元素
	c.OnHTML(".hotnews a", func(e *colly.HTMLElement) {
		// 获取 href 和文本内容
		href := e.Attr("href")
		title := e.Text
		fmt.Printf("新闻: %s - %s\n", title, href)
		// 写入数据到文件
		_, err := file.WriteString("新闻: " + title + " - " + href + "\n")
		if err != nil {
			fmt.Println("Error writing to file:", err)
		}

		// 访问链接
		c.Visit(href)
	})

	// 提取状态码
	c.OnResponse(func(r *colly.Response) {
		fmt.Println("response received", r.StatusCode)
		fmt.Println(r.Ctx.Get("uri"))
	})

	// 对 visit 的线程数做限制，可以同时运行多个
	c.Limit(&colly.LimitRule{
		Parallelism: 2,
		Delay:       5 * time.Second,
	})

	// 启动访问
	err = c.Visit("http://news.baidu.com")
	if err != nil {
		fmt.Println("Error visiting website:", err)
	}
}
