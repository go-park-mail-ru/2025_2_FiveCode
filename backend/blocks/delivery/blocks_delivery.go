package blocksDelivery

type BlocksUsecase interface {
}

type BlocksDelivery struct {
	Usecase BlocksUsecase
}

func NewBlocksDelivery(usecase BlocksUsecase) *BlocksDelivery {
	return &BlocksDelivery{
		Usecase: usecase,
	}
}
