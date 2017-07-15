# aliyun-id-ocr

## 阿里云身份证 OCR 介绍

- 链接地址: [印刷文字识别-身份证识别](https://market.aliyun.com/products/57124001/cmapi010401.html?spm=5176.doc51066.765261.490.zn1VeX#sku=yuncode440100000)
- 调用方式: 上传图片的 base64 编码值
- 接口速度: 主要取决于图片大小，一般2M以内响应比较及时，超过2M可能会超时，返回 HTTP 状态码 408

## 使用方法

添加自己的产品密钥

```go
vi idocr.go
// TODO 修改此处，使用配置文件
const APPCODE = "hehe"
```

类结构

```go
▼ package
    idocr

▼ imports
    crypto/tls
    encoding/base64
    encoding/json
    errors
    fmt
    github.com/golang/glog
    io/ioutil
    net/http
    strings

▼ constants
   +APPCODE

▼+IDOCR : struct
    [fields]
   +Address : string
   +Birth : string
   +EndDate : string
   +Issue : string
   +Name : string
   +Nationality : string
   +Num : string
   +Sex : string
   +StartDate : string
    [methods]
   +Back(base64img string) : error
   +Face(base64img string) : error
   -formatResult(input string) : aliDataValue, error
   -post(img, face string) : string, error
    [functions]
   +GetIDCard(faceUrl, backUrl string) : IDOCR, error

▼-aliDataValue : struct
    [fields]
   +Address : string
   +Birth : string
   +EndDate : string
   +Issue : string
   +Name : string
   +Nationality : string
   +Num : string
   +RequestId : string
   +Sex : string
   +StartDate : string
   +Success : bool

▼-aliDetail : struct
    [fields]
   +OutputLabel : string
   +OutputValue : aliOutputValue

▼-aliOutputValue : struct
    [fields]
   +DataValue : string

▼-aliResult : struct
    [fields]
   +Outputs : []aliDetail

▼ functions
   -get(url string) : string, error

```

## 测试效果

```go
 go test -v                                                                                                                                                                                146 ↵
 === RUN   TestFace
 --- PASS: TestFace (2.42s)
 === RUN   TestBack
 --- PASS: TestBack (1.02s)
 === RUN   TestGet
 --- PASS: TestGet (2.52s)
 === RUN   TestGetIDCard
 --- PASS: TestGetIDCard (5.87s)
 === RUN   TestFormatResult
 --- PASS: TestFormatResult (0.00s)
 PASS
 ok  	_/Users/philiptang/Code/aliyun-id-ocr	11.855s
 ```

详细打印

```go
go test -v -stderrthreshold=INFO
```


