package snowflake

import (
    "sync"

    "github.com/bwmarrin/snowflake"
)

var (
    once sync.Once
    node *snowflake.Node
)

func initNode() {
    n, err := snowflake.NewNode(1)
    if err != nil {
        panic(err)
    }
    node = n
}

func NextID() int64 {
    once.Do(initNode)
    return node.Generate().Int64()
}
