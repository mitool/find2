package main

import (
	"flag"
	"fmt"
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
	//
	log.Info(*model.CmdOptions.TargetFilePath, `----`)
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
	err := filepath.Walk(*model.CmdOptions.TargetFilePath, func(path string, info os.FileInfo, err error) error {
		nameBytes := []byte(info.Name())
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
		if regexpContent != nil {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			b = regexpContent.ReplaceAll(b, replaceWithBytes)

		}
		saveAs := strings.TrimPrefix(path, root)
		saveAs = filepath.Join(save, saveAs)
		err = os.MkdirAll(filepath.Dir(saveAs), os.ModePerm)
		if err == nil {
			file, err := os.Create(saveAs)
			if err == nil {
				_, err = file.WriteString(content)
			}
		}
		if err != nil {
			return err
		}
		fmt.Println(`Autofix ` + path + `.`)
		return nil
	})

	defer time.Sleep(5 * time.Minute)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(`Autofix complete.`)
}
