package convert

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mr-tron/base58/base58"
)

var regexDigit = regexp.MustCompile(`^\d+$`)

func IsDigit(str string) bool {
	if regexDigit.MatchString(str) {
		return true
	}
	return false
}

func JoinStrings(multiString ...string) string {
	return strings.Join(multiString, "")
}

func JoinSepStrings(sep string, multiString ...string) string {
	return strings.Join(multiString, sep)
}

func JoinIntSlice2String(intSlice []int, sep string) string {
	return strings.Join(IntSlice2StrSlice(intSlice), sep)
}

func StrSlice2IntSlice(strSlice []string) []int {
	var intSlice []int
	for _, s := range strSlice {
		i, _ := strconv.Atoi(s)
		intSlice = append(intSlice, i)
	}
	return intSlice
}

func StrSplit2IntSlice(str, sep string) []int {
	return StrSlice2IntSlice(StrFilterSliceEmpty(strings.Split(str, sep)))
}

func IntSlice2StrSlice(intSlice []int) []string {
	var strSlice []string
	for _, i := range intSlice {
		s := strconv.Itoa(i)
		strSlice = append(strSlice, s)
	}
	return strSlice
}

func StrFilterSliceEmpty(strSlice []string) []string {
	var filterSlice []string
	for _, s := range strSlice {
		ss := strings.TrimSpace(s)
		if ss != "" {
			filterSlice = append(filterSlice, ss)
		}
	}
	return filterSlice
}

func Str2Int(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func Str2Int64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func Int2Str(i int) string {
	return strconv.Itoa(i)
}

var randomBytes = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomCustomStr(strs []byte, length int) string {
	maxL := len(strs)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, strs[r.Intn(maxL)])
	}
	return string(result)
}

func RandomStr(length int) string {
	maxL := len(randomBytes)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, randomBytes[r.Intn(maxL)])
	}
	return string(result)
}

var randomBytesInt = []byte("0123456789")

func RandomInt(length int) string {
	maxL := len(randomBytesInt)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, randomBytesInt[r.Intn(maxL)])
	}
	return string(result)
}

func StrMd5(s string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(s))
	return hex.EncodeToString(md5Ctx.Sum(nil))
}

func StringSliceRsort(s []string) []string {
	sort.Strings(s)
	for from, to := 0, len(s)-1; from < to; from, to = from+1, to-1 {
		s[from], s[to] = s[to], s[from]
	}
	return s
}

func JsonEncode(obj interface{}) (string, error) {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func JsonDecode(jsonString string, obj interface{}) error {
	return json.Unmarshal([]byte(jsonString), obj)
}

func Base64Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func Base64Decode(s string) ([]byte, error) {
	ds, err := base64.StdEncoding.DecodeString(s)
	return ds, err
}

func Base64UrlEncode(b []byte) string {
	return base64.URLEncoding.EncodeToString(b)
}

func Base64UrlDecode(s string) ([]byte, error) {
	ds, err := base64.URLEncoding.DecodeString(s)
	return ds, err
}

// Base58加密
func Base58Encode(str string) string {
	return base58.Encode([]byte(str))
}

func Base58Decode(str string) (string, error) {
	if bytes, e := base58.Decode(str); e == nil {
		return string(bytes), nil
	} else {
		return "", e
	}
}

// 生成guid 作为sessionID
func GenerateGuid() string {
	return StrMd5(uuid.New().String())
}

func GenerateUUID() string {
	return uuid.New().String()
}

//MD5 加密
func GetMD5(str string) (md5str string) {
	data := []byte(str)
	has := md5.Sum(data)
	md5str = fmt.Sprintf("%x", has)
	//md5str = string(has)
	return
}

//MD5 加密
func MD5(data []byte) []byte {
	hash := md5.New()
	md := hash.Sum(nil)
	return md
}

//SHA1 加密
func SHA1(data []byte) []byte {
	hash := sha1.New()
	hash.Write(data)
	md := hash.Sum(nil)
	return md
}

func GetSHA1(data string) string {
	return fmt.Sprintf("%x", SHA1([]byte(data)))
}

//SHA256 加密
func SHA256(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

func GetSHA256(data string) string {
	return fmt.Sprintf("%x", SHA256([]byte(data)))
}

//HmacSHA1 加密
func HmacSHA1(secret string, data []byte) []byte {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write(data)
	md := h.Sum(nil)
	return md
}

func GetHmacSHA1(secret string, data string) string {
	return fmt.Sprintf("%x", HmacSHA1(secret, []byte(data)))
}

func HmacSHA256(secret string, data []byte) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	md := h.Sum(nil)
	return md
}

func GetHmacSHA256(secret string, data string) string {
	return fmt.Sprintf("%x", HmacSHA256(secret, []byte(data)))
}

//HmacMD5 加密
func HmacMD5(secret string, data []byte) []byte {
	h := hmac.New(md5.New, []byte(secret))
	h.Write(data)
	md := h.Sum(nil)
	return md
}

func GetHmacMD5(secret string, data string) string {
	return fmt.Sprintf("%x", HmacMD5(secret, []byte(data)))
}
