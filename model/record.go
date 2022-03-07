package model

import (
    "encoding/json"
    "errors"
    "github.com/rhythm_2019/mydocker/cfg"
    "io/ioutil"
    "reflect"
)

const (
    ImageType     = "Image"
    ContainerType = "Container"
)

const (
    INSERT = iota
    UPDATE
    DELETE
)
func ApplyRecord(v interface{}, operate int) error {
    var (
        err            error
        listInterfaces interface{}
        loc            string
        found           bool
        idx         int
    )
    switch reflect.TypeOf(v).Name() {
    case ImageType:
        listInterfaces, err = ListRecord(ImageType)
        loc = cfg.ImageInfoPath()
    case ContainerType:
        listInterfaces, err = ListRecord(ContainerType)
        loc = cfg.ContainerInfoPath()
    }
    if err != nil {
        return err
    }

    listInterfacesSlice := listInterfaces.([]interface{})
    models := listInterfaces.([]*BaseModel)
    targetModel := v.(*BaseModel)

    // find
    for i, model := range models {
        if model.Id == targetModel.Id {
            found = true
            idx = i
        }
    }

    switch operate {
    case INSERT:
        if ! found {
            listInterfacesSlice = append(listInterfacesSlice, v)
        }
    case UPDATE:
        if found {
            listInterfacesSlice[idx] = v
        }
    case DELETE:
        if found {
            if len(listInterfacesSlice) - 1 == idx {
                listInterfacesSlice = listInterfacesSlice[:idx]
            } else {
                listInterfacesSlice = append(listInterfacesSlice[:idx], listInterfacesSlice[idx + 1:]...)
            }
        }
    }
    newList, _ := json.Marshal(listInterfacesSlice)
    if err = ioutil.WriteFile(loc, newList, 0755); err != nil {
        return err
    }
    return nil
}

func ListRecord(t string) (rtn interface{}, err error)  {
    var (
        loc string
        buf []byte
    )
    switch t {
    case ContainerType:
        loc = cfg.ContainerInfoPath()
        rtn = []*Container{}
    case ImageType:
        loc = cfg.ImageInfoPath()
        rtn = []*Image{}
    }

    buf, err = ioutil.ReadFile(loc)
    if err != nil {
        return
    }
    err = json.Unmarshal(buf, &rtn)
    if err != nil {
        return
    }
    return
}
func GetRecord(key, t string) (interface{}, error) {
    list, err := ListRecord(t)
    if err != nil {
        return nil, err
    }
    models := list.([]*BaseModel)
    for _, model := range models {
        if model.Id == key || model.Name == key {
            return model, nil
        }
    }
    return nil, errors.New("record no found")
}

