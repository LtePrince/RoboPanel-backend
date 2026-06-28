package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ServiceFunc[Req, Resp any] func(context.Context, *Req) (*Resp, error)

func handleQuery[Req, Resp any](fn ServiceFunc[Req, Resp]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, Response{Code: "BAD_REQUEST", Message: err.Error()})
			return
		}
		resp, err := fn(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{Code: "INTERNAL_ERROR", Message: err.Error()})
			return
		}
		c.JSON(http.StatusOK, Response{Code: "OK", Message: "success", Data: resp})
	}
}

func handleJSON[Req, Resp any](fn ServiceFunc[Req, Resp]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, Response{Code: "BAD_REQUEST", Message: err.Error()})
			return
		}
		resp, err := fn(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{Code: "INTERNAL_ERROR", Message: err.Error()})
			return
		}
		c.JSON(http.StatusOK, Response{Code: "OK", Message: "success", Data: resp})
	}
}
