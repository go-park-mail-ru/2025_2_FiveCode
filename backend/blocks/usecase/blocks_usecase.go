package blocksUsecase

type BlocksRepository interface {
}

type BlocksUsecase struct {
	Repository BlocksRepository
}

func NewBlocksUsecase(Repository BlocksRepository) *BlocksUsecase {
	return &BlocksUsecase{
		Repository: Repository,
	}
}
