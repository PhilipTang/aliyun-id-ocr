// 专门身份证的 OCR 识别
package idocr

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang/glog"
)

// TODO 修改此处，使用配置文件
const APPCODE = "hehe"

type IDOCR struct {
	Name        string `json:"name"`        // 姓名
	Sex         string `json:"sex"`         // 性别: 男|女
	Nationality string `json:"nationality"` // 民族: 汉
	Birth       string `json:"birth"`       // 出生: 19890714
	Address     string `json:"address"`     // 住址
	Num         string `json:"num"`         // 身份证号 带x和纯数字
	Issue       string `json:"issue"`       // 签发机关
	StartDate   string `json:"start_date"`  // 有效期限，开始时间 20170714
	EndDate     string `json:"end_date"`    // 有效期限，结束时间 20370714
}

type aliResult struct {
	Outputs []aliDetail `json:"outputs"`
}

type aliDetail struct {
	OutputLabel string         `json:"outputLabel"`
	OutputValue aliOutputValue `json:"outputValue"`
}

type aliOutputValue struct {
	DataValue string `json:"dataValue"`
}

type aliDataValue struct {
	RequestId string `json:"request_id"`
	Success   bool   `json:"success"` // 识别结果，true 表示成功，false 表示失败
	IDOCR
}

// 识别正脸面
func (id *IDOCR) Face(base64img string) (err error) {
	fn := "*IDOCR.Face"

	var result string
	if result, err = id.post(base64img, "face"); err != nil {
		glog.Errorf("@%s, err=%s", fn, err)
		return
	}
	var value aliDataValue
	if value, err = id.formatResult(result); err != nil {
		glog.Errorf("@%s, err=%s", fn, err)
		return
	}

	id.Address = value.Address
	id.Birth = value.Birth
	id.Name = value.Name
	id.Nationality = value.Nationality
	id.Num = value.Num
	id.Sex = value.Sex
	glog.Infof("@%s, id=%+v", fn, id)
	return
}

// 识别国徽面
func (id *IDOCR) Back(base64img string) (err error) {
	fn := "*IDOCR.Back"

	var result string
	if result, err = id.post(base64img, "back"); err != nil {
		glog.Errorf("@%s, err=%s", fn, err)
		return
	}
	var value aliDataValue
	if value, err = id.formatResult(result); err != nil {
		glog.Errorf("@%s, err=%s", fn, err)
		return
	}

	id.EndDate = value.EndDate
	id.StartDate = value.StartDate
	id.Issue = value.Issue
	glog.Infof("@%s, id=%+v", fn, id)
	return
}

func GetIDCard(faceUrl, backUrl string) (idcard IDOCR, err error) {
	fn := "idocr.GetIDCard"

	imgFace, imgBack, imgErr := getIDCardImg(faceUrl, backUrl)
	if imgErr != nil {
		err = imgErr
		glog.Errorf("@%s, getIDCardImg failed, err=%s", fn, err)
		return
	}

	var ocrErr error
	idcard, ocrErr = getIDCardOCR(imgFace, imgBack)
	if ocrErr != nil {
		err = ocrErr
		glog.Errorf("@%s, getIDCardOCR failed, err=%s", fn, err)
		return
	}

	return
}

func getIDCardImg(faceUrl, backUrl string) (imgFace, imgBack string, err error) {
	fn := "getIDCardImg"

	c1 := make(chan string)
	c2 := make(chan string)

	go func(url string) {
		imgData, _ := getAndBase64(url)
		c1 <- imgData
	}(faceUrl)
	go func(url string) {
		imgData, _ := getAndBase64(url)
		c2 <- imgData
	}(backUrl)

	for i := 0; i < 2; i++ {
		select {
		case imgFace = <-c1:
		case imgBack = <-c2:
		}
	}

	if imgFace == "" {
		err = errors.New(fmt.Sprintf("@%s, 请求正脸面照片失败, url=%s", fn, faceUrl))
		return
	}
	if imgBack == "" {
		err = errors.New(fmt.Sprintf("@%s, 请求国徽面照片失败, url=%s", fn, backUrl))
		return
	}
	return
}

