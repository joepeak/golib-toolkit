package oss

import (
	"errors"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	_ "github.com/joepeak/golib-conf"
	"github.com/joepeak/golib-util/convert"
)

const (
	OSSConfKeyAws = "oss.aws"
)

var (
	awsPrefix = "[oss-aws]"
	awsConf   AwsConf

	uploader *s3manager.Uploader
)

type AwsConf struct {
	AccessKeyId     string
	AccessKeySecret string
	EndPoint        string
	Region          string
	BucketName      string
	UploadPath      string
}

func init() {
	// 初始化aws
	if viper.IsSet(OSSConfKeyAws) {
		if err := viper.UnmarshalKey(OSSConfKeyAws, &awsConf); err != nil {
			logrus.Error(awsPrefix, "初始化亚马逊OSS错误, ", err)
		} else {
			logrus.Info(awsPrefix, "初始化亚马逊OSS成功")
		}
	} else {
		logrus.Info(awsPrefix, "未配置亚马逊OSS")
	}
}

// 初始化AWS客户端
func initAwsClient(conf AwsConf) {
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    aws.String(conf.EndPoint),
		Region:      aws.String(conf.Region),
		Credentials: credentials.NewStaticCredentials(conf.AccessKeyId, conf.AccessKeySecret, ""),
	}))

	// Create an uploader with S3 client and default options
	uploader = s3manager.NewUploaderWithClient(s3.New(sess))
}

// 上传文件流
func UploadToAws(fileName string, r io.Reader) (string, error) {
	prefix := "【上传文件至亚马逊】"

	if uploader == nil {
		return "", errors.New(prefix + "亚马逊会话对象为空，无法上传")
	}

	// 上传路径
	fiePath := awsConf.UploadPath + "/" + fileName

	// 上传文件
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(awsConf.BucketName),
		Key:    aws.String(fiePath),
		Body:   r,
		ACL:    aws.String("public-read"),
	})

	if err != nil {
		logrus.Error("上传文件失败, ", err)
		return "", err
	}

	logrus.Info("上传文件成功, ", convert.ObjToJson(result))

	fileUrl := result.Location
	// fileUrl := awsConf.Domain + "/" + fiePath
	logrus.Info("文件上传成功, 文件地址:", result.Location)

	return fileUrl, nil
}
