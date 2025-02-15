package http_server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-faster/errors"
	"github.com/kingxl111/merch-store/internal/shop"
	"github.com/kingxl111/merch-store/internal/users"
	merchstoreapi "github.com/kingxl111/merch-store/pkg/api/merch-store"
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

func (h *Handler) PostApiAuth(w http.ResponseWriter, r *http.Request) {
	var req merchstoreapi.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx := r.Context()
	resp, err := h.userService.Authenticate(ctx, &users.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, users.ErrorWrongPassword) {
			h.respondWithError(w, http.StatusBadRequest, "wrong password")
			return
		}
		h.respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	h.respondWithJSON(w, http.StatusOK, merchstoreapi.AuthResponse{Token: &resp.Token})
}

func (h *Handler) GetApiBuyItem(w http.ResponseWriter, r *http.Request, item string) {
	ctx := r.Context()
	err := h.shopService.BuyMerch(ctx, shop.InventoryItem{
		Type:     item,
		Quantity: 1,
	})
	if err != nil {
		var status int
		var message string
		switch {
		case errors.Is(err, shop.ErrUserNotFound):
			status, message = http.StatusNotFound, "user not found"
		case errors.Is(err, shop.ErrItemNotFound):
			status, message = http.StatusNotFound, "item not found"
		case errors.Is(err, shop.ErrInsufficientFunds):
			status, message = http.StatusPaymentRequired, "not enough money"
		default:
			slog.Error("Unexpected error in BuyMerch", slog.Any("error", err))
			status, message = http.StatusInternalServerError, "internal server error"
		}
		h.respondWithError(w, status, message)
		return
	}
	h.respondWithJSON(w, http.StatusOK, "Item purchased")
}

func (h *Handler) GetApiInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username, ok := ctx.Value("username").(string)
	if !ok {
		h.respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userInfo, err := h.userService.GetUserInfo(ctx, username)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, userInfo)
}

func (h *Handler) PostApiSendCoin(w http.ResponseWriter, r *http.Request) {
	var req merchstoreapi.SendCoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx := r.Context()
	fromUser, ok := ctx.Value("username").(string)
	if !ok {
		h.respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	err := h.userService.TransferCoins(ctx, &users.CoinTransfer{
		FromUser: fromUser,
		ToUser:   req.ToUser,
		Amount:   req.Amount,
	})
	if err != nil {
		var status int
		var message string
		if errors.Is(err, users.ErrorInsufFunds) {
			status, message = http.StatusBadRequest, "insufficient funds"
		} else if errors.Is(err, users.ErrorInvalidAmount) {
			status, message = http.StatusBadRequest, "wrong amount format"
		} else {
			status, message = http.StatusInternalServerError, "internal server error"
		}
		h.respondWithError(w, status, message)
		return
	}

	h.respondWithJSON(w, http.StatusOK, "Coins sent")
}

func (h *Handler) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(merchstoreapi.ErrorResponse{Errors: &message})
}

func (h *Handler) respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}
