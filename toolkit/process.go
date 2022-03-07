package toolkit

import (
    "bufio"
    "errors"
    "os"
    "strings"
)

func GetCgroupRoot() (string, error) {
    focus := "memory"

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
            return items[4][:strings.Index(items[4], focus)-1], nil
        }
    }
    if err := scanner.Err(); err != nil {
        return "", nil
    }

    return "", errors.New("no found")
}

