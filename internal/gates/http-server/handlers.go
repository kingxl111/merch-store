package http_server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/kingxl111/merch-store/internal/shop"

	"github.com/go-faster/errors"
	env "github.com/kingxl111/merch-store/internal/environment"

	"github.com/kingxl111/merch-store/internal/users"

	merchstoreapi "github.com/kingxl111/merch-store/pkg/api/merch-store"
	"github.com/labstack/echo/v4"
)

const (
	internalServerError = "internal server error"
)

var _ merchstoreapi.ServerInterface = (*Handler)(nil)

type Handler struct {
	userService UserService
	shopService ShopService
}

func NewHandler(userService UserService, shopService ShopService) *Handler {
	return &Handler{
		userService: userService,
		shopService: shopService,
	}
}

func (h *Handler) PostApiAuth(echoCtx echo.Context) error {
	var req merchstoreapi.AuthRequest
	if err := echoCtx.Bind(&req); err != nil {
		errMsg := err.Error()
		return echoCtx.JSON(
			http.StatusBadRequest,
			merchstoreapi.ErrorResponse{Errors: &errMsg},
		)
	}
	ctx := echoCtx.Request().Context()
	// dto
	servReq := users.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	}
	resp, err := h.userService.Authenticate(ctx, &servReq)
	if err != nil {
		if errors.Is(err, users.ErrorWrongPassword) {
			errMsg := "wrong password"
			return echoCtx.JSON(
				http.StatusBadRequest,
				merchstoreapi.ErrorResponse{Errors: &errMsg},
			)
		}
		errMsg := internalServerError
		return echoCtx.JSON(
			http.StatusInternalServerError,
			merchstoreapi.ErrorResponse{Errors: &errMsg},
		)
	}

	return echoCtx.JSON(http.StatusOK, merchstoreapi.AuthResponse{Token: &resp.Token})
}

func (h *Handler) GetApiBuyItem(echoCtx echo.Context, item string) error {
	ctx := echoCtx.Request().Context()
	req := shop.InventoryItem{
		Type:     item,
		Quantity: 1,
	}

	err := h.shopService.BuyMerch(ctx, req)
	if err != nil {
		var errMsg string

		switch {
		case errors.Is(err, shop.ErrUserNotFound):
			errMsg = "user not found"
			return echoCtx.JSON(http.StatusNotFound, merchstoreapi.ErrorResponse{Errors: &errMsg})

		case errors.Is(err, shop.ErrItemNotFound):
			errMsg = "item not found"
			return echoCtx.JSON(http.StatusNotFound, merchstoreapi.ErrorResponse{Errors: &errMsg})

		case errors.Is(err, shop.ErrInsufficientFunds):
			errMsg = "not enough money"
			return echoCtx.JSON(http.StatusPaymentRequired, merchstoreapi.ErrorResponse{Errors: &errMsg})

		case errors.Is(err, shop.ErrBuildQuery),
			errors.Is(err, shop.ErrUpdateBalance),
			errors.Is(err, shop.ErrUpdateInventory),
			errors.Is(err, shop.ErrTransactionFailed):
			errMsg = internalServerError
			return echoCtx.JSON(http.StatusInternalServerError, merchstoreapi.ErrorResponse{Errors: &errMsg})

		default:
			slog.Error("Unexpected error in BuyMerch", slog.Any("error", err))
			errMsg = internalServerError
			return echoCtx.JSON(http.StatusInternalServerError, merchstoreapi.ErrorResponse{Errors: &errMsg})
		}
	}

	return echoCtx.JSON(http.StatusOK, "Предмет куплен")
}

/*
Список купленных мерчовых товаров
Сгруппированную информацию о перемещении монеток в его кошельке, включая:
Кто ему передавал монетки и в каком количестве
Кому сотрудник передавал монетки и в каком количестве
*/
func (h *Handler) GetApiInfo(ctx echo.Context) error {
	// TODO: call service layer
	info := merchstoreapi.InfoResponse{
		Coins: new(int),
	}
	*info.Coins = 100
	return ctx.JSON(http.StatusOK, info)
}

func (h *Handler) PostApiSendCoin(echoCtx echo.Context) error {
	var req merchstoreapi.SendCoinRequest
	if err := echoCtx.Bind(&req); err != nil {
		errMsg := err.Error()
		return echoCtx.JSON(
			http.StatusBadRequest,
			merchstoreapi.ErrorResponse{Errors: &errMsg},
		)
	}

	ctx := echoCtx.Request().Context()
	fromUser := ctx.Value(env.UsernameContextKey).(string)
	transfer := users.CoinTransfer{
		FromUser: fromUser,
		ToUser:   req.ToUser,
		Amount:   req.Amount,
	}
	err := h.userService.TransferCoins(ctx, &transfer)
	fmt.Println(err)
	if err != nil {
		if errors.Is(err, users.ErrorInsufFunds) {
			errMsg := "insufficient funds in the sender's balance"
			return echoCtx.JSON(
				http.StatusBadRequest,
				merchstoreapi.ErrorResponse{Errors: &errMsg},
			)
		}
		if errors.Is(err, users.ErrorInvalidAmount) {
			errMsg := "wrong amount format"
			return echoCtx.JSON(
				http.StatusBadRequest,
				merchstoreapi.ErrorResponse{Errors: &errMsg},
			)
		}
		errMsg := internalServerError
		return echoCtx.JSON(
			http.StatusInternalServerError,
			merchstoreapi.ErrorResponse{Errors: &errMsg},
		)
	}

	return echoCtx.JSON(http.StatusOK, "Монеты отправлены")
}
