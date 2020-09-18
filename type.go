package soapclientgo

import (
	"encoding/xml"
	"net/http"
	"net/url"
)

type ParamItem struct {
	Key   string
	Value interface{}
	Attr  map[string]interface{}
}

type ClientConfig struct {
	Version  string //版本 1.1 or 1.2
	ProxyURL string //代理
	UserName string //授权验证用户名称
	Password string //授权验证用户密码
}

type client struct {
	client        *http.Client //http请求客户端
	version       string       //版本 1.1 or 1.2 默认为1.2
	proxyURL      *url.URL     //代理地址
	userName      string       //授权验证用户名称
	password      string       //授权验证用户密码
	authorization string       //认证字符串
}

type resEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    *resBody `xml:"Body"`
}

type resBody struct {
	ResponseData []byte `xml:",innerxml"`
}
