package rand

import (
	"fmt"
	"math/rand"
	"time"
)

// tokenId随机获取
const (
	TokenIdMin = 1000000000
	TokenIdMax = 9999999999
)

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	numbers = []rune("0123456789")
)

// 获取6位长度的随机码
func RandNum() string {
	return fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
}

// 获取指定长度的随机数字字符串
func RandNumStr(length int32) string {
	str := "%0" + fmt.Sprintf("%d", length) + "v"
	return fmt.Sprintf(str, rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
}

// 获取给定范围的随机值
func RandInt(min int, max int) int {
	if min > max {
		return 0
	}

	// 将时间戳设置成种子数
	return rand.Intn(max-min) + min
}

// 随机tokenId
func RandTokenId() int64 {
	return int64(rand.Intn(TokenIdMax-TokenIdMin)) + TokenIdMin
}
