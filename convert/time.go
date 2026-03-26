package convert

import (
	"fmt"
	"math"
	"time"

	"github.com/sirupsen/logrus"
)

// 时间戳转换为日期字符串
func TimestampToStr(sec int64) string {
	if sec == 0 {
		return ""
	}

	tm := time.Unix(sec, 0)
	return tm.Format("2006-01-02 15:04:05")
}

func NowToDateTimeStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// 时间戳转换为日期字符串
func TimestampToStrDate(sec int64) string {
	if sec == 0 {
		return ""
	}

	tm := time.Unix(sec, 0)
	return tm.Format("2006-01-02")
}

// 时间字符串转化为时间戳
func StrToTimestamp(str string) int64 {
	// 如果不包含时间，则加上时间段
	if len(str) == 10 {
		str = str + " 00:00:00"
	}

	//loc, _ := time.LoadLocation("Local") //获取时区
	t, err := time.ParseInLocation("2006-01-02 15:04:05", str, time.Local)
	if err != nil {
		logrus.Error("StrToTimestamp转换错误, str=", str, ", ", err)
		return 0
	}

	return t.Unix()
}

func DateToStartTimestamp(str string) int64 {
	// 如果不包含时间，则加上时间段
	if len(str) == 10 {
		str = str + " 00:00:00"
	}

	//loc, _ := time.LoadLocation("Local") //获取时区
	t, err := time.ParseInLocation("2006-01-02 15:04:05", str, time.Local)
	if err != nil {
		logrus.Error("StrToTimestamp转换错误, str=", str, ", ", err)
		return 0
	}

	return t.Unix()
}

func DateToEndTimestamp(str string) int64 {
	// 如果不包含时间，则加上时间段
	if len(str) == 10 {
		str = str + " 23:59:59"
	}

	//loc, _ := time.LoadLocation("Local") //获取时区
	t, err := time.ParseInLocation("2006-01-02 15:04:05", str, time.Local)
	if err != nil {
		logrus.Error("StrToTimestamp转换错误, str=", str, ", ", err)
		return 0
	}

	return t.Unix()
}

// 时间点转换为时间戳
func TimePointToTimestamp(timePoint string) int64 {

	layout := "2006-01-02 15:04" //转化所需模板
	date := time.Now().Format("2006-01-02")
	value := date + " " + timePoint                        //待转化为时间戳的字符串 注意 这里的小时和分钟还要秒必须写 因为是跟着模板走的
	loc, _ := time.LoadLocation("Local")                   //获取时区
	theTime, _ := time.ParseInLocation(layout, value, loc) //使用模板在对应时区转化为time.time类型

	return theTime.Unix()
}

func GetCurrentDateStr() string {
	return time.Now().Format("2006-01-02")
}

func GetCurrentDateNumber() string {
	return time.Now().Format("20060102")
}

func GetCurrentDateTimeNumber() string {
	return time.Now().Format("20060102150405")
}

// 获得今天0时0分0秒时间戳
func GetTodayStartTimestamp() int64 {
	t := time.Now()
	tm1 := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	return tm1.Unix()
}

// 获得今天23时59分59秒时间戳
func GetTodayEndTimestamp() int64 {
	t := time.Now()
	tm1 := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999, t.Location())

	return tm1.Unix()
}

// 获得今天0时0分0秒UTC时间戳
func GetTodayStartUtcTimestamp() int64 {
	t := time.Now()
	tm1 := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)

	return tm1.Unix()
}

// 获得昨天0时0分0秒UTC时间戳
func GetYesterdayStartUtcTimestamp() int64 {
	t := time.Now().AddDate(0, 0, -1)
	tm1 := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)

	return tm1.Unix()
}

// 获得今天23时59分59秒UTC时间戳
func GetTodayEndUtcTimestamp() int64 {
	t := time.Now()
	tm1 := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999, time.UTC)

	return tm1.Unix()
}

// 获得明天0时0分0秒时间戳
func GetTomorrowStartTimestamp() int64 {
	t := time.Now().AddDate(0, 0, 1)
	tm1 := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	return tm1.Unix()
}

// 获得昨天0时0分0秒时间戳
func GetYesterdayStartTimestamp() int64 {
	t := time.Now().AddDate(0, 0, -1)
	tm1 := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	return tm1.Unix()
}

// 获取2个时间戳之间的天数
func GetDaysBetweenTwoTimestamp(start int64, end int64) float64 {
	c := math.Abs(float64(end - start))

	days := c / (24 * 60 * 60)

	return days
}

func DateString(t time.Time) string {
	return t.Format("2006-01-02")
}

func DateNumberString(t time.Time) string {
	return t.Format("20060102")
}

func DateTimeString(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// 0点时间戳
func ZeroUnix() int64 {
	t1 := time.Now().Year()  //年
	t2 := time.Now().Month() //月
	t3 := time.Now().Day()   //日
	//t4:=time.Now().Hour()        //小时
	//t5:=time.Now().Minute()      //分钟
	//t6:=time.Now().Second()      //秒
	//t7:=time.Now().Nanosecond()  //纳秒
	currentTimeData := time.Date(t1, t2, t3, 0, 0, 1, 0, time.Local) //获取当前时间，返回当前时间Time
	fmt.Println("currentTimeData", currentTimeData.Unix())
	return currentTimeData.Unix()

}

// 获取当前时间
func GetCurrentTime() time.Time {
	return time.Now()
}

// 将毫秒时间戳转换为UTC时间格式
func MillisecondsToUTC(ms int64) string {
	if ms == 0 {
		return ""
	}

	// 取绝对值
	ms = int64(math.Abs(float64(ms)))

	// 检查是否为秒级别时间戳
	if ms < 1e12 {
		// 转换为毫秒级别
		ms *= 1000
	}

	// 转换为 time.Time 对象
	t := time.Unix(0, ms*int64(time.Millisecond))

	// 格式化为 UTC 时间字符串
	return t.UTC().Format("2006-01-02 15:04:05 UTC")
}
