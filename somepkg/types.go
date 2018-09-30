
// Code generated by types_generator.go - DO NOT EDIT.

package somepkg

// Kind corresponds to the possible kinds of ground tiles in the
// game.  These enumerators are automatically generated from a text
// file that defines all of the possiblities.
type Kind int

const (
	_ TileType = iota
	Grass
	Dirt
	Bush
	Tree
	Water
	
)

type Property int

const (
	_ Property = (1 << iota) >> 1
	Burns
	Nowalk
	
)

var DefaultProperty = map[Kind]Property {
	Bush: Burns,
	Water: Nowalk,
	Grass: Burns,
	Dirt: 0,
	Tree: Burns | Nowalk,

}
