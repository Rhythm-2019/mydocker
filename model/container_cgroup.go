package model

import (
	"errors"
	"github.com/rhythm_2019/mydocker/cfg"
	"github.com/rhythm_2019/mydocker/limiter"
	"github.com/rhythm_2019/mydocker/toolkit"
	"os"
	"path/filepath"
)

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
