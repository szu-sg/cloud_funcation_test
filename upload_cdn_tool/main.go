package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"code.byted.org/ide/code_templates/upload_cdn_tool/config"
	"code.byted.org/ide/code_templates/upload_cdn_tool/oss"
	"code.byted.org/ide/code_templates/upload_cdn_tool/util"
	CdnUploadX "code.byted.org/net-fe/cdn-uploadx-go-sdk"
	"github.com/otiai10/copy"
	"golang.org/x/sync/errgroup"
	yaml "gopkg.in/yaml.v2"
)

type Region string

const (
	RegionCN = "cn"
	RegionSG = "sg"
	RegionVA = "va"
)

type Locale string

const (
	LocalZH = "zh"
	LocalEN = "en"

	projectTemplateDir = "project_template"
)

type RegionConf struct {
	Locale    Locale   `yaml:"locale"`
	Slug      []string `yaml:"slug"`
	CdnPrefix string   `yaml:"cdnPrefix"`
}

var (
	env          = "staging"
	version      = ""
	conf         = map[Region]RegionConf{}
	cdnClient    *CdnUploadX.DeliverClient
	cdnDir       = "ljhwZthlaukjlkulzlp/cloudide_space/project_template/"
	outputDir    = "output" // 指定要读取的文件夹路径
	templatesDir = "templates"
	imageDir     = "images"
	confDir      = "conf"
	confPath     = ""

	c config.Config

	cnOSS   oss.OSS
	cnNxOSS oss.OSS
	sgOSS   oss.OSS
	vaOSS   oss.OSS
)

func init() {
	// TODO 后续修改为从环境变量取或者运行参数取
	// 创建实例时初始化默认参数
	cdnClient = CdnUploadX.NewClient(CdnUploadX.CdnUploadXClientConfig{
		Region:      "VA2",
		Email:       "wangjiayi.yume@bytedance.com",
		AutoRefresh: "true",
		IsOfficeNet: "false",
	})

	c = config.Config{}
	err := config.LoadConfig(&c, "conf", "CUSTOM")
	if err == nil {
		fmt.Printf("load config finish: %+v", c)
	}

	if c.ENV == "prod" {
		env = "prod"
	}
	if c.Version == "" {
		ExitIfErr(fmt.Errorf("version is empty"))
	}
	version = c.Version

	if c.OssCN.Enbale {
		if c.OssCN.AccesskeyID == "" || c.OssCN.SecretAccessKey == "" || c.OssCN.Region == "" || c.OssCN.BucketName == "" || c.OssCN.DownloadURLPrefix == "" {
			ExitIfErr(fmt.Errorf("cn oss config is empty"))
			return
		}
		cnOSS, err = oss.NewS3OSS(c.OssCN.AccesskeyID, c.OssCN.SecretAccessKey, c.OssCN.Region, c.OssCN.BucketName, c.OssCN.DownloadURLPrefix)
		if err != nil {
			ExitIfErr(fmt.Errorf("init cn oss failed: %v", err))
			return
		}
	}
	if c.OssCNNX.Enbale {
		if c.OssCNNX.AccesskeyID == "" || c.OssCNNX.SecretAccessKey == "" || c.OssCNNX.Region == "" || c.OssCNNX.BucketName == "" || c.OssCNNX.DownloadURLPrefix == "" {
			ExitIfErr(fmt.Errorf("cn oss config is empty"))
			return
		}
		cnNxOSS, err = oss.NewS3OSS(c.OssCNNX.AccesskeyID, c.OssCNNX.SecretAccessKey, c.OssCNNX.Region, c.OssCNNX.BucketName, c.OssCNNX.DownloadURLPrefix)
		if err != nil {
			ExitIfErr(fmt.Errorf("init cn oss failed: %v", err))
			return
		}
	}
	if c.OssSG.Enbale {
		if c.OssSG.AccesskeyID == "" || c.OssSG.SecretAccessKey == "" || c.OssSG.Region == "" || c.OssSG.BucketName == "" || c.OssSG.DownloadURLPrefix == "" {
			ExitIfErr(fmt.Errorf("sg oss config is empty"))
			return
		}
		sgOSS, err = oss.NewS3OSS(c.OssSG.AccesskeyID, c.OssSG.SecretAccessKey, c.OssSG.Region, c.OssSG.BucketName, c.OssSG.DownloadURLPrefix)
		if err != nil {
			ExitIfErr(fmt.Errorf("init sg oss failed: %v", err))
			return
		}
	}
	if c.OssVA.Enbale {
		if c.OssVA.AccesskeyID == "" || c.OssVA.SecretAccessKey == "" || c.OssVA.Region == "" || c.OssVA.BucketName == "" || c.OssVA.DownloadURLPrefix == "" {
			ExitIfErr(fmt.Errorf("va oss config is empty"))
			return
		}
		vaOSS, err = oss.NewS3OSS(c.OssVA.AccesskeyID, c.OssVA.SecretAccessKey, c.OssVA.Region, c.OssVA.BucketName, c.OssVA.DownloadURLPrefix)
		if err != nil {
			ExitIfErr(fmt.Errorf("init va oss failed: %v", err))
			return
		}
	}
}

