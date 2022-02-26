package limiter

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const MEM_SUBSYSTEM = "memory"
const CPU_SUBSYSTEM = "cpu"

const (
	MEMORY_USAGE_IN_BYTES = "memory.limit_in_bytes"
	CPU_QUOTE             = "cpu.cfs_quota_us"
	CPU_PERIOD            = "cpu.cpu.cfs_period_us"
)

var CgroupRootPath string

type SubsystemLimiter interface {
	Item() string
	Subsystem() string
	Set(path, value string) error
	Apply(path string, pid int, operate int) error
}

type BaseLimiter struct {
	SubsystemLimiter
}

func NewLimiter(name string) *BaseLimiter {
	var limiter SubsystemLimiter
	switch name {
	case MEMORY_USAGE_IN_BYTES:
		limiter = &MemoryLimitInBytesLimiter{}
	case CPU_QUOTE:
		limiter = &CpuCfsQuotaLimiter{}
	case CPU_PERIOD:
		limiter = &CpuCfsPeriodLimiter{}
	}
	return &BaseLimiter{
		SubsystemLimiter: limiter,
	}
}

func (b *BaseLimiter) Set(path, limit string) error {
	return os.WriteFile(path, []byte(limit), 0644)
}

// operaate 1 mean opne, 0 mean remove
const (
	O_APPLY_ADD = iota
	O_APPLY_REMOVE
)

func (b *BaseLimiter) Apply(path string, pid int, operaate int) error {
	fd, err := os.OpenFile(filepath.Join(path, "tasks"), os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fd.Close()

	stat, err := fd.Stat()
	if err != nil {
		return err
	}
	var buf = make([]byte, stat.Size())
	idx, err := fd.Read(buf)
	if err != nil && err != io.EOF {
		return err
	}
	exist := strings.Contains(string(buf[:idx]), strconv.Itoa(pid))

	if operaate == O_APPLY_ADD && !exist {
		log.Infof("add a process %d in tasks %s", pid, path)
		_, err := fd.Seek(0, 2)
		if err != nil {
			return err
		}
		_, err = fd.Write([]byte(strconv.Itoa(pid)))
	} else if operaate == O_APPLY_REMOVE && exist {
		log.Infof("remove a process %d in tasks %s", pid, path)
		_, err := fd.Seek(0, 0)
		if err != nil {
			return err
		}
		_, err = fd.Write([]byte(strings.Replace(string(buf[:idx]), strconv.Itoa(pid), "", 1)))
	} else {
		err = errors.New("invail args")
	}
	return err
}
