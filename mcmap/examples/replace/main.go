package main

import (
	"flag"
	"fmt"
	"github.com/kch42/gomcmap/mcmap"
	"os"
)

func main() {
	path := flag.String("path", "", "Path to region directory")
	flag.Parse()

	if *path == "" {
		flag.Usage()
		os.Exit(1)
	}

	region, err := mcmap.OpenRegion(*path, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open region: %s\n", err)
		os.Exit(1)
	}

chunkLoop:
	for chunkPos := range region.AllChunks() {
		cx, cz := chunkPos.X, chunkPos.Z
		chunk, err := region.Chunk(cx, cz)
		switch err {
		case nil:
		case mcmap.NotAvailable:
			continue chunkLoop
		default:
			fmt.Fprintf(os.Stderr, "Error while getting chunk (%d, %d): %s\n", cx, cz, err)
			os.Exit(1)
		}

		modified := false
		for y := 0; y < 256; y++ {
			for x := 0; x < 16; x++ {
				for z := 0; z < 16; z++ {
					blk := chunk.Block(x, y, z)
					if blk.ID == mcmap.BlkBlockOfIron {
						blk.ID = mcmap.BlkBlockOfDiamond
						modified = true
					}
				}
			}
		}

		if modified {
			fmt.Printf("Modified chunk %d, %d.\n", cx, cz)
			chunk.MarkModified()
		}

		if err := region.UnloadChunk(cx, cz); err != nil {
			fmt.Fprintf(os.Stderr, "Error while unloading chunk %d, %d: %s\n", cx, cz, err)
			os.Exit(1)
		}
	}

	if err := region.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error while saving: %s\n", err)
		os.Exit(1)
	}
}