func main() {
	var (
		regionStr = ""
		regions   = []Region{}
	)

	flag.StringVar(&confDir, "conf", "./conf", "配置文件路径")
	flag.StringVar(&regionStr, "region", "", "上传区域, 使用`,`分隔多个区域")
	flag.Parse()

	if confDir == "" {
		ExitIfErr(fmt.Errorf("conf path is empty\n"))
		return
	}
	confPath = path.Join(confDir, "publish_conf.yaml")
	if regionStr == "" {
		regionStr = c.Region
		if regionStr == "" {
			ExitIfErr(fmt.Errorf("region is empty\n"))
			return
		}
	}
	if version == "" {
		ExitIfErr(fmt.Errorf("version is empty\n"))
		return
	}

	regionStrs := strings.Split(regionStr, ",")
	for _, regionStr := range regionStrs {
		regions = append(regions, Region(regionStr))
	}

	err := LoadConf(confPath)
	if err != nil {
		ExitIfErr(fmt.Errorf("load conf failed: %v\n", err))
		return
	}

	// cdnClient.GetCurrentDir(CdnUploadX.WithDir(cdnDir), CdnUploadX.WithEmail("liyuxing@bytedance.com"))
	for _, region := range regions {
		outputRegionDir := path.Join(outputDir, string(region))
		err = fileProcess(string(region), joinEnvAndVersion(conf[region].CdnPrefix), outputRegionDir)
		if err != nil {
			fmt.Printf("file process error: %v\n", err)
			ExitIfErr(fmt.Errorf("file process error: %v", err))
			return
		}
		err = uploadCDN(region, path.Join(outputDir, string(region)))
		if err != nil {
			fmt.Printf("upload cdn error: %v\n", err)
			ExitIfErr(fmt.Errorf("upload cdn error: %v", err))
		}
	}

}

func LoadConf(confPath string) error {
	content, err := os.ReadFile(confPath)
	if err != nil {
		return fmt.Errorf("read conf file failed: %v, conf path: %s", err, confPath)
	}
	err = yaml.Unmarshal(content, &conf)
	if err != nil {
		return fmt.Errorf("unmarshal conf file failed: %v, content: %v", err, string(content))
	}
	return nil
}

func uploadCDN(region Region, outputRedionDir string) error {
	cdnRegion := ""
	var oss []oss.OSS
	switch region {
	case RegionCN:
		cdnRegion = "CN"
		oss = append(oss, cnOSS)
		if cnNxOSS != nil {
			oss = append(oss, cnNxOSS)
		}
	case RegionSG:
		cdnRegion = "SG"
		oss = append(oss, sgOSS)
	case RegionVA:
		cdnRegion = "VA2"
		oss = append(oss, vaOSS)
	default:
		return fmt.Errorf("unknown region: %s", region)
	}

	eg := errgroup.Group{}

	err := filepath.Walk(outputRedionDir, func(fpath string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if f.IsDir() || f.Name() == "bin" {
			return nil
		}
		targetFile := path.Join(env, version, strings.Replace(fpath, outputRedionDir+"/", "", 1))

		eg.Go(func() error {
			// fmt.Println("cdndir: " + cdnDir)
			// fmt.Println("path: " + path)
			// fmt.Println("targetFile: " + targetFile)
			// fmt.Println("cdnRegion: " + cdnRegion)
			_, resp, err := cdnClient.Upload(CdnUploadX.WithDir(cdnDir), CdnUploadX.WithFilePath(fpath), CdnUploadX.WithFileName(targetFile), CdnUploadX.WithRegion(cdnRegion))
			if err != nil {
				return fmt.Errorf("upload cdn error: %v, resp: %s", err, string(resp.Body))
			}

			// 将文件转换为 io.Reader
			for _, o := range oss {
				// 上传对象存储
				file, err := os.Open(fpath)
				if err != nil {
					return err
				}
				defer file.Close()
				reader := io.Reader(file)
				name := path.Join(projectTemplateDir, targetFile)
				u, err := o.UploadObject(context.Background(), name, reader, f.Size())
				if err != nil {
					return fmt.Errorf("upload object %s failed: %v", fpath, err)
				}
				fmt.Printf("upload oss object %s success, url: %s\n", fpath, u)
			}
			return nil
		})
		time.Sleep(time.Millisecond * 200)
		return nil
	})

	if err = eg.Wait(); err != nil {
		return err
	}
	return err
}

