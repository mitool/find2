package model

import (
	"flag"
)

var CmdOptions = &Options{}

type Options struct {
	TargetFilePath  *string //目标文件路径
	TargetFileRule  *string //目标文件名称规则
	IgnoreFileRule  *string
	FindRule        *string //内容查询规则
	ReplaceWith     *string //替换为
	SaveToPath      *string //将匹配到的文件按照相同的目录结构另存
	CompressSave    *bool   //是否保存为压缩文件
	ConvertEncoding *string //转换编码(例如：utf-8->gbk，也可以采用"->gbk"来自动确认原始文件的编码)
	ReplaceMode     *bool
	RestoreVer      *string
	ClearBackup     *bool
}

func (o *Options) DefineFlag() {
	o.TargetFilePath = flag.String(`findPath`, ``, `搜索路径`)
	o.TargetFileRule = flag.String(`fileRule`, ``, `文件名称规则(正则表达式)`)
	o.IgnoreFileRule = flag.String(`ignoreRule`, ``, `要忽略的文件名称规则(正则表达式)`)
	o.FindRule = flag.String(`contentRule`, ``, `内容搜索规则(正则表达式)`)
	o.ReplaceWith = flag.String(`replaceWith`, ``, `替换为`)
	o.SaveToPath = flag.String(`savePath`, ``, `搜索到的文件保存路径`)
	o.CompressSave = flag.Bool(`compress`, false, `是否将文件保存为压缩包`)
	o.ConvertEncoding = flag.String(`convertEncoding`, ``, `内容编码转换(例如：utf-8->gbk，也可以采用"->gbk"来自动确认原始文件的编码)`)
	o.ReplaceMode = flag.Bool(`replace`, false, `是否替换`)
	o.RestoreVer = flag.String(`restore`, ``, `还原备份文件`)
	o.ClearBackup = flag.Bool(`clearBackup`, false, `删除所有备份文件`)
}
