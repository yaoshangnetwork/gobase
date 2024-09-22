package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yaoshangnetwork/gobase/response/commerrs"
)

func Success(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

func Error(ctx *gin.Context, err error) {
	var e *commerrs.APIError
	if errors.As(err, &e) {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    e.Code,
			"message": e.Message,
		})
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    999999,
			"message": err.Error(),
		})
	}
}

func JSON(ctx *gin.Context, data any, err error) {
	if err != nil {
		Error(ctx, err)
		return
	}
	Success(ctx, data)
}
