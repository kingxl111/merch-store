package http

import (
	"context"
	shopServ "github.com/kingxl111/merch-store/internal/shop"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	infoService *shopServ.InfoService
}

func NewHandler(infoService *service.InfoService) *Handler {
	return &Handler{infoService: infoService}
}

func (h *Handler) GetApiInfo(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user")
	}

	info, err := h.infoService.GetInfo(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, toResponse(info))
}

func toResponse(info *repository.UserInfo) InfoResponse {
	// Конвертация структуры репозитория в API response
}
