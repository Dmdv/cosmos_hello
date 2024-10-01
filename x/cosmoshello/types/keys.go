package types

const (
	// ModuleName defines the module name
	ModuleName = "cosmoshello"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_cosmoshello"
)

var (
	ParamsKey = []byte("p_cosmoshello")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
