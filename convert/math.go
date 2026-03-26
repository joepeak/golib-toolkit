package convert

import (
	"fmt"
	"math"
	"strings"
)

func IsOdd(num int) bool {
	return num&1 == 1
}

//value是否为base的整数倍
func IsIntegerMultiple(value int32, base int32) bool {
	if value < base {
		return false
	}

	return value%base == 0
}

// 保留小数位
//	val表示浮点数
//	n表示要保留的小数位
func KeepDecimal(val float64, n int) float64 {
	n10 := math.Pow10(n)
	f := math.Trunc((val+0.5/n10)*n10) / n10
	return f
}

// 保留小数位, 并转换为string
//	val表示浮点数
//	n表示要保留的小数位
func KeepDecimalToString(val float64, n int, zeroFill bool) string {
	f := KeepDecimal(val, n)
	s := fmt.Sprintf("%v", f)

	arr := strings.Split(s, ".")

	if len(arr) > 1 && zeroFill {

		l := len(arr[1])
		if l < n {
			l = n
		}

		format := fmt.Sprintf("%%.%df", l)
		s = fmt.Sprintf(format, f)
	}

	return s
}
