package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
)

func CreateComment(comment *models.Comment) error {
	return mysql.CreateComment(comment)
}

func GetCommentList(postId uint64, commentList *[]models.Comment) error {
	return mysql.GetCommentList(postId, commentList)
}
