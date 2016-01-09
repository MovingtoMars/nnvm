package platform

type Platform int

const (
	Invalid Platform = iota
	Linux
	MaxOSX
	Windows
)

func (v Platform) IsUnixLike() bool {
	return v == Linux || v == MaxOSX
}
