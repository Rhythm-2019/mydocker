package model

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rhythm_2019/mydocker/cfg"
	"github.com/rhythm_2019/mydocker/toolkit"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"strings"
)

var db *buntdb.DB

const (
//CONATINER_ID_IDX = "container_id_idx"
//IMAGE_ID_IDX = "image_id_idx"
)

func Init() {
	var err error
	var first bool
	if !toolkit.HasFile(cfg.DBFilePath()) {
		// create file
		err := toolkit.CreateFile(cfg.DBFilePath())
		if err != nil {
			log.Fatalf("open db failed, detail is %v", err)
		}
		first = true
	}

	db, err = buntdb.Open(cfg.DBFilePath())
	if err != nil {
		log.Fatalf("open db failed, detail is %v", err)
	}

	if first {
		// TODO 该库的索引是构建在内存中的，当前场景下不太适合使用。后期改成 C/S 架构时可以用上
		//handleErr := func (err error) {
		//    if err != nil {
		//        log.Fatalf("create index failed, detail is %v", err)
		//    }
		//}
		//err := db.CreateIndex(CONATINER_ID_IDX, "container:*", buntdb.IndexInt)
		//handleErr(err)
		//err = db.CreateIndex(IMAGE_ID_IDX, "image:*", buntdb.IndexInt)
		//handleErr(err)
	}
}

func ReleaseDB() {
	defer db.Close()
}
func SaveContainerRecord(c *Container) error {
	return db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(fmt.Sprintf("container:%s", c.Id), string(c.Serialize()), nil)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(fmt.Sprintf("container_name_id:%s", c.Name), c.Id, nil)
		return err
	})
}

func RemoveContainerRecord(c *Container) error {
	return db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(fmt.Sprintf("container:%s", c.Id))
		if err != nil {
			return err
		}
		_, err = tx.Delete(fmt.Sprintf("container_name_id:%s", c.Name))
		return err
	})
}
func ListContainerRecord() ([]*Container, error) {
	var rtn []*Container
	err := db.View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("", func(key, value string) bool {
			data, err := UnserializeContainer([]byte(value))
			if err != nil {
				log.Errorf("iterating container info failed. check file state")
				return false
			}
			rtn = append(rtn, data)
			return true
		})
		return err
	})
	return rtn, err
}
func GetContainerRecordByName(name string) (*Container, error) {
	var rtn *Container
	err := db.View(func(tx *buntdb.Tx) error {
		id, err := tx.Get(fmt.Sprintf("container_name_id:%s", name))
		if err != nil {
			return err
		}
		rtn, err = GetContainerRecord(id)
		if err != nil {
			return err
		}
		return nil
	})
	return rtn, err
}
func GetContainerRecord(id string) (*Container, error) {
	var rtn *Container
	err := db.View(func(tx *buntdb.Tx) error {
		value, err := tx.Get(fmt.Sprintf("container:%s", id))
		if err != nil {
			return err
		}
		rtn, err = UnserializeContainer([]byte(value))
		if err != nil {
			return err
		}
		return nil
	})
	return rtn, err
}

func SaveImageRecord(i *Image) error {
	return db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(fmt.Sprintf("image:%s", i.Id), string(i.Serialize()), nil)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(fmt.Sprintf("image_name_id:%s", i.Name), i.Id, nil)
		return err
	})
}

func RemoveImageRecord(c *Container) error {
	return db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(fmt.Sprintf("image:%s", c.Id))
		if err != nil {
			return err
		}
		_, err = tx.Delete(fmt.Sprintf("image_name_id:%s", c.Name))
		return err
	})
}
func ListImageRecord() ([]*Image, error) {
	var rtn []*Image
	err := db.View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("", func(key, value string) bool {
			if !strings.HasPrefix(key, "image:") {
				return true
			}
			data, err := UnserializeImage([]byte(value))
			if err != nil {
				log.Errorf("iterating image info failed. check file state, detual is %v", err)
				return false
			}
			rtn = append(rtn, data)
			return true
		})
		return err
	})
	return rtn, err
}

func GetImageRecord(id string) (*Image, error) {
	var rtn *Image
	err := db.View(func(tx *buntdb.Tx) error {
		value, err := tx.Get(fmt.Sprintf("image:%s", id))
		if err != nil {
			return err
		}
		rtn, err = UnserializeImage([]byte(value))
		if err != nil {
			return err
		}
		return nil
	})
	return rtn, err
}
func GetImageRecordByName(name string) (*Image, error) {
	var rtn *Image
	err := db.View(func(tx *buntdb.Tx) error {
		id, err := tx.Get(fmt.Sprintf("image_name_id:%s", name))
		if err != nil {
			return err
		}
		rtn, err = GetImageRecord(id)
		if err != nil {
			return err
		}
		return nil
	})
	return rtn, err
}
