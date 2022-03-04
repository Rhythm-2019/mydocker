package cgroup

import (
	"bufio"
	"errors"
	log "github.com/sirupsen/logrus"
	"os"
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
