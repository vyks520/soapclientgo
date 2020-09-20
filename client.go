package soapclientgo

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func NewClient(config Config) (*client, error) {
	soapClient := new(client)
	//版本号
	switch config.Version {
	case "1.1":
		soapClient.version = "1.1"
	case "1.2":
		soapClient.version = "1.2"
	default: //未赋值或版本号不正确设置为1.2版本
		soapClient.version = "1.2"
	}
	//代理URL处理
	if config.ProxyURL != "" {
		proxy, err := url.Parse(config.ProxyURL)
		if err != nil {
			errMsg := fmt.Sprintf("NewClient error: %s", err.Error())
			return nil, errors.New(errMsg)
		}
		soapClient.proxyURL = proxy
	}
	//授权认证
	if config.UserName != "" {
		soapClient.userName = config.UserName
		soapClient.password = config.Password
	}
	//请求客户端初始化
	if soapClient.proxyURL != nil {
		soapClient.client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(soapClient.proxyURL),
			},
		}
	} else {
		soapClient.client = &http.Client{}
	}
	return soapClient, nil
}

func (c client) Request(reqURL string, reqBody string, soapAction string) ([]byte, error) {
	//创建请求
	req, err := http.NewRequest("POST", reqURL, strings.NewReader(reqBody))
	if nil != err {
		errMsg := fmt.Sprintf("NewRequest error: %s", err.Error())
		return nil, errors.New(errMsg)
	}
	if c.userName != "" {
		req.SetBasicAuth(c.userName, c.password)
	}

	//SOAP Header
	switch c.version {
	case "1.1":
		req.Header.Set("Content-Type", "text/xml; charset=utf-8")
		req.Header.Set("SOAPAction", soapAction)
	case "1.2":
		var action string
		if soapAction != "" {
			action = fmt.Sprintf(`;action="%s"`, soapAction)
		}
		req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8"+action)
	}

	//发起请求，响应结果
	res, err := c.client.Do(req)
	if nil != err {
		errMsg := fmt.Sprintf("WebService soap%s request http post fail: %s", c.version, err.Error())
		return nil, errors.New(errMsg)
	}

	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Println(err.Error())
		}
	}()

	//读取响应数据
	resBytes, err := ioutil.ReadAll(res.Body)
	var (
		readErr   error
		statusErr error
	)
	//响应状态
	if http.StatusOK != res.StatusCode {
		errMsg := fmt.Sprintf("WebService soap%s request fail,http status: %d", c.version, res.StatusCode)
		statusErr = errors.New(errMsg)
	}
	if err != nil {
		errMsg := fmt.Sprintf("WebService soap%s ioutil ReadAll err: %s", c.version, err.Error())
		readErr = errors.New(errMsg)
	}
	envelope := resEnvelope{}
	if resBytes != nil {
		err = xml.Unmarshal(resBytes, &envelope)
		if err != nil {
			errMsg := fmt.Sprintf("WebService soap%s resonseEnvelope xmlUnmarshal fail: %s", c.version, err.Error())
			return resBytes, errors.New(errMsg)
		}
	}

	if statusErr != nil {
		return envelope.Body.ResponseData, statusErr
	} else if readErr != nil {
		return envelope.Body.ResponseData, readErr
	}
	return envelope.Body.ResponseData, nil
}

//格式化SOAP请求Body
func (c client) GenSOAPXml(
	nameSpace string,      //命名空间
	methodName string,     //调用方法名称
	reqParams interface{}, //请求参数
) (string, error) {
	var (
		soapBodyXml  string
		reqParamsXml string
	)
	switch params := reqParams.(type) {
	case string:
		reqParamsXml = params
	case []ParamItem:
		reqParamsXml = c.GenSoapParamsXlm(params, "        ")
	default:
		return "", errors.New(`GenSOAPXml "reqParams" The data type is not supported`)
	}
	switch c.version {
	case "1.1":
		soapBodyXml = `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
    <soap:Body>
        <%s xmlns="%s">%s
        </%s>
    </soap:Body>
</soap:Envelope>`
	case "1.2":
		soapBodyXml = `<?xml version="1.0" encoding="utf-8"?>
<soap12:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
	<soap12:Body>
		<%s xmlns="%s">%s
		</%s>
	</soap12:Body>
</soap12:Envelope>`
	}
	return fmt.Sprintf(soapBodyXml, methodName, XMLEscape(nameSpace), reqParamsXml, methodName), nil
}

//参数转换XML字符串
func (c client) GenSoapParamsXlm(soapParams []ParamItem, prefix string) string {
	result := ""
	for _, item := range soapParams {
		attrs := ""
		for attrKey, attrValue := range item.Attr {
			attrs += fmt.Sprintf(` %s="%s"`, attrKey, XMLEscape(fmt.Sprint(attrValue)))
		}
		result += fmt.Sprintf("\n%s<%s%s>%s</%s>", prefix, item.Key, attrs, XMLEscape(fmt.Sprint(item.Value)), item.Key)
	}
	return result
}

//XML转义
func XMLEscape(value string) string {
	buf := bytes.NewBufferString("")
	xml.Escape(buf, []byte(value))
	return string(buf.Bytes())
}
