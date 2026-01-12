package utils

import "github.com/bwmarrin/snowflake"

func CreateID() string {
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}

	return node.Generate().String()
}
