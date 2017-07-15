# aliyun-id-ocr

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


