package http_server

import (
	"net/http"

	merch_store_api "github.com/kingxl111/merch-store/pkg/api/merch-store"
	"github.com/labstack/echo/v4"
)

var _ merch_store_api.ServerInterface = (*Handler)(nil)

type Handler struct{}

func (h *Handler) PostApiAuth(ctx echo.Context) error {
	var req merch_store_api.AuthRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, merch_store_api.ErrorResponse{Errors: &err.Error()})
	}

	// TODO: call jwt token creation
	token := "generated-jwt-token"
	return ctx.JSON(http.StatusOK, merch_store_api.AuthResponse{Token: &token})
}

func (h *Handler) GetApiBuyItem(ctx echo.Context, item string) error {
	// TODO: call service layer
	return ctx.JSON(http.StatusOK, "Предмет куплен")
}

/*
Список купленных мерчовых товаров
Сгруппированную информацию о перемещении монеток в его кошельке, включая:
Кто ему передавал монетки и в каком количестве
Кому сотрудник передавал монетки и в каком количестве
*/
func (h *Handler) GetApiInfo(ctx echo.Context) error {
	// TODO: call service layer
	info := merch_store_api.InfoResponse{
		Coins: new(int),
	}
	*info.Coins = 100
	return ctx.JSON(http.StatusOK, info)
}

func (h *Handler) PostApiSendCoin(ctx echo.Context) error {
	var req merch_store_api.SendCoinRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, merch_store_api.ErrorResponse{Errors: &err.Error()})
	}

	// TODO: call service layer
	return ctx.JSON(http.StatusOK, "Монеты отправлены")
}

//
//func main() {
//	e := echo.New()
//	handler := &Handler{}
//	merch_store_api.RegisterHandlers(e, handler)
//	e.Logger.Fatal(e.Start(":8080"))
//}
