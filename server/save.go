package server

// PoolSaver provides a way to change the storage backend.
type PoolSaver interface {
	// SavePool takes a PoolInfo structure and saves it somehow.
	// Returns an error if there was an error saving.
	// While for any given `name`, at most one request will be made
	// at a time, there might be more than one simultanious request
	// with different values for `name`.
	SavePool(name string, info PoolInfo) error

	// LoadAllPools returns a list of the saved pools, or an error
	// it all the pools couldn't be read for some reason.
	LoadAllPools() ([]PoolInfo, error)
}
