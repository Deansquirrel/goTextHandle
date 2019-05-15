package object

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
)

import log "github.com/Deansquirrel/goToolLog"

const (
	ArgsFlagInstall   = "install"
	ArgsFlagUninstall = "uninstall"
)

type ProgramArgs struct {
	IsInstall   bool
	IsUninstall bool
}

func (pa *ProgramArgs) Definition() {
	flag.BoolVar(&pa.IsInstall, ArgsFlagInstall, false, "安装服务")
	flag.BoolVar(&pa.IsUninstall, ArgsFlagUninstall, false, "卸载服务")
}

func (pa *ProgramArgs) Parse() {
	flag.Parse()
}

func (pa *ProgramArgs) Check() error {
	//安装为服务和卸载服务参数不可同时存在
	if pa.IsInstall && pa.IsUninstall {
		return errors.New(fmt.Sprintf("参数 %s 和 %s 不可同时存在", ArgsFlagInstall, ArgsFlagUninstall))
	}
	return nil
}

func (pa *ProgramArgs) ToString() string {
	d, err := json.Marshal(pa)
	if err != nil {
		log.Warn(fmt.Sprintf("ProgramArgs转换为字符串时遇到错误：%s", err.Error()))
		return ""
	}
	return string(d)
}
