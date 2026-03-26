package oss

import (
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
)

var (
	OsType     = runtime.GOOS
	RootPath   string
	DataPath   string
	UploadPath string
)

func init() {

	RootPath = GetRootPath()
	DataPath = GetDataPath()
	UploadPath = GetUploadPath()
}
func GetRootPath() string {
	dir, err := os.Getwd()
	if err != nil {
		logrus.Error("获取根路径错误", err)
	}
	return dir
}

func GetDataPath() string {
	path := GetRootPath()

	if OsType == "windows" {
		path = path + "\\" + "data\\"
	} else {
		path = path + "/" + "data/"
	}
	return path
}

func GetUploadPath() string {
	path := GetRootPath()
	if OsType == "windows" {
		path = path + "\\" + "upload\\"
	} else {
		path = path + "/" + "upload/"
	}
	return path
}
