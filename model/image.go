package model

import (
	"encoding/json"
	"fmt"
	"github.com/rhythm_2019/mydocker/cfg"
	"github.com/rhythm_2019/mydocker/toolkit"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

type Image struct {
	BaseModel
	LowerDir string `json:"lower_dir"`
}

func FindImage(imageName string) (*Image, error) {
	image, err := GetImageRecordByName(imageName)
	if err != nil {
		return nil, fmt.Errorf("image no found, detail is %v", err)
	}
	return image, nil
}

func NewImage(imageUrl string) error {
	// 入参检查
	absPath, err := filepath.Abs(imageUrl)
	if err != nil {
		return fmt.Errorf("store: get image abs path failed, detail is %v", err)
	}
	if !toolkit.HasFile(absPath) {
		return fmt.Errorf("image file no found")
	}
	name, ext := toolkit.NameAndExtInPath(absPath)
	if ext != "tar.gz" {
		return fmt.Errorf("image file ext should be tar.gz")
	}

	image := &Image{
		BaseModel: BaseModel{
			Id:         toolkit.SnowflakeId(),
			Name:       name,
			CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	// 解压
	image.LowerDir = filepath.Join(cfg.StorePath(), image.Id, "lower")
	if err = toolkit.CreateDir(image.LowerDir); err != nil {
		return fmt.Errorf("store: create LowerLaye failed, detail is %v", err)
	}
	if err = toolkit.UnTarGz(absPath, image.LowerDir); err != nil {
		return fmt.Errorf("store: untar %s failed, detail is %v", absPath, err)
	}
	log.Infof("image: new image untar in %s", image.LowerDir)

	// 创建记录
	err = SaveImageRecord(image)
	if err != nil {
		return fmt.Errorf("image: insert record failed, detail is %v", err)
	}
	return nil
}

func (ctn *Image) Serialize() []byte {
	s, _ := json.Marshal(ctn)
	return s
}
func UnserializeImage(data []byte) (*Image, error) {
	var i Image
	err := json.Unmarshal(data, &i)
	return &i, err
}
