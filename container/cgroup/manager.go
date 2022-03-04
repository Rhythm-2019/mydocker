package cgroup

import (
	"errors"
	"github.com/rhythm_2019/mydocker/container/cgroup/limiter"
	"github.com/rhythm_2019/mydocker/toolkit"
	"os"
	"path/filepath"
)

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
