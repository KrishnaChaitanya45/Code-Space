package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *App) Router() http.Handler {
	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/api/v1/", app.GetLoginForm)
	router.HandlerFunc(http.MethodGet, "/api/v1/auth/github", app.GithubLoginHandler)
	router.HandlerFunc(http.MethodGet, "/auth/callback", app.GithubCallBackHandler)
	router.HandlerFunc(http.MethodPost, "/execute", app.ExecuteCode)
	return router
}
