package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"code.byted.org/ide/code_templates/upload_cdn_tool/util"
	"github.com/caarlos0/env/v6"
	"github.com/mcuadros/go-defaults"
	"gopkg.in/yaml.v2"
)

type OSSConfig struct {
	AccesskeyID       string `yaml:"accesskey_id" env:"ACCESS_KEY_ID"`
	SecretAccessKey   string `yaml:"secret_access_key" env:"SECRET_ACCESS_KEY"`
	Region            string `yaml:"region" env:"REGION"`
	BucketName        string `yaml:"bucket_name" env:"BUCKET_NAME"`
	DownloadURLPrefix string `yaml:"download_url_prefix" env:"DOWNLOAD_URL_PREFIX"`
	Enbale            bool   `yaml:"enable" env:"ENABLE" default:"false"`
}

type Config struct {
	ENV     string    `yaml:"ENV" env:"ENV" default:"staging"`
	Version string    `yaml:"Version" env:"VERSION"`
	Region  string    `yaml:"Region" env:"REGION"`
	OssCN   OSSConfig `yaml:"OssCN" envPrefix:"OSS_CN_"`
	OssCNNX OSSConfig `yaml:"OssCN" envPrefix:"OSS_CN_NX_"`
	OssSG   OSSConfig `yaml:"OssSG" envPrefix:"OSS_SG_"`
	OssVA   OSSConfig `yaml:"OssVA" envPrefix:"OSS_VA_"`
}

// LoadConfig 加载配置
func LoadConfig(data interface{}, confDir string, envPrefix string) error {
	// 1. 填充默认值
	resolveDefaultValue(data)
	// 2. 读取 yaml 配置文件
	confProfile := os.Getenv(envPrefix + "_DEPLOY_ENV")
	confFilename := "conf.yaml"
	if confProfile != "" {
		confFilename = fmt.Sprintf("conf-%s.yaml", confProfile)
	}
	err := resolveYAMLFileValue(data, path.Join(confDir, confFilename))
	if err != nil {
		return err
	}
	// 3. 读取环境变量
	err = resolveEnvValue(data, envPrefix)
	if err != nil {
		return err
	}
	// 4. 特殊值处理
	err = expandValue(data)
	if err != nil {
		return err
	}
	return nil
}

func resolveDefaultValue(data interface{}) {
	defaults.SetDefaults(data)
}

func resolveYAMLFileValue(data interface{}, confFilePath string) error {
	_, err := os.Stat(confFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warnf("load config from %s, file not exist, skip it", confFilePath)
			return nil
		}
		return err
	}
	confContent, err := os.ReadFile(confFilePath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(confContent, data)
	if err != nil {
		return err
	}
	return nil
}

func resolveEnvValue(data interface{}, envPrefix string) error {
	return env.Parse(data, env.Options{Prefix: envPrefix + "_"})
}

// 文件必须是以 / ../ ./ ~/ 开头，后面一个字符不能是 /
var ReIsFileLike = regexp.MustCompile(`^@(~/|\./|\.\./|/)[\w.~-]`)
var ReFilePath = regexp.MustCompile(`^[\w-./\\~@]+$`)
var ReLabel = regexp.MustCompile(`^[a-z0-9A-Z][A-Za-z0-9_.-]*[a-z0-9A-Z]$`) // k8s label rule
var HomeDir, _ = os.UserHomeDir()

// expandValue expand special values, like @file etc.
func expandValue(data interface{}) error {
	// env包已经确保了 data.Kind 一定是 Ptr + Struct
	values := reflect.ValueOf(data).Elem()
	types := values.Type()
	for i := 0; i < values.NumField(); i++ {
		// key := types.Field(i).Name
		value := values.Field(i)
		keyType := types.Field(i).Type.Name()
		switch keyType {
		default:
			// skip
		case "string":
			// 借鉴curl语法：以"@"开头表示文件，后面可传Query带参数
			// @~/.kube/config?encode=base64
			// 虽然 env 包也可展开文件，但无法处理成 base64，也无法展开"~"
			str := value.String()
			if ReIsFileLike.MatchString(str) {
				u, err := url.Parse(strings.TrimPrefix(str, "@"))
				if err != nil {
					return err
				}
				file := u.Path
				if !ReFilePath.MatchString(file) {
					return errors.New("invalid file path:" + file)
				}

				if strings.HasPrefix(file, "~") {
					file = strings.Replace(file, "~", HomeDir, 1)
				}
				b, err := os.ReadFile(file)
				if err != nil {
					return err
				}

				if u.Query().Get("encode") == "base64" {
					value.SetString(base64.StdEncoding.EncodeToString(b))
				} else {
					value.SetString(util.UnsafeSliceToString(b))
				}
			}
		}
	}
	return nil
}
