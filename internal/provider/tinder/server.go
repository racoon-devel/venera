package tinder

import (
	"encoding/json"
	"net/http"
)

type loginResponse struct {
	LoginToken string `json:"login_token"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	telParam, ok := r.URL.Query()["tel"]
	if !ok || len(telParam) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	auth := newTinderAuth(telParam[0])

	err := auth.RequestCode()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := loginResponse{LoginToken: auth.LoginToken}
	body, _ := json.Marshal(&response)
	w.Write(body)
}
