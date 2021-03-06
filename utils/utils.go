package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	INT_MAX = int(^uint(0) >> 1)
	INT_MIN = int(^INT_MAX)
	INT64_MAX = int64(INT_MAX)
	INT64_MIN = int64(INT_MIN)
)


type BindError struct {
	reason string
}

func (this BindError) Error() string {
	return this.reason
}

func Bind(c *gin.Context, obj interface{}) error {
	contentType := c.GetHeader("Content-Type")
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	fmt.Println(contentType)
	if contentType == binding.MIMEJSON {
		if err := c.BindJSON(obj); err != nil {
			//panic(err)
			return BindError{
				reason:	fmt.Sprintf("error occurred in json bind due to:\n%s", err.Error()),
			}
		}
	} else if contentType == binding.MIMEYAML {
		yamlBody, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			//panic(err)
			return BindError{
				reason:	fmt.Sprintf("error occurred in request body get due to:\n%s", err.Error()),
			}
		}
		jsonBody, err := yaml.ToJSON(yamlBody)
		if err != nil {
			//panic(err)
			return BindError{
				reason:	fmt.Sprintf("error occurred in yaml convert due to:\n%s", err.Error()),
			}
		}
		if err := json.Unmarshal(jsonBody, obj); err != nil {
			//panic(err)
			return BindError{
				reason:	fmt.Sprintf("error occurred in yaml-json unmarshal due to:\n%s", err.Error()),
			}
		}

	} else {
		fmt.Println("no content type match")
		return BindError{
			reason: "invalid content type",
		}
	}
	return nil
}

func DecodeK8SResources(c *gin.Context) ([]runtime.Object, error) {
	yamlFiles, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		//panic(err)
		return nil, BindError{
			reason:	fmt.Sprintf("error occurred in request body get due to:\n%s", err.Error()),
		}
	}
	return ParseK8SYaml(yamlFiles), nil
}

func ParseK8SYaml(fileR []byte) []runtime.Object {
	filesAsString := string(fileR[:])
	sepYamlFiles := strings.Split(filesAsString, "---")
	retObj := []runtime.Object{}
	for _, file := range sepYamlFiles {
		if file == "\n" || file == "" {
			continue
		}
		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode([]byte(file), nil, nil)
		if err != nil {
			fmt.Printf("error occurred when decoding yaml file\n %s ", err.Error())
			panic(err)
		}
		retObj = append(retObj, obj)
	}
	return retObj
}

func TrimSpace(strs []string) []string {
	ret := []string{}
	for _, str := range strs {
		ret = append(ret, strings.TrimSpace(str))
	}
	return ret
}

func Float64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func NewInt32(n int) *int32 {
	num := int32(n)
	return &num
}

func Int64Min(a int64, b int64) int64 {
	if a > b {
		return b
	}
	return a
}

func Int64Max(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

type Int64 int64

func (num Int64) KBtoB() int64 {
	return int64(num*1024)
}

func (num Int64) MBtoKB() int64 {
	return int64(num*1024)
}

func (num Int64) MBtoB() int64 {
	return int64(num*1024*1024)
}

func (num Int64) BtoMB() int64 {
	return int64(num/1024/1024)
}

func MustParseFloat64(str string) float64 {
	num, err := strconv.ParseFloat(str, 64)
	if err != nil {
		panic(err)
	}
	return num
}

func GetSelectedLineInFile(path string, num int) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for line := 0; line < num-1; line++ {
		_, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", err
		}
	}
	return reader.ReadString('\n')
}

func Split(str string, sep string) []string {
	arr := strings.Split(str, sep)
	return TrimSpace(arr)
}

func BadResponse(response *http.Response) bool {
	return response.StatusCode/100 > 3
}

func Stringify(obj interface{}) string {
	str, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		return fmt.Sprintf("%+v", obj)
	}
	return string(str)
}