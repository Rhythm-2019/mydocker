package toolkit

import (
	"github.com/bwmarrin/snowflake"
	"github.com/google/uuid"
	"log"
)

var snowflakeNode *snowflake.Node

func init() {
	node, err := snowflake.NewNode(1)
	if err != nil {
		log.Panicf("create snowflake node failed, err is %v", err)
	}
	snowflakeNode = node
}
func RandUUID() string {
	return uuid.New().String()
}
func SnowflakeId() string {
	return snowflakeNode.Generate().String()
}
