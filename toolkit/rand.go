package toolkit

import "github.com/google/uuid"

func RandUUID() string {
    return uuid.New().String()
}
