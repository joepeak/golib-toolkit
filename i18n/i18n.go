package i18n

import (
	"log"
	"os"
	"strings"
	"time"

	"regexp"

	"github.com/fsnotify/fsnotify"
	"github.com/naoina/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
)

const ()

var (
	// 支持的语言
	i18nMap map[string]*i18n.Localizer

	// 正则表达式
	re *regexp.Regexp

	baseDir = "i18n"
)

func init() {

	if !viper.GetBool("app.i18n.enabled") {
		return
	}

	initI18n()
}

// initI18n 初始化文件多语言
func initI18n() {
	if filePaths, err := LoadI18n(); err != nil {
		logrus.Error("load i18n file error, ", err)
		return
	} else {
		for _, filePath := range filePaths {
			go watchI18n(filePath)
		}
	}
}

// LoadI18n 从文件加载多语言数据
func LoadI18n() ([]string, error) {
	// 初始化对象
	i18nMap = make(map[string]*i18n.Localizer)

	// 初始化正则
	re = regexp.MustCompile(`message\.(\S+)\.toml`)

	filePaths := make([]string, 0)

	workFolder, err := os.Getwd()
	if err != nil {
		logrus.Error("get current work dir error, ", err)
		return filePaths, err
	} else {
		baseDir = workFolder + "/" + baseDir
	}

	logrus.Info("i18n folder: ", baseDir)

	// 获取指定目录下所有语言文件
	files, err := os.ReadDir(baseDir)
	if err != nil {
		logrus.Error("read i18n folder error, ", err)
		return filePaths, err
	}

	for _, f := range files {
		if f.IsDir() {
			logrus.Info("skip folder ", f.Name())
			continue
		}

		// 只读取符合标准的多语言文件, []string{"message.zh_CN.toml", "zh_CN"}
		matchs := re.FindStringSubmatch(f.Name())
		if len(matchs) < 2 {
			logrus.Info("language is empty, filename: ", f.Name())
			continue
		}

		// 文件路径
		filePath := baseDir + "/" + f.Name()

		// 语言
		lang := matchs[1]

		tag, err := language.Parse(lang)
		if err != nil {
			logrus.Error("parse lang file error, file: ", filePath)
			continue
		}

		if _, ok := i18nMap[lang]; !ok {

			bundle := i18n.NewBundle(tag)
			bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

			if _, err := bundle.LoadMessageFile(filePath); err != nil {
				logrus.Error("load lang file error, ", err)
				continue
			}

			localizer := i18n.NewLocalizer(bundle, lang)
			i18nMap[lang] = localizer
		}

		filePaths = append(filePaths, filePath)
	}

	logrus.Info("i18n init from file success, langs: ", i18nMap)
	return filePaths, nil
}

// 是否支持指定语言
func IsSupported(lang string) bool {
	if _, ok := i18nMap[lang]; ok {
		return true
	}

	return false
}

// 多语言转换
func Translate(lang string, key string) string {
	return translate(lang, key, nil)
}

func TranslateWithData(lang string, key string, data map[string]any) string {
	return translate(lang, key, data)
}

// 多语言转换及数据
func translate(lang string, key string, data map[string]any) string {
	defaultResult := lang + "_" + key

	if len(key) == 0 || len(lang) == 0 {
		logrus.Info("language or key is empty")
		return defaultResult
	}

	if localizer, ok := i18nMap[strings.ToLower(lang)]; ok {
		text, err := localizer.Localize(&i18n.LocalizeConfig{
			MessageID:    key,
			PluralCount:  1, // 英文系的有用，单数对应one, 复数再对应other, 非英文系的全都对应other
			TemplateData: data,
		})
		if err != nil {
			logrus.Error("language translate error, lang=", lang, ", key=", key, ", error=", err)
			return defaultResult
		}

		return text
	}

	logrus.Info("language not supported, defaultResult=", defaultResult)
	return defaultResult
}

// 监听文件变更, 只适合普通文件系统，不适合k8s挂载的configmap
func watchI18n(filePath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	if err = watcher.Add(filePath); err != nil {
		log.Fatal(err)
	}

	logrus.Info("watch i18n file: ", filePath)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// 监听文件修改
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				logrus.Info("file changed: ", event.Name)
				time.Sleep(100 * time.Millisecond) // 等待文件写入完成

				LoadI18n()
				logrus.Info("file reloaded: ", event.Name)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logrus.Error("watch i18n file error: ", err)
		}
	}
}
