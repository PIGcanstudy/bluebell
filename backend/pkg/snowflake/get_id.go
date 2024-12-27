package snowflake

import (
	"fmt"

	"github.com/bwmarrin/snowflake"
)

func GetID() int64 {
	node, err := snowflake.NewNode(1)

	if err != nil {
		fmt.Println(err)
		return -1
	}

	id := node.Generate().Int64()

	return id
}
