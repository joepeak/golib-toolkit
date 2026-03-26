package oss

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	_ "github.com/joepeak/golib-conf"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tencentyun/cos-go-sdk-v5"
)

var (
	txyunPrefix = "[oss-txyun]"

	txyunConf TxyunConf
	cosClient *cos.Client
)

const (
	OSSConfKeyTxyun = "oss.txyun"
)

type TxyunConf struct {
	BucketUrl  string
	UploadPath string
	SecretId   string
	SecretKey  string
	Timeout    time.Duration
}

func init() {
	// 初始化txyun
	if viper.IsSet(OSSConfKeyTxyun) {
		if err := viper.UnmarshalKey(OSSConfKeyTxyun, &txyunConf); err != nil {
			logrus.Error(txyunPrefix, "初始化腾讯云OSS错误, ", err)
		} else {
			if err := initTxyunClient(txyunConf); err != nil {
				logrus.Error(txyunPrefix, "初始化腾讯云OSS错误, ", err)
			} else {
				logrus.Info(txyunPrefix, "初始化腾讯云OSS成功")
			}
		}
	} else {
		logrus.Info(txyunPrefix, "未配置腾讯云OSS")
	}
}

// 初始化腾讯云客户端
func initTxyunClient(conf TxyunConf) error {
	u, err := url.Parse(conf.BucketUrl)
	if err != nil {
		logrus.Error(txyunPrefix, "初始化腾讯云BucketUrl错误, ", err)
		return fmt.Errorf("初始化腾讯云BucketUrl错误")
	}

	cosClient = cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Timeout: conf.Timeout * time.Second,
		Transport: &cos.AuthorizationTransport{
			SecretID:  conf.SecretId,
			SecretKey: conf.SecretKey,
		},
	})

	logrus.Info(txyunPrefix, "初始化腾讯云OSS成功, 超时时间：", conf.Timeout*time.Second)

	return nil
}

// 上传文件流
func UploadToTxyun(fileName string, r io.Reader) (string, error) {
	uploadName := fmt.Sprintf("%s/%s", txyunConf.UploadPath, fileName)

	_, err := cosClient.Object.Put(context.Background(), uploadName, r, nil)
	if err != nil {
		logrus.Error(txyunPrefix, "上传文件流异常：", err)
		return "", fmt.Errorf("上传文件流失败")
	}

	fileUrl := fmt.Sprintf("%s/%s", txyunConf.BucketUrl, uploadName)
	logrus.Info(txyunPrefix, "Put success, fileUrl=", fileUrl)

	return fileUrl, nil
}

// 上传到腾讯云OSS指定路径
func UploadToTxyunWithPath(path, fileName string, r io.Reader) (string, error) {
	uploadName := fmt.Sprintf("%s/%s", path, fileName)

	_, err := cosClient.Object.Put(context.Background(), uploadName, r, nil)
	if err != nil {
		logrus.Error(txyunPrefix, "上传文件流异常：", err)
		return "", fmt.Errorf("上传文件流失败")
	}

	fileUrl := fmt.Sprintf("%s/%s", txyunConf.BucketUrl, uploadName)
	logrus.Info(txyunPrefix, "Put success, fileUrl=", fileUrl)

	return fileUrl, nil
}

// 上传文件
func UploadFileToTxyun(filePath string) (string, error) {
	uploadName := fmt.Sprintf("%s/%s", txyunConf.UploadPath, filepath.Base(filePath))

	_, err := cosClient.Object.PutFromFile(context.Background(), uploadName, filePath, nil)
	if err != nil {
		logrus.Error(txyunPrefix, "上传文件异常：", err)
		return "", fmt.Errorf("上传文件失败")
	}

	fileUrl := fmt.Sprintf("%s/%s", txyunConf.BucketUrl, uploadName)
	logrus.Info(txyunPrefix, "PutFromFile success, fileUrl=", fileUrl)

	return fileUrl, nil
}
