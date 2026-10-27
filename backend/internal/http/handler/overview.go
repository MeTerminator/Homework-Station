package handler

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) Overview(c *gin.Context) { c.JSON(200, h.OverviewService.Get()) }
