package convert

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// float64 - float64
func Float64DivFloat64(a float64, b float64) float64 {
	da := decimal.NewFromFloat(a)
	db := decimal.NewFromFloat(b)
	res, _ := da.Div(db).Float64()
	return res
}

func IntDivInt(a int, b int) int {
	da := decimal.New(int64(a), 0)
	db := decimal.New(int64(b), 0)
	return int(da.Sub(db).IntPart())
}

func Float64AddFloat64(a float64, b float64) float64 {
	da := decimal.NewFromFloat(a)
	db := decimal.NewFromFloat(b)
	aa, _ := da.Add(db).Float64()
	return aa
}

func Float64SubFloat64(a float64, b float64) string {
	da := decimal.NewFromFloat(a)
	db := decimal.NewFromFloat(b)
	res, _ := da.Sub(db).Float64()

	return fmt.Sprintf("%.2f", res)
}

func ObjToJson(obj interface{}) string {

	if obj == nil {
		return ""
	}

	b, err := json.Marshal(obj)
	if err != nil {
		logrus.Info("ObjToJson, error, ", err)
		return ""
	}

	return string(b)
}

func JsonToObj(jsonString string, obj interface{}) error {

	if len(jsonString) == 0 {
		return errors.New("JSON字符串为空")
	}

	err := json.Unmarshal([]byte(jsonString), &obj)
	if err != nil {
		logrus.Info("JsonToObj, error, ", err)
		return err
	}

	return nil
}

func ConvertStr2GBK(str string) string {
	data, err := simplifiedchinese.GBK.NewEncoder().String(str)
	if err != nil {
		logrus.Error("ConvertStr2GBK err, ", err)
		return ""
	}

	return data
}

func ConvertGBK2Str(gbkStr string) string {
	data, err := simplifiedchinese.GBK.NewDecoder().String(gbkStr)
	if err != nil {
		logrus.Error("ConvertGBK2Str err, ", err)
		return ""
	}

	return data
}

// string转int
func StringToInt(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}

	return n
}

// string转int32
func StringToInt32(s string) int32 {
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0
	}

	return int32(n)
}

// string转int64
func StringToInt64(s string) int64 {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}

	return n
}

// convert number to big.Int type
func StringToBigInt(s string) *big.Int {
	// convert number to big.Int type
	ip := new(big.Int)
	ip.SetString(s, 10) //base 10

	return ip
}

// string转float32
func StringToFloat32(s string) float32 {
	if f, err := strconv.ParseFloat(s, 32); err == nil {
		return float32(f)
	}

	return 0
}

// string转float64
func StringToFloat64(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	return 0
}

/*
*
Int64ToString
*/
func Int64ToString(value int64) string {
	str := strconv.FormatInt(value, 10)
	return str
}

func Float32ToString(n float32) string {
	return strconv.FormatFloat(float64(n), 'f', -1, 32)
}

func Float64ToString(n float64) string {
	return strconv.FormatFloat(n, 'f', -1, 64)
}

/*
浮点数转为百分数字符串
*/
func Float64ToPercentStr(x float64) string {

	s := fmt.Sprintf("%.2f%%", x*100)

	return s
}

func InterfaceToString(x interface{}) string {

	s := fmt.Sprintf("%v", x)

	return s
}

func InterfaceToFloat64(x interface{}) float64 {

	s := fmt.Sprintf("%v", x)

	f := StringToFloat64(s)

	return f
}

func IntToString(x int) string {

	s := strconv.FormatInt(int64(x), 10)

	return s
}

func Int32ToString(x int32) string {

	s := strconv.FormatInt(int64(x), 10)

	return s
}

func InterfaceToAny(x interface{}) string {

	if x == nil {
		return ""
	}

	s := ""

	if v, ok := x.(string); ok {
		s = v
	} else if v, ok := x.(int); ok {
		s = IntToString(v)
	} else if v, ok := x.(int32); ok {
		s = Int32ToString(v)
	} else if v, ok := x.(int64); ok {
		s = Int64ToString(v)
	} else if v, ok := x.(float64); ok {
		s = fmt.Sprintf("%.6f", v) // 对所有浮点型数据，指定6位小数
	} else if v, ok := x.(bool); ok {
		s = InterfaceToString(v)
	} else {
		s = InterfaceToString(x)
	}

	return s
}

func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
