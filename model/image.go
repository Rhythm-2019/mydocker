package model

import (
    "fmt"
    "github.com/rhythm_2019/mydocker/cfg"
    "github.com/rhythm_2019/mydocker/toolkit"
    "path/filepath"
    "time"
)

type Image struct {
    BaseModel
    LowerDir   string  `json:"lower_dir"`
}

func FindImage(imageName string) (*Image, error) {
    image, err := GetRecord(imageName, ImageType)
    if err != nil {
        return nil, fmt.Errorf("image no found, detail is %v", err)
    }
    return image.(*Image), nil
}

func NewImage(imageUrl string) error {
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
            Id: toolkit.RandUUID(),
            Name: name,
            CreateTime: time.Now().Format("2006-01-02 15:04:05"),
        },
    }

    image.LowerDir = filepath.Join(cfg.StorePath(), image.Id, "lower")
    if err = toolkit.CreateDir(image.LowerDir); err != nil {
        return fmt.Errorf("store: create LowerLaye failed, detail is %v", err)
    }
    err = toolkit.UnTarGz(absPath, image.LowerDir)
    if err != nil {
        return fmt.Errorf("store: untar %s failed, detail is %v", absPath, err)
    }

    // make a record
    err = ApplyRecord(image, INSERT)
    if err != nil {
        return fmt.Errorf("image: insert record failed, detail is %v", err)
    }
    return nil
}


