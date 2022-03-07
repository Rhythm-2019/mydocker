package model

import (
    "errors"
    "fmt"
    "github.com/rhythm_2019/mydocker/cfg"
    "github.com/rhythm_2019/mydocker/limiter"
    "github.com/rhythm_2019/mydocker/toolkit"
    log "github.com/sirupsen/logrus"
    "os"
    "os/exec"
    "path/filepath"
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
    Image           *Image
    LogFd         *os.File
    LogFile       string
    CgroupManager *CgroupManager
    MergeDir      string
    WorkDir       string
    UpperDir      string
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

func (cm *Container) Start() {
    // os.Environ() 复制一份环境变量
    log.Infof("run a command %s", cm.Command)
    if err := syscall.Exec(cm.Command[0], cm.Command, os.Environ()); err != nil {
        log.Fatal("exec:", err)
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

    err = toolkit.MakeMountPoint(c.UpperDir, c.WorkDir, c.Image.LowerDir, c.WorkDir)
    if err != nil {
        log.Fatalf("run: failed to make mount point, detail is %v", err)
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
    if err = ApplyRecord(c, INSERT); err != nil {
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
    if err = ApplyRecord(ctn, DELETE); err != nil {
        return err
    }
    return nil
}

// https://xigang.github.io/2018/07/08/cgroups
type CgroupManager struct {
    ctn     *Container
    configs []*CgroupConfig
}

func NewCgroupManager(ctn *Container, configs []*CgroupConfig) *CgroupManager {
    manager := &CgroupManager{ctn: ctn, configs: configs}
    ctn.CgroupManager = manager
    return manager
}
func (m *CgroupManager) Apply(pid int) error {
    for _, config := range m.configs {
        // perpare path
        path, err := m.path(config)
        if err != nil {
            return err
        }

        err = config.Limiter.Set(path, config.Value)
        if err != nil {
            return err
        }
        err = config.Limiter.Apply(filepath.Dir(path), pid, limiter.O_APPLY_ADD)
        if err != nil {
            return err
        }
    }
    return nil
}
func (m *CgroupManager) Destory() error {
    // just remove dir
    for _, config := range m.configs {
        path, err := m.path(config)
        if err != nil {
            return err
        }
        err = os.Remove(filepath.Dir(path))
        if err != nil {
            return err
        }
    }
    return nil
}

func (m *CgroupManager) path(config *CgroupConfig) (string, error) {
    if len(m.ctn.Name) == 0 {
        return "", errors.New("name empty")
    }
    path := filepath.Join(cfg.CgroupRootPath(), config.Limiter.Subsystem(), m.ctn.Name, config.Limiter.Item())
    err := toolkit.CreateFile(path)
    if err != nil {
        return "", err
    }
    return path, nil
}

type CgroupConfig struct {
    Limiter limiter.SubsystemLimiter
    Value   string
}
