package model

import (
	"flag"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/admpub/chardet"
	"github.com/admpub/log"
	sc "github.com/admpub/mahonia"
)

var CmdOptions = &Options{}

type Options struct {
	TargetFilePath  string //目标文件路径
	TargetFileRule  string //目标文件名称规则
	IgnoreFileRule  string
	FindRule        string //内容查询规则
	ReplaceWith     string //替换为
	SaveToPath      string //将匹配到的文件按照相同的目录结构另存
	CompressSave    bool   //是否保存为压缩文件
	ConvertEncoding string //转换编码(例如：utf-8->gbk，也可以采用"->gbk"来自动确认原始文件的编码)
	ReplaceMode     bool
	RestoreVer      string
	ClearBackup     bool
}

func (o *Options) DefineFlag() {
	flag.StringVar(&o.TargetFilePath, `findPath`, ``, `搜索路径`)
	flag.StringVar(&o.TargetFileRule, `fileRule`, ``, `文件名称规则(正则表达式)`)
	flag.StringVar(&o.IgnoreFileRule, `ignoreRule`, ``, `要忽略的文件名称规则(正则表达式)`)
	flag.StringVar(&o.FindRule, `contentRule`, ``, `内容搜索规则(正则表达式)`)
	flag.StringVar(&o.ReplaceWith, `replaceWith`, ``, `替换为`)
	flag.StringVar(&o.SaveToPath, `savePath`, ``, `搜索到的文件保存路径`)
	flag.BoolVar(&o.CompressSave, `compress`, false, `是否将文件保存为压缩包`)
	flag.StringVar(&o.ConvertEncoding, `convertEncoding`, ``, `内容编码转换(例如：utf-8->gbk，也可以采用"->gbk"来自动确认原始文件的编码)`)
	flag.BoolVar(&o.ReplaceMode, `replace`, false, `是否替换`)
	flag.StringVar(&o.RestoreVer, `restore`, ``, `还原备份文件`)
	flag.BoolVar(&o.ClearBackup, `clearBackup`, false, `删除所有备份文件`)
}

