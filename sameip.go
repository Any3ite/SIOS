package main

import (
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/liushuochen/gotable"
	"github.com/thanhpk/randstr"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func main() {
	fmt.Println(startmsg())
	//写一个函数用来转换IP地址然后交给censys的API去取结果
	var domain string
	flag.StringVar(&domain, "d", "", os.Args[0]+" -d www.example.com")
	flag.Parse()
	if domain == "" {
		flag.Usage()
		return
	} else {
		domain2ip(domain)
	}

}
func domain2ip(d string) {
	/*
		要判断一下错误的类型，然后根据相应的类型进行提示然后做处理;
		但是遇到一个ip地址是多个的情况下需要一个切片，所以还得做一个slice切片
	*/
	ip, err := net.LookupIP(d)
	errs(err)
	for i := 0; i < len(ip); i++ {
		// 将数据类型从IP转为string字符串并添加到slice中
		// 需要将IPv6地址排除在外，并且slice的结果是重复的 会有些恶心
		if strings.Count(ip[i].String(), ":") < 2 {
			requester(ip[i].String())
		} else {
			break
		}
	}

}

func requester(ipaddress string) {

	apikey := "08eba93f-xxxx-xxxx-xxxx-5eeacef47736:8KwwdxxxxxxxxxxxxxxxxxxxxxxxxXuMQbpzTymv"
	//先要弄一个http请求发起工具，配置好http请求超时、TLS证书错误忽略
	cli := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: time.Second * 15,
	}
	// 讲道理 硬编码把api写到了代码里面肯定不合适，但是懒得写读取文件什么的那些见了鬼的代码；
	base64APIstring := base64.StdEncoding.EncodeToString([]byte(apikey))
	/*需要一个循环来判断参数ipaddress的容量然后发送请求*/

	request, err := http.NewRequest(http.MethodGet, "https://search.censys.io/api/v2/hosts/"+ipaddress+"/names?per_page=999", nil)
	request.Header.Add("accept", "application/json")
	request.Header.Add("Authorization", "Basic "+base64APIstring)
	do, err := cli.Do(request)
	errs(err)
	defer func() {
		err := do.Body.Close()
		errs(err)
	}()
	all, err := ioutil.ReadAll(do.Body)
	errs(err)
	resultParser(all)

}

func resultParser(b []byte) {
	s := randstr.String(8)
	table, err := gotable.Create("id", "domain")
	errs(err)
	rows := make([]map[string]string, 0)
	json, err := simplejson.NewJson(b)
	errs(err)
	array, err := json.Get("result").Get("names").StringArray()
	for k, v := range array {
		row := make(map[string]string)
		row["id"] = strconv.Itoa(k)
		row["domain"] = v
		rows = append(rows, row)
	}
	_ = table.AddRows(rows)
	fmt.Println(table)
	defer func() {
		file, err := os.OpenFile("result"+s+".txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		errs(err)
		_, err = file.Write([]byte(table.String()))
		errs(err)
		_ = file.Close()
	}()
}

func errs(err error) {
	defer func() {
		if err := recover(); err != nil {
			switch err.(type) {
			case os.File:
				fmt.Println("[!x!] 哦！奇怪咯！真的是！哦！竟然无法保存结果！哦！。")
				return
			case net.Error:
				fmt.Println("[!x!] 当前域名无法被转换为有效的IPv4地址。")
				return
			case runtime.Error:
				fmt.Println("[!x!] 运行时错误！哦！真是见了鬼了！哦！为什么会报错呢？哦！。")
				return
			default:
				fmt.Println(err)
			}
		}
	}()
	if err != nil {
		panic(err)
	}
}

func startmsg() string {
	str := `  #####                          ### ######     #######            #####                                     
#     #   ##   #    # ######     #  #     #    #     # #    #    #     # ###### #####  #    # ###### #####  
#        #  #  ##  ## #          #  #     #    #     # ##   #    #       #      #    # #    # #      #    # 
 #####  #    # # ## # #####      #  ######     #     # # #  #     #####  #####  #    # #    # #####  #    # 
      # ###### #    # #          #  #          #     # #  # #          # #      #####  #    # #      #####  
#     # #    # #    # #          #  #          #     # #   ##    #     # #      #   #   #  #  #      #   #  
 #####  #    # #    # ######    ### #          ####### #    #     #####  ###### #    #   ##   ###### #    # 
                                                                                                            `
	return str
}
