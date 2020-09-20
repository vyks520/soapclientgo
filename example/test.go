package main

import (
	"encoding/xml"
	"fmt"
	"github.com/vyks520/soapclientgo"
)

func main() {
	//SOAP客户端配置
	soapConfig := soapclientgo.Config{
		Version:  "1.1",                                      //版本号，留空默认1.2
		//ProxyURL: "http://proxyName:proxyPwd@127.0.0.1:8888", //http代理，不使用留空
		//UserName: "",                                         //认证用户
		//Password: "",                                         //认证密码
	}

	//创建SOAP客户端
	soapClient, err := soapclientgo.NewClient(soapConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	//SOAP请求参数（其他格式的参数可自定义转换为XML字符串传入即可）
	reqParams := []soapclientgo.ParamItem{
		{Key: "byProvinceName", Value: "广东"},
		//带属性参数示例
		/*{
			Key:   "byProvinceName",
			Value: "广东",
			Attr: map[string]interface{}{
				"attr1": "value1",
				"attr2": "value2",
			},
		},*/
	}

	//构造SOAP XML请求主体
	reqBody, err := soapClient.GenSOAPXml("http://WebXml.com.cn/", "getSupportCity", reqParams)
	if err != nil {
		panic(err)
	}
	fmt.Println(reqBody)

	//发起SOAP请求
	resBytes, err := soapClient.Request(
		"http://www.webxml.com.cn/WebServices/WeatherWebService.asmx",
		reqBody, "http://WebXml.com.cn/getSupportCity")

	if nil != err {
		fmt.Println(string(resBytes)) //错误时也可能返回数据，可用于调试
		panic(err)
	}

	cityData := SupportCityResponse{}
	err = xml.Unmarshal(resBytes, &cityData)
	if err != nil {
		panic(err)
	}
	fmt.Println(cityData)
}

//响应数据Body元素结构
type SupportCityResponse struct {
	XMLName              xml.Name `xml:"getSupportCityResponse"`
	Text                 string   `xml:",chardata"`
	Xmlns                string   `xml:"xmlns,attr"`
	GetSupportCityResult struct {
		Text   string   `xml:",chardata"`
		String []string `xml:"string"`
	} `xml:"getSupportCityResult"`
}
