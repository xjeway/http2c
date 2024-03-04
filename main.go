package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/spf13/pflag"
	"golang.org/x/net/http2"
)

func main() {
	// 定义命令行参数
	url := pflag.StringP("url", "u", "", "URL to send the HTTP/2 request. eg: -u https://localhost/index")
	method := pflag.StringP("method", "X", "GET", "HTTP method (GET, POST, etc.). eg: -X POST")
	headers := pflag.StringArrayP("header", "H", []string{}, "HTTP headers in key:value format")
	//
	data := pflag.StringP("data", "d", "", "Data to include in the request body.JSON and YAML formats are accepted.")

	// 解析命令行参数
	pflag.Parse()

	// 验证必需参数
	if *url == "" {
		fmt.Println("Error: URL is required")
		pflag.PrintDefaults()
		os.Exit(1)
	}

	// 创建一个HTTP客户端
	client := &http.Client{
		Transport: transport(url),
	}

	// 创建请求
	requestBody, err := readData(data)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(strings.ToUpper(*method), *url, requestBody)
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}

	// 设置请求头
	for _, header := range *headers {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		os.Exit(1)
	}

	// 打印响应
	fmt.Println("Status Code:", resp.Status)
	fmt.Println("Response Body:", string(responseBody))
}

func readData(data *string) (*bytes.Buffer, error) {
	var b []byte
	if *data == "" {
		return nil, nil
	}
	if strings.HasPrefix(*data, "@") {
		bdata, err := os.ReadFile(strings.Replace(*data, "@", "", 1))
		if err != nil {
			return nil, err
		}
		b = bdata
	} else {
		b = []byte(*data)
	}
	codec := encoding.GetCodec(format(b))
	if codec == nil {
		return nil, fmt.Errorf("format not found")
	}
	tmpMap := map[string]interface{}{}
	err := codec.Unmarshal(b, &tmpMap)
	if err != nil {
		return nil, err
	}
	dat, err := json.Marshal(tmpMap)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("Data: %s - %s", b, dat)
	return bytes.NewBuffer(dat), nil
}
func transport(URL *string) *http2.Transport {
	u, err := url.Parse(*URL)
	if err != nil {
		log.Fatal("URL parse err", err)
	}

	switch u.Scheme {
	case "http":
		return &http2.Transport{
			DisableCompression: true,
			AllowHTTP:          true,
			DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		}
	default:
		return &http2.Transport{}
	}
}
