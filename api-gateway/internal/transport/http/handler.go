package httpServ

import (
	"encoding/json"
	"net/http"

	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/internal/entity"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/internal/service"
)

type AggregatorHandler struct {
	aggregatorService *service.AggregatorService
}

func NewAggregatorHandler(aggregatorService *service.AggregatorService) *AggregatorHandler {
	return &AggregatorHandler{aggregatorService: aggregatorService}
}

func (h *AggregatorHandler) SignUpUserWithCreateProfile(w http.ResponseWriter, r *http.Request) {
	var input entity.UserSignUpInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Вызываем SignUpUser с декодированным input
	data, err := h.aggregatorService.SignUpUser(r.Context(), input)
	if err != nil {
		http.Error(w, "failed to sign up user and create profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовок и возвращаем результат
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