func (o *Options) Run() error {

	var regexpFileName, regexpContent, regexpIgnoreFile *regexp.Regexp
	if len(CmdOptions.TargetFileRule) > 0 {
		regexpFileName = regexp.MustCompile(CmdOptions.TargetFileRule)
	}
	if len(CmdOptions.FindRule) > 0 {
		regexpContent = regexp.MustCompile(CmdOptions.FindRule)
	}
	if len(CmdOptions.IgnoreFileRule) > 0 {
		regexpIgnoreFile = regexp.MustCompile(CmdOptions.IgnoreFileRule)
	}

	// fixed bug
	if CmdOptions.ReplaceMode && len(CmdOptions.ReplaceWith) == 0 {
		var i int
		for k, v := range os.Args {
			if v == `-replaceWith` {
				i = k
				break
			}
		}
		if i > 0 && i < len(os.Args)-1 {
			CmdOptions.ReplaceWith = os.Args[i+1]
		}
	}
	replaceWithBytes := []byte(CmdOptions.ReplaceWith)

	doFn := func(b []byte) []byte {
		return b
	}
	if len(CmdOptions.ConvertEncoding) > 0 {
		encs := strings.SplitN(CmdOptions.ConvertEncoding, `->`, 2)
		var fromEnc, toEnc string
		if len(encs) == 2 {
			fromEnc = strings.TrimSpace(encs[0])
			toEnc = strings.TrimSpace(encs[1])
			if len(fromEnc) == 0 {
				doFn = func(b []byte) []byte {
					charset := chardet.Mostlike(b)
					dec := sc.NewDecoder(charset)
					s := dec.ConvertString(string(b))
					enc := sc.NewEncoder(toEnc)
					s = enc.ConvertString(s)
					b = []byte(s)
					return b
				}
			} else {
				doFn = func(b []byte) []byte {
					dec := sc.NewDecoder(fromEnc)
					s := dec.ConvertString(string(b))
					enc := sc.NewEncoder(toEnc)
					s = enc.ConvertString(s)
					b = []byte(s)
					return b
				}
			}
		}
	}
	if len(CmdOptions.TargetFilePath) == 0 {
		CmdOptions.TargetFilePath = `./`
	}
	root, err := filepath.Abs(CmdOptions.TargetFilePath)
	if err != nil {
		log.Fatal(err)
	}
	log.Info(`Scan dir: `, root)
	timeString := time.Now().Format(`20060102150405`)
	//单纯压缩指定文件
	if regexpContent == nil && CmdOptions.CompressSave {
		if len(CmdOptions.SaveToPath) == 0 {
			CmdOptions.SaveToPath = `./`
		}
		savePath := filepath.Join(CmdOptions.SaveToPath, `compress.zip`)
		log.Info(`Compressed ` + CmdOptions.TargetFilePath + ` and save to ` + savePath + `.`)
		_, err = Zip(CmdOptions.TargetFilePath, savePath, regexpFileName, regexpIgnoreFile)
		if err != nil {
			log.Error(err)
		}
		log.Info(`Compression is completed.`)
	} else if len(CmdOptions.RestoreVer) > 0 { //还原备份文件
		regexpRestoreFile := regexp.MustCompile(`\.` + regexp.QuoteMeta(CmdOptions.RestoreVer) + `\.fbak$`)
		err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			nameBytes := []byte(info.Name())
			if !regexpRestoreFile.Match(nameBytes) {
				return nil
			}
			original := strings.TrimSuffix(info.Name(), `.`+CmdOptions.RestoreVer+`.fbak`)
			os.Remove(original)
			log.Info(`Restore ` + original + ` from ` + path + `.`)
			return err
		})
		log.Info(`Restore file is completed.`)
	} else if CmdOptions.ClearBackup { //删除所有备份文件
		regexpRestoreFile := regexp.MustCompile(`\.[0-9]+\.fbak$`)
		err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			nameBytes := []byte(info.Name())
			if !regexpRestoreFile.Match(nameBytes) {
				return nil
			}
			log.Info(`Removed ` + path + `.`)
			return os.Remove(path)
		})
		log.Info(`Backup file is removed.`)
	} else {
		err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			name := info.Name()
			nameBytes := []byte(name)
			if info.IsDir() {
				if regexpIgnoreFile != nil && regexpIgnoreFile.Match(nameBytes) {
					return filepath.SkipDir
				}
				return nil
			} else if regexpIgnoreFile != nil {
				if regexpIgnoreFile.Match(nameBytes) {
					return nil
				}
			}
			if regexpFileName != nil && !regexpFileName.Match(nameBytes) {
				return nil
			}
			savePath := CmdOptions.SaveToPath
			if regexpContent != nil {
				b, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				if !regexpContent.Match(b) {
					return nil
				}
				if CmdOptions.ReplaceMode {
					b = regexpContent.ReplaceAll(b, replaceWithBytes)
				}
				b = doFn(b)
				writeOriginalFile := false
				if len(savePath) == 0 && !CmdOptions.CompressSave { //如果没有指定另存路径,且不需要压缩保存修改后的文件，则覆盖原文件
					savePath = path
					writeOriginalFile = true
				} else {
					if CmdOptions.CompressSave {
						savePath = filepath.Join(savePath, `_tmp`)
						err := os.MkdirAll(savePath, os.ModePerm)
						if err != nil {
							return err
						}
					}
					filePath := strings.TrimPrefix(path, root)
					savePath = filepath.Join(savePath, filePath)
					err := os.MkdirAll(filepath.Dir(savePath), os.ModePerm)
					if err != nil {
						return err
					}
				}
				log.Info(`Modified ` + path + ` and save to ` + savePath + `.`)
				if writeOriginalFile {
					//写原始文件时，先备份
					err = os.Rename(path, path+`.`+timeString+`.fbak`)
					if err != nil {
						return err
					}
				}
				err = ioutil.WriteFile(savePath, b, os.ModePerm)
				if err != nil {
					return err
				}
			} else if len(savePath) > 0 { //不需要访问内容，直接拷贝文件
				filePath := strings.TrimPrefix(root, path)
				savePath = filepath.Join(savePath, filePath)
				err := os.MkdirAll(filepath.Dir(savePath), os.ModePerm)
				if err != nil {
					return err
				}
				sr, err := os.Open(path)
				if err != nil {
					return err
				}
				defer sr.Close()

				dw, err := os.Create(savePath)
				if err != nil {
					return err
				}
				defer dw.Close()

				if _, err = io.Copy(dw, sr); err != nil {
					return err
				}
				log.Info(`Copy ` + path + ` to ` + savePath + `.`)
			}
			return nil
		})
		done := make(chan int)
		block := false
		if err == nil && CmdOptions.CompressSave {
			srcPath := filepath.Join(CmdOptions.SaveToPath, `_tmp`)
			savePath := filepath.Join(CmdOptions.SaveToPath, `compress.zip`)
			log.Info(`Compressed ` + srcPath + ` and save to ` + savePath + `.`)
			_, err = Zip(srcPath, savePath)
			if err == nil {
				log.Info(`Delete ` + srcPath + `.`)
				err = os.RemoveAll(srcPath)
				if err != nil {
					block = true
					go func() {
						for i := 1; err != nil && i < 10; i++ {
							log.Error(err)
							time.Sleep(1 * time.Second)
							err = os.RemoveAll(srcPath)
						}
						close(done)
					}()
				}
			}
		}

		if err != nil {
			log.Error(err)
		}
		log.Info(`Find complete.`)
		if block {
			<-done
		}
	}
	return err
}
