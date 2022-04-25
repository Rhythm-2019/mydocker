package cfg

import (
	"encoding/json"
	"github.com/rhythm_2019/mydocker/toolkit"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sync"
)

const ConfigLocation = "/var/local/mydocker/config.json"

type config struct {
	AppName      string `json:"app_name" default:"mydocker"`
	BaseDir      string `json:"base_dir" default:"/var/local/mydocker"`
	StoreDirName string `json:"store_dir_name" default:"store"`
	CgroupRoot   string `json:"-"`
}

// --
var appConfig = struct {
	M      sync.Mutex
	config *config
}{}

func Init() {
	if !toolkit.HasFile(ConfigLocation) {
		log.Fatalf("cfg: config file %s no found", ConfigLocation)
	}
	buf, err := ioutil.ReadFile(ConfigLocation)
	if err != nil {
		log.Fatalf("cfg: read config file %s failed ", ConfigLocation)
	}

	appConfig.M.Lock()
	defer appConfig.M.Unlock()

	err = json.Unmarshal(buf, &appConfig.config)
	if err != nil {
		log.Fatalf("cfg: deserialize config file %s failed ", ConfigLocation)
	}

	// set default value
	t := reflect.TypeOf(appConfig.config)
	v := reflect.ValueOf(appConfig.config)

	for k := 0; k < t.Elem().NumField(); k++ {
		defaultValue := t.Elem().Field(k).Tag.Get("default")
		if len(defaultValue) != 0 && v.Elem().Field(k).Type().Kind() == reflect.String {
			v.Elem().Field(k).SetString(defaultValue)
		}
	}

	appConfig.config.CgroupRoot, err = toolkit.GetCgroupRoot()
	if err != nil {
		log.Fatalf("cfg: get cgroup root path failed, detail is %v ", err)
	}
}

func Config() *config {
	appConfig.M.Lock()
	defer appConfig.M.Unlock()

	return appConfig.config
}

func StorePath() string {
	return filepath.Join(Config().BaseDir, Config().StoreDirName)
}
func ImageInfoPath() string {
	return filepath.Join(StorePath(), "images.db")
}
func CgroupRootPath() string {
	return Config().CgroupRoot
}

func DBFilePath() string {
	return filepath.Join(StorePath(), "store.db")
}
