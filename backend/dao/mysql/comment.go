package mysql

import "bluebell/models"

func CreateComment(comment *models.Comment) error {
	rt := DB.Table("comment").Create(comment)
	err := rt.Error
	return err
}

func GetCommentList(postId uint64, commentList *[]models.Comment) error {
	rt := DB.Table("comment").Where("post_id = ?", postId).Scan(commentList)
	err := rt.Error
	return err
}
