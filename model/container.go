package model

import (
	"encoding/json"
	"fmt"
	"github.com/rhythm_2019/mydocker/toolkit"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
)

type Container struct {
	BaseModel
	Pid           int
	Command       []string
	State         string
	Environment   map[string]string
	Port          string
	Tty           bool
	Image         *Image
	LogFd         *os.File
	LogFile       string
	CgroupManager *CgroupManager
	MergeDir      string
	WorkDir       string
	UpperDir      string
	Volumns       []*Volumn
}
type Volumn struct {
	HostDir   string // 宿主机路径
	VolumnDIr string // 容器路径
}

func (c *Container) Init() {
	// 自动补全前面的目录位置，如用户输入bash，需要替换为/bin/bash
	path, err := exec.LookPath(c.Command[0])
	if err != nil {
		log.Fatalf("init: %s command no found, detail is %v", c.Command[0], err)
	}
	c.Command[0] = path

	targetRoot, err := os.Getwd()
	if err != nil {
		log.Fatalf("init: get process work dir failed, detail is %v", err)
	}
	err = toolkit.PivotRoot(targetRoot)
	if err != nil {
		log.Fatalf("init: pivot root (%s) failed, detail is %v", targetRoot, err)
	}

	// mount proc
	if err := toolkit.MountProc(); err != nil {
		log.Fatalf("mount proc failed, detail is %v", err)
	}
	// mount dev
	if err := toolkit.MountDev(); err != nil {
		log.Fatalf("mount dev failed, detail is %v", err)
	}
}

func (cm *Container) Start() {
	// os.Environ() 复制一份环境变量
	log.Infof("run a command %s", cm.Command)
	if err := syscall.Exec(cm.Command[0], cm.Command, os.Environ()); err != nil {
		log.Fatalf("exec %s failed, defail is %v:", cm.Command, err)
	}
}

func (c *Container) Bootstrap() {
	// create a pipe to pass command
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		log.Fatalf("run: failed to open pipe, detail is %v", err)
	}
	// create a child process
	process := exec.Command("/proc/self/exe", "init")
	process.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	err = toolkit.MakeMountPoint(c.UpperDir, c.WorkDir, c.Image.LowerDir, c.MergeDir)
	if err != nil {
		log.Fatalf("run: failed to make mount point, detail is %v", err)
	}
	for _, volumn := range c.Volumns {
		if err := toolkit.MakeBindMount(volumn.HostDir, volumn.VolumnDIr); err != nil {
			log.Fatalf("run: failed to make volumn, err is %v", err)
		}
	}
	process.Dir = c.MergeDir

	process.ExtraFiles = []*os.File{readPipe}
	if c.Tty {
		process.Stdin = os.Stdin
		process.Stdout = os.Stdout
		process.Stderr = os.Stderr
	} else {
		// detach
		fd, err := os.OpenFile(c.LogFile, os.O_CREATE|os.O_RDWR, 0755)
		if err != nil {
			log.Errorf("open log file %s failed, detail is %v", c.LogFile, err)
			fd, err := os.Open("/dev/null")
			if err != nil {
				log.Errorf("open log file /dev/null failed, detail is %v", err)
			} else {
				process.Stdout = fd
			}
		} else {
			process.Stdout = fd
			c.LogFd = fd
		}
	}
	if err = SaveContainerRecord(c); err != nil {
		log.Fatalf("inset container record failed, detail is %v", err)
	}

	if err := process.Start(); err != nil {
		log.Fatal(err)
	}
	c.Pid = process.Process.Pid
	log.Infof("process %d is running", c.Pid)

	// send container id
	_, err = writePipe.WriteString(c.Id)
	if err != nil {
		_ = process.Process.Kill()
		log.Fatal(err)
	}
	writePipe.Close()

	// limit resource
	if err := c.CgroupManager.Apply(process.Process.Pid); err != nil {
		log.Fatal(err)
	}
	process.Wait()

	c.Destroy()

	log.Infof("process %d is exit", process.Process.Pid)
}

func (c *Container) Destroy() {
	if err := c.CgroupManager.Destory(); err != nil {
		log.Fatal(err)
	}
	for _, volumn := range c.Volumns {
		if err := toolkit.RemoveBindMount(volumn.VolumnDIr); err != nil {
			log.Fatal(err)
		}
	}
	if c.LogFd != nil {
		c.LogFd.Close()
	}
}

func (ctn *Container) Delete() error {
	// remove file
	err := syscall.Unmount(ctn.MergeDir, 0)
	if err != nil {
		return fmt.Errorf("container: umount merge dir failed, detail is %v", err)
	}
	var dirs = []string{ctn.MergeDir, ctn.WorkDir, ctn.UpperDir, ctn.LogFile}
	for _, dir := range dirs {
		err := os.RemoveAll(dir)
		if err != nil {
			log.Errorf("container: delete dir %s failed, detail is %v", dir, err)
		}
	}
	// remove record
	if err = RemoveContainerRecord(ctn); err != nil {
		return err
	}
	return nil
}
func (ctn *Container) Serialize() []byte {
	s, _ := json.Marshal(ctn)
	return s
}
func UnserializeContainer(data []byte) (*Container, error) {
	var c Container
	err := json.Unmarshal(data, &c)
	return &c, err
}