func getIDCardOCR(imgFace, imgBack string) (idcard IDOCR, err error) {
	fn := "getIDCardOCR"

	c1 := make(chan IDOCR)
	c2 := make(chan IDOCR)

	go func(img string) {
		var id IDOCR
		_ = id.Face(img)
		c1 <- id
	}(imgFace)
	go func(img string) {
		var id IDOCR
		_ = id.Back(img)
		c2 <- id
	}(imgBack)

	var idcardBack IDOCR

	for i := 0; i < 2; i++ {
		select {
		case idcard = <-c1:
		case idcardBack = <-c2:
		}
	}

	if idcard.Name == "" {
		err = errors.New(fmt.Sprintf("@%s, 识别正脸面OCR失败, face_idcard=%+v", fn, idcard))
		return
	}
	if idcardBack.EndDate == "" {
		err = errors.New(fmt.Sprintf("@%s, 识别国徽面OCR失败, back_idcard=%+v", fn, idcardBack))
	}

	idcard.StartDate = idcardBack.StartDate
	idcard.EndDate = idcardBack.EndDate
	idcard.Issue = idcardBack.Issue

	return
}

func getAndBase64(url string) (result string, err error) {
	fn := "idocr.get"
	glog.Infof("@%s, url=%s", fn, url)

	req, _ := http.NewRequest("GET", url, nil)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("@%s, err=%s", fn, err)
		return
	}
	defer res.Body.Close()

	var body []byte
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("@%s, ioutil.ReadAll failed, err=%s", fn, err)
		return
	}

	if res.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("response status != 200, but = %d, res=%+v", res.StatusCode, res))
		glog.Errorf("@%s, err=%s", fn, err)
		return
	}

	result = base64.StdEncoding.EncodeToString(body)
	return
}

func (id *IDOCR) post(img, face string) (result string, err error) {
	fn := "*IDOCR.post"

	url := "https://dm-51.data.aliyun.com/rest/160601/ocr/ocr_idcard.json"
	payload := strings.NewReader("{\"inputs\":[{\"image\":{\"dataType\":50,\"dataValue\":\"" + img + "\"},\"configure\":{\"dataType\":50,\"dataValue\":\"{\\\"side\\\":\\\"" + face + "\\\"}\"}}]}")
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("authorization", "APPCODE "+APPCODE)
	req.Header.Add("content-type", "application/json")

	glog.Infof("@%s, request=%+v", fn, req)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("@%s, err=%s", fn, err)
		return
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("@%s, err=%s", fn, err)
		return
	}

	if res.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("response status != 200, but = %d, res=%+v", res.StatusCode, res))
		glog.Errorf("@%s, err=%s", fn, err)
		return
	}

	result = string(body)
	glog.Infof("@%s, response=%s", fn, result)
	return
}

func (id *IDOCR) formatResult(input string) (value aliDataValue, err error) {
	fn := "*IDOCR.formatResult"

	var info aliResult
	var jsonBlob = []byte(input)
	if err = json.Unmarshal(jsonBlob, &info); err != nil {
		glog.Errorf("@%s, json.Unmarshal failed, err=%s", fn, err)
		return
	}

	if len(info.Outputs) < 1 {
		err = errors.New(fmt.Sprintf("info.Outputs is empty, info=%+v", fn, info))
		glog.Errorf("@%s, err=%s", fn, err)
		return
	}

	dataValue := info.Outputs[0].OutputValue.DataValue
	glog.Infof("@%s, dataValue=%s", fn, dataValue)

	jsonBlob = []byte(dataValue)
	if err = json.Unmarshal(jsonBlob, &value); err != nil {
		glog.Errorf("@%s, json.Unmarshal failed, err=%s", fn, err)
		return
	}
	glog.Infof("@%s, value=%+v", fn, value)

	if value.Success != true {
		err = errors.New(fmt.Sprintf("身份证识别失败, value=%+v", fn, value))
		glog.Errorf("@%s, err=%s", fn, err)
		return
	}

	return
}
