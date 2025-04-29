package httpServ

import (
	"github.com/gorilla/mux"
)

func NewRouter(aggregatorHandler *AggregatorHandler) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/usprofile/profile-with-auth", aggregatorHandler.SignUpUserWithCreateProfile).Methods("GET")

	return router
}