func fileProcess(region string, replaceURLPrefix string, outputRedionDir string) error {
	// 首先将 templates copy 到 /tmp/templates 路径处理
	tmpTemplatesDir := "./tmp/templates"
	tmpImagesDir := "./tmp/images"
	listPath := path.Join(templatesDir+"/", "list.txt")
	output, err := util.ExecShell(fmt.Sprintf(`
pwd
rm -rf %s 
mkdir -p %s %s
cp -r %s/. %s
cp -r %s/. %s
echo "" > %s
`,
		tmpTemplatesDir,
		tmpTemplatesDir, tmpImagesDir,
		templatesDir, tmpTemplatesDir,
		imageDir, tmpImagesDir,
		listPath))
	if err != nil {
		return fmt.Errorf("copy templates failed: %v, output: %s", err, output)
	}

	// 处理 cn-zh readme 文件
	/*
		对于 非 cn 区域，删除 README-zh.md 文件
		对于 cn 区域，如果存在 README-zh.md 文件，则替换 README.md 文件
		对于 README-zh markdown，将文档中引用的图片下载到 output/cn/images/xxxx 中
	*/
	for _, t := range conf[Region(region)].Slug {
		err := processReadme(region, path.Join(tmpTemplatesDir, t), path.Join(tmpImagesDir, t), t)
		if err != nil {
			return fmt.Errorf("process readme failed: %v, region: %s, template: %s", err, region, t)
		}
	}

	// 处理 templates 文件
	/*
		1. 对 ./templates 下的所有子路径做以下操作
			1). 处理 ReadME 文件，替换图片链接，将相对路径替换为绝对路径
			2). 将 ReadME 提取至 output
			3). 打包代码至 output
		2. cp image 至 output
	*/

	err = filepath.Walk(tmpTemplatesDir, func(p string, f os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk dir failed: %v, path: %s", err, p)
		}
		if !f.IsDir() || f.Name() == "templates" {
			return nil
		}
		fmt.Println("will process template " + f.Name())
		// 判断 README 文件是否存在
		readmePath := path.Join(tmpTemplatesDir, f.Name(), "README.md")
		if _, err := os.Stat(readmePath); err != nil {
			return nil
		}
		contentBytes, err := os.ReadFile(readmePath)
		if err != nil {
			return fmt.Errorf("read file failed: %v, path: %s", err, p)
		}
		content := string(contentBytes)
		content = strings.ReplaceAll(content, "](../../", "]("+replaceURLPrefix)
		err = os.WriteFile(readmePath, []byte(content), 0644)
		if err != nil {
			return fmt.Errorf("write file failed: %v, path: %s", err, p)
		}

		// 提取 README
		outputPath := path.Join(outputRedionDir, f.Name()+"_README.md")
		copy.Copy(readmePath, outputPath)

		// 打包到 output
		cmd := fmt.Sprintf(`
workpath=$(pwd)
cd %s
tar -czvf "$workpath/%s/%s.tar.gz" ./
echo "%s.tar.gz" >> $workpath/%s
		`,
			path.Join(tmpTemplatesDir, f.Name()), outputRedionDir, f.Name(), f.Name(), listPath)
		out, err := util.ExecShell(cmd)
		if err != nil {
			return fmt.Errorf("tar script failed: %v, output: %s, script: %s", err, out, cmd)
		}
		fmt.Println("tar output: " + out)
		return nil
	})
	if err != nil {
		return fmt.Errorf("process templates failed: %v", err)
	}
	// cp 图片至 output
	err = copy.Copy(tmpImagesDir, path.Join(outputRedionDir, "images"))
	if err != nil {
		return fmt.Errorf("copy images failed: %v", err)
	}

	// 处理 conf 配置文件（各模板配置文件写入到 output，发布模板时读取配置，更新或发布模板）
	// 将 conf/templates 下的模板配置文件写入到 output/$region/ 下
	confOutputDir := path.Join(outputRedionDir, "conf")
	err = os.MkdirAll(confOutputDir, 0755)
	if err != nil {
		return fmt.Errorf("mkdir output conf dir %s failed: %v", confOutputDir, err)
	}
	regionConfig := conf[Region(region)]
	body, err := json.Marshal(regionConfig)
	if err != nil {
		return fmt.Errorf("marshal region config failed: %v", err)
	}
	err = os.WriteFile(path.Join(confOutputDir, "publish_conf.json"), body, 0644)
	if err != nil {
		return fmt.Errorf("write region config to %s failed: %v", path.Join(confOutputDir, "publish_conf.json"), err)
	}
	for _, t := range conf[Region(region)].Slug {
		tconf := path.Join(confDir, "templates", t+".json")
		confOutputPath := path.Join(confOutputDir, t+".json")
		err = copy.Copy(tconf, confOutputPath)
		if err != nil {
			return fmt.Errorf("copy conf from %s to %s failed: %v", tconf, confOutputPath, err)
		}
	}
	return nil
}

