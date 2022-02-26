package cgroup

import (
	"bufio"
	"errors"
	"github.com/rhythm_2019/mydocker/container/cgroup/limiter"
	"github.com/rhythm_2019/mydocker/toolkit"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

var CgroupRootPath string

func init() {
	// Get cgroup root
	root, err := getCgroupRoot("memory")
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("host cgroup root in %s", root)
	CgroupRootPath = root
}

func getCgroupRoot(focus string) (string, error) {
	if len(CgroupRootPath) != 0 {
		return CgroupRootPath, nil
	}
	fd, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		items := strings.Split(scanner.Text(), " ")
		if len(items) < 4 {
			return "", errors.New("bad format in mountinfo")
		}
		if strings.Contains(items[4], focus) {
			CgroupRootPath = items[4][:strings.Index(items[4], focus)-1]
			return CgroupRootPath, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", nil
	}

	return "", errors.New("no found")
}

// https://xigang.github.io/2018/07/08/cgroups
type CgroupManager struct {
	configs []*CgroupConfig
	Name    string
}

func NewCgroupManager(name string, configs []*CgroupConfig) *CgroupManager {
	return &CgroupManager{Name: name, configs: configs}
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
	for _, cfg := range m.configs {
		path, err := m.path(cfg)
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

func (m *CgroupManager) path(cfg *CgroupConfig) (string, error) {
	if len(m.Name) == 0 {
		return "", errors.New("name empty")
	}
	path := filepath.Join(CgroupRootPath, cfg.Limiter.Subsystem(), m.Name, cfg.Limiter.Item())
	err := toolkit.SureFile(path)
	if err != nil {
		return "", err
	}
	return path, nil
}

type CgroupConfig struct {
	Limiter limiter.SubsystemLimiter
	Value   string
}
