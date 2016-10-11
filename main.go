package main

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
	"github.com/mitool/find2/model"
)

func main() {
	model.CmdOptions.DefineFlag()
	flag.Parse()
	log.Sync()

	var regexpFileName, regexpContent, regexpIgnoreFile *regexp.Regexp
	if model.CmdOptions.TargetFileRule != nil && len(*model.CmdOptions.TargetFileRule) > 0 {
		regexpFileName = regexp.MustCompile(*model.CmdOptions.TargetFileRule)
	}
	if model.CmdOptions.FindRule != nil && len(*model.CmdOptions.FindRule) > 0 {
		regexpContent = regexp.MustCompile(*model.CmdOptions.FindRule)
	}
	if model.CmdOptions.IgnoreFileRule != nil && len(*model.CmdOptions.IgnoreFileRule) > 0 {
		regexpIgnoreFile = regexp.MustCompile(*model.CmdOptions.IgnoreFileRule)
	}
	replaceWithBytes := []byte{}
	if model.CmdOptions.ReplaceWith == nil {
		replaceWithBytes = []byte(``)
	} else {
		replaceWithBytes = []byte(*model.CmdOptions.ReplaceWith)
	}
	doFn := func(b []byte) []byte {
		return b
	}
	if model.CmdOptions.ConvertEncoding != nil {
		encs := strings.SplitN(*model.CmdOptions.ConvertEncoding, `->`, 2)
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
	if model.CmdOptions.TargetFilePath == nil || len(*model.CmdOptions.TargetFilePath) == 0 {
		s := `./`
		model.CmdOptions.TargetFilePath = &s
	}
	root, err := filepath.Abs(*model.CmdOptions.TargetFilePath)
	if err != nil {
		log.Fatal(err)
	}
	log.Info(`Scan dir: `, root)
	timeString := time.Now().Format(`20060102150405`)
	//单纯压缩指定文件
	if regexpContent == nil && *model.CmdOptions.CompressSave {
		if len(*model.CmdOptions.SaveToPath) == 0 {
			*model.CmdOptions.SaveToPath = `./`
		}
		savePath := filepath.Join(*model.CmdOptions.SaveToPath, `compress.zip`)
		log.Info(`Compressed ` + *model.CmdOptions.TargetFilePath + ` and save to ` + savePath + `.`)
		_, err = model.Zip(*model.CmdOptions.TargetFilePath, savePath, regexpFileName, regexpIgnoreFile)
		if err != nil {
			log.Error(err)
		}
		log.Info(`Compression is completed.`)
	} else if model.CmdOptions.RestoreVer != nil && len(*model.CmdOptions.RestoreVer) > 0 { //还原备份文件
		regexpRestoreFile := regexp.MustCompile(`\.` + regexp.QuoteMeta(*model.CmdOptions.RestoreVer) + `\.fbak$`)
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
			original := strings.TrimSuffix(info.Name(), `.`+*model.CmdOptions.RestoreVer+`.fbak`)
			os.Remove(original)
			log.Info(`Restore ` + original + ` from ` + path + `.`)
			return err
		})
		log.Info(`Restore file is completed.`)
	} else if model.CmdOptions.ClearBackup != nil && *model.CmdOptions.ClearBackup { //删除所有备份文件
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
			savePath := *model.CmdOptions.SaveToPath
			if regexpContent != nil {
				b, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				if *model.CmdOptions.ReplaceMode {
					b = regexpContent.ReplaceAll(b, replaceWithBytes)
				} else {
					if !regexpContent.Match(b) {
						return nil
					}
				}
				b = doFn(b)
				writeOriginalFile := false
				if len(savePath) == 0 && !*model.CmdOptions.CompressSave { //如果没有指定另存路径,且不需要压缩保存修改后的文件，则覆盖原文件
					savePath = path
					writeOriginalFile = true
				} else {
					if *model.CmdOptions.CompressSave {
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
		if err == nil && *model.CmdOptions.CompressSave {
			srcPath := filepath.Join(*model.CmdOptions.SaveToPath, `_tmp`)
			savePath := filepath.Join(*model.CmdOptions.SaveToPath, `compress.zip`)
			log.Info(`Compressed ` + srcPath + ` and save to ` + savePath + `.`)
			_, err = model.Zip(srcPath, savePath)
			if err == nil {
				log.Info(`Delete ` + srcPath + `.`)
				err = os.RemoveAll(srcPath)
				if err != nil {
					go func() {
						for i := 1; err != nil && i < 10; i++ {
							log.Error(err)
							time.Sleep(1 * time.Second)
							err = os.RemoveAll(srcPath)
						}
						close(done)
					}()
				} else {
					close(done)
				}
			}
		}

		if err != nil {
			log.Error(err)
		}
		log.Info(`Find complete.`)
		<-done
	}
}
