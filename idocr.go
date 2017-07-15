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
	Name        string // 姓名
	Sex         string // 性别: 男|女
	Nationality string // 民族: 汉
	Birth       string // 出生: 19890714
	Address     string // 住址
	Num         string // 身份证号 带x和纯数字
	Issue       string // 签发机关
	StartDate   string // 有效期限，开始时间 20170714
	EndDate     string // 有效期限，结束时间 20370714
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
	// 正脸面
	Address     string `json:"address"`
	Birth       string `json:"birth"`
	Name        string `json:"name"`
	Nationality string `json:"nationality"`
	Num         string `json:"num"`
	RequestId   string `json:"request_id"`
	Sex         string `json:"sex"`
	Success     bool   `json:"success"` // 识别结果，true表示成功，false表示失败
	// 国徽面
	Issue     string `json:"issue"` // 签发机关
	EndDate   string `json:"end_date"`
	StartDate string `json:"start_date"`
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

	var (
		imgFace       string
		imgBack       string
		base64imgFace string
		base64imgBack string
		idcardBack    IDOCR
		err1          error
		err2          error
		err3          error
		err4          error
	)

	imgFace, err1 = get(faceUrl)
	imgBack, err2 = get(backUrl)

	if err1 != nil || err2 != nil {
		glog.Errorf("@%s, get img failed, faceUrl=%s, backUrl=%s, face_err=%s, back_err=%s", fn, faceUrl, backUrl, err1, err2)
		return
	}

	base64imgFace = base64.StdEncoding.EncodeToString([]byte(imgFace))
	base64imgBack = base64.StdEncoding.EncodeToString([]byte(imgBack))

	err3 = idcard.Face(base64imgFace)
	err4 = idcardBack.Back(base64imgBack)

	if err3 != nil || err4 != nil {
		glog.Errorf("@%s, img OCR failed, face_err=%s, back_err=%s", fn, err3, err4)
		return
	}

	idcard.Issue = idcardBack.Issue
	idcard.StartDate = idcardBack.StartDate
	idcard.EndDate = idcardBack.EndDate

	return
}

func get(url string) (result string, err error) {
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

	result = string(body)
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
