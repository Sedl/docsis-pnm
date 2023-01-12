package main

import (
	_ "github.com/lib/pq"
	"github.com/sedl/docsis-pnm/internal/cmd"
	"github.com/sedl/docsis-pnm/internal/types"
)

var cfg = &types.Config{}

func main() {
    cmd.CobraInit(cfg)

    cmd.CobraExecute()
}
