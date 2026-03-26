package oss

import (
	"bytes"
	"encoding/base64"
	"io"

	_ "github.com/joepeak/golib-conf"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//说明文档：https://help.aliyun.com/document_detail/88601.html

// 上传配置
type AliyunConf struct {
	AccessKeyId     string
	AccessKeySecret string
	BucketName      string
	Domain          string // 域名
	EndPoint        string
	UploadPath      string // oss上传路径
	Region          string // 地区
}

var (
	aliyunPrefix = "[oss-aliyun]"
	aliyunConf   AliyunConf
	aliClient    *oss.Client
)

const (
	OSSConfKeyAliyun = "oss.aliyun"
)

func init() {
	// 初始化阿里云
	if viper.IsSet(OSSConfKeyAliyun) {
		if err := viper.UnmarshalKey(OSSConfKeyAliyun, &aliyunConf); err != nil {
			logrus.Error(aliyunPrefix, "初始化阿里云OSS错误, ", err)
		} else {
			logrus.Info(aliyunPrefix, "初始化阿里云OSS成功")
			initAliyunClient(aliyunConf)
		}
	} else {
		logrus.Info(aliyunPrefix, "未配置阿里云OSS")
	}
}

// 初始化阿里云OSS客户端
func initAliyunClient(conf AliyunConf) {
	if aliClient != nil {
		logrus.Info(aliyunPrefix, "阿里云OSS客户端已初始化")
		return
	}

	var err error
	aliClient, err = oss.New(conf.EndPoint, conf.AccessKeyId, conf.AccessKeySecret)
	if err != nil {
		logrus.Error(aliyunPrefix, "创建OSSClient实例错误, ", err)
	}
}

// 上传文件流
func ToAliyun(fileName string, r io.Reader) (string, error) {

	// 获取存储空间
	bucket, err := aliClient.Bucket(aliyunConf.BucketName)
	if err != nil {
		logrus.Error(aliyunPrefix, "获取存储空间错误, ", err)
		return "", err
	}

	// 上传文件
	uploadPath := aliyunConf.UploadPath + "/" + fileName

	err = bucket.PutObject(uploadPath, r)
	if err != nil {
		logrus.Error(aliyunPrefix, "上传文件流异常：", err)
		return "", err
	}

	fileUrl := aliyunConf.Domain + "/" + uploadPath
	logrus.Info(aliyunPrefix, "文件上传成功, 文件地址:", fileUrl)

	return fileUrl, nil
}

// 上传文件Base64
func Base64ToAliyun(fileName string, fileData string) (string, error) {
	byteData, err := base64.StdEncoding.DecodeString(fileData)
	if err != nil {
		return "", err
	}

	return ToAliyun(fileName, bytes.NewBuffer(byteData))
}