func ExitIfErr(errs ...error) {
	msgs := []string{}
	for i, err := range errs {
		if err != nil {
			msgs = append(msgs, fmt.Sprintf("[%d]: %s", i, err.Error()))
		}
	}
	if len(msgs) != 0 {
		msgStr := strings.Join(msgs, "\n")
		// 程序直接退出场景，同时打印到标准输出，否则，在 k8s 中，看不到错误信息。
		fmt.Printf("%s %s\n", time.Now().Format(time.RFC3339Nano), msgStr)
		log.Panic(msgStr)
		log.Info()
	}
}

func joinEnvAndVersion(cdnPrefix string) string {
	return cdnPrefix + env + "/" + version + "/"
}

func processReadme(region string, templatePath string, imagePath string, template string) error {
	// 如果存在 README-{region}.md 文件，则将其保存为 README.md
	// 删除除了 README.md 之外的其余所有 README 文件
	readmePath := path.Join(templatePath, "README.md")
	regionREADME := path.Join(templatePath, "README-"+region+".md")
	_, err := os.Stat(regionREADME)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("stat %s failed: %v", regionREADME, err)
		}
	} else {
		err = os.Rename(regionREADME, readmePath)
		if err != nil {
			return fmt.Errorf("rename %s to %s failed: %v", regionREADME, readmePath, err)
		}
	}
	rmCmd := fmt.Sprintf("rm -f %s/README-*.md", templatePath)
	out, err := util.ExecShell(rmCmd)
	if err != nil {
		return fmt.Errorf("rm readme failed: %v, output: %s", err, out)
	}

	contentByte, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("read readme file %s failed: %v", readmePath, err)
	}
	content := string(contentByte)

	wget := func(url string, imageName string) error {
		cmd := fmt.Sprintf("wget %s -O %s", url, path.Join(imagePath, imageName))
		output, err := util.ExecShell(cmd)
		if err != nil {
			return fmt.Errorf("wget %s failed: %v, output: %s", url, err, output)
		}
		return nil
	}

	regex := regexp.MustCompile(`!\[.*?\]\((.*?)\)`)
	matches := regex.FindAllStringSubmatch(content, -1)

	for idx, match := range matches {
		url := match[1]
		if strings.HasPrefix(url, "../../images") {
			continue
		}
		newImageName := "image-" + strconv.Itoa(idx) + ".jpg"
		err := wget(url, newImageName)
		if err != nil {
			return fmt.Errorf("template %s get image failed: %v, region: %s", template, err, region)
		}
		content = strings.ReplaceAll(content, url, "../../images/"+template+"/"+newImageName)
	}
	err = os.WriteFile(readmePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("write %s zh readme file failed: %v", template, err)
	}
	return nil
}
