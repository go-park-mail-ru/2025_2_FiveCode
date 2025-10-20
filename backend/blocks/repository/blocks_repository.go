package blocksRepository

import "backend/store"

type BlocksRepository struct {
	Store *store.Store
}

func NewBlocksRepository(store *store.Store) *BlocksRepository {
	return &BlocksRepository{
		Store: store,
	}
}
