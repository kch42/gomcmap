// addchunk adds a chunk at 200, 200 that consists of sandstone.
package main

import (
	"flag"
	"fmt"
	"github.com/silvasur/gomcmap/mcmap"
	"os"
)

func main() {
	path := flag.String("path", "", "Path to region directory")
	flag.Parse()

	if *path == "" {
		flag.Usage()
		os.Exit(1)
	}

	region, err := mcmap.OpenRegion(*path, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open region: %s\n", err)
		os.Exit(1)
	}

	chunk, err := region.NewChunk(200, 200)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create a Chunk at 200,200: %s\n", err)
		os.Exit(1)
	}

	chunk.Iter(func(x, y, z int, blk *mcmap.Block) {
		blk.ID = mcmap.BlkSandstone
	})

	chunk.RecalcHeightMap()
	chunk.MarkModified()
	if err := chunk.MarkUnused(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not MarkUnused(): %s\n", err)
		os.Exit(1)
	}

	if err := region.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not save region: %s\n", err)
		os.Exit(1)
	}
}
