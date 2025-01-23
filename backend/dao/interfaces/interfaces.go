package interfaces

type PostService interface {
	InsertPostVoteRecord(postId, userId uint64, score int) error
}
