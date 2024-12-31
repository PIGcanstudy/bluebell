package controller

import (
	"github.com/gin-gonic/gin"
)

type ResponseData struct {
	Code MyCode      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data, omitempty"` // 当此字段为空不显示
}

func ResponseError(ctx *gin.Context, c MyCode) {
	rd := ResponseData{
		Code: c,
		Msg:  c.Msg(),
		Data: nil,
	}

	ctx.JSON(200, rd)
}

func ResponseErrorWithMsg(ctx *gin.Context, code MyCode, data interface{}) {
	rd := &ResponseData{
		Code: code,
		Msg:  code.Msg(),
		Data: data,
	}
	ctx.JSON(200, rd)
}

func ResponseSuccessed(ctx *gin.Context, data interface{}) {
	rd := &ResponseData{
		Code: ResponseSuccess,
		Msg:  ResponseSuccess.Msg(),
		Data: data,
	}
	ctx.JSON(200, rd)
}
