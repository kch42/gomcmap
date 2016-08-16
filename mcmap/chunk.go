package mcmap

import (
	"errors"
	"github.com/silvasur/gonbt/nbt"
	"time"
)

func calcBlockOffset(x, y, z int) int {
	if (x < 0) || (y < 0) || (z < 0) || (x >= ChunkSizeXZ) || (y >= ChunkSizeY) || (z >= ChunkSizeXZ) {
		return -1
	}

	return x | (z << 4) | (y << 8)
}

func offsetToPos(off int) (x, y, z int) {
	x = off & 0xf
	z = (off >> 4) & 0xf
	y = (off >> 8) & 0xff
	return
}

// BlockToChunk calculates the chunk (cx, cz) and the block position in this chunk(rbx, rbz) of a block position given global coordinates.
func BlockToChunk(bx, bz int) (cx, cz, rbx, rbz int) {
	cx = bx >> 4
	cz = bz >> 4
	rbx = ((bx % ChunkSizeXZ) + ChunkSizeXZ) % ChunkSizeXZ
	rbz = ((bz % ChunkSizeXZ) + ChunkSizeXZ) % ChunkSizeXZ
	return
}

// ChunkToBlock calculates the global position of a block, given the chunk position (cx, cz) and the plock position in that chunk (rbx, rbz).
func ChunkToBlock(cx, cz, rbx, rbz int) (bx, bz int) {
	bx = cx*ChunkSizeXZ + rbx
	bz = cz*ChunkSizeXZ + rbz
	return
}

// Chunk represents a 16*16*256 Chunk of the region.
type Chunk struct {
	Entities []nbt.TagCompound

	x, z int32

	lastUpdate    int64
	populated     bool
	inhabitedTime int64
	ts            time.Time

	heightMap []int32 // Ordered ZX

	modified bool
	blocks   []Block // Ordered YZX
	biomes   []Biome // Ordered ZX

	deleted bool

	reg *Region
}

func newChunk(reg *Region, x, z int) *Chunk {
	biomes := make([]Biome, ChunkRectXZ)
	for i := range biomes {
		biomes[i] = BioUncalculated
	}

	heightMap := make([]int32, ChunkRectXZ)
	for i := range heightMap {
		heightMap[i] = ChunkSizeY - 1
	}

	return &Chunk{
		x:         int32(x),
		z:         int32(z),
		ts:        time.Now(),
		blocks:    make([]Block, ChunkSize),
		biomes:    biomes,
		heightMap: heightMap,
		reg:       reg,
	}
}

// MarkModified needs to be called, if some data of the chunk was modified.
func (c *Chunk) MarkModified() { c.modified = true }

// Coords returns the Chunk's coordinates.
func (c *Chunk) Coords() (X, Z int32) { return c.x, c.z }

// Block gives you a reference to the Block located at x, y, z. If you modify the block data, you need to call the MarkModified() function of the chunk.
//
// x and z must be in [0, 15], y in [0, 255]. Otherwise a nil pointer is returned.
func (c *Chunk) Block(x, y, z int) *Block {
	off := calcBlockOffset(x, y, z)
	if off < 0 {
		return nil
	}

	return &(c.blocks[off])
}

// Height returns the height at x, z.
//
// x and z must be in [0, 15]. Height will panic, if this is violated!
func (c *Chunk) Height(x, z int) int {
	if (x < 0) || (x >= ChunkSizeXZ) || (z < 0) || (z >= ChunkSizeXZ) {
		panic(errors.New("x or z parameter was out of range"))
	}

	return int(c.heightMap[z*ChunkSizeXZ+x])
}

// Iter iterates ofer all blocks of this chunk and calls the function fx with the coords (x,y,z) and a pointer to the block.
func (c *Chunk) Iter(fx func(int, int, int, *Block)) {
	for x := 0; x < ChunkSizeXZ; x++ {
		for y := 0; y < ChunkSizeY; y++ {
			for z := 0; z < ChunkSizeXZ; z++ {
				fx(x, y, z, &(c.blocks[calcBlockOffset(x, y, z)]))
			}
		}
	}
}

// Biome gets the Biome at x,z.
func (c *Chunk) Biome(x, z int) Biome { return c.biomes[z*ChunkSizeXZ+x] }

// SetBiome sets the biome at x,z.
func (c *Chunk) SetBiome(x, z int, bio Biome) { c.biomes[z*ChunkSizeXZ+x] = bio }

// MarkUnused marks the chunk as unused. If all chunks of a superchunk are marked as unused, the underlying superchunk will be unloaded and saved (if needed).
//
// You must not use the chunk any longer, after you called this function.
//
// If the chunk was modified, call MarkModified BEFORE.
func (c *Chunk) MarkUnused() error { return c.reg.unloadChunk(int(c.x), int(c.z)) }

// MarkDeleted marks this chunk as deleted. After marking it as unused, it will be deleted and can no longer be used.
func (c *Chunk) MarkDeleted() { c.deleted = true }

// RecalcHeightMap recalculates the internal height map.
//
// You should use this function before marking the chunk as unused, if you modified the chunk
// (unless you know, your changes wouldn't affect the height map).
func (c *Chunk) RecalcHeightMap() {
	i := 0
	for z := 0; z < ChunkSizeXZ; z++ {
		for x := 0; x < ChunkSizeXZ; x++ {
			for y := ChunkSizeY - 1; y >= 0; y-- {
				blkid := c.blocks[calcBlockOffset(x, y, z)].ID
				if (blkid != BlkAir) && (blkid != BlkGlass) && (blkid != BlkGlassPane) {
					c.heightMap[i] = int32(y)
					break
				}
			}
			i++
		}
	}
}
