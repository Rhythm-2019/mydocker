package container

import (
	"github.com/rhythm_2019/mydocker/toolkit"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type ContainerManager struct {
	FullCmd     string
	FullcmdArgs []string
}

func (cm *ContainerManager) Init() {
	// 初始化：获取启动命令、初始化Mount namespace

	// 获取启动命令、
	fd := os.NewFile(uintptr(3), "pipe")
	fullCmdBytes, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Fatalf("init: read cmd failed, detail is %v", err)
	}
	defer fd.Close()

	cm.FullCmd = string(fullCmdBytes)
	cm.FullcmdArgs = strings.Split(cm.FullCmd, " ")

	// 自动补全前面的目录位置，如用户输入bash，需要替换为/bin/bash
	path, err := exec.LookPath(cm.FullcmdArgs[0])
	if err != nil {
		log.Fatalf("init: %s command no found, detail is %v", cm.FullcmdArgs[0], err)
	}
	cm.FullcmdArgs[0] = path

	targetRoot, err := os.Getwd()
	if err != nil {
		log.Fatalf("init: get process work dir failed, detail is %v", err)
	}
	err = toolkit.PivotRoot(targetRoot)
	if err != nil {
		log.Fatalf("init: pivot root (%s) failed, detail is %v", targetRoot, err)
	}

	// mount proc
	flags := syscall.MS_NOEXEC | syscall.MS_NODEV | syscall.MS_NOSUID
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(flags), ""); err != nil {
		log.Fatal("mount:", err)
	}
	// mount dev
	flags = syscall.MS_NOEXEC | syscall.MS_STRICTATIME
	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", uintptr(flags), ""); err != nil {
		log.Fatal("mount:", err)
	}
}

func (cm *ContainerManager) Start() {
	// os.Environ() 复制一份环境变量
	log.Infof("run a command %s", cm.FullcmdArgs)
	if err := syscall.Exec(cm.FullcmdArgs[0], cm.FullcmdArgs, os.Environ()); err != nil {
		log.Fatal("exec:", err)
	}
}
