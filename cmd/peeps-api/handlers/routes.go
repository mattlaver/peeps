package handlers

import (
	"github.com/mattlaver/peeps/internal/mid"
	"github.com/mattlaver/peeps/internal/platform/auth"
	"github.com/mattlaver/peeps/internal/platform/db"
	"github.com/mattlaver/peeps/internal/platform/web"
	"log"
	"net/http"
	"os"
)

func API(shutdown chan os.Signal, log *log.Logger, masterDB *db.DB, authenticator *auth.Authenticator) http.Handler {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, log, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics())


	// Register health check endpoint. This route is not authenticated.
	 check := Check{
		MasterDB: masterDB,
	}
	//app.han
	app.Handle("GET", "/v1/health", check.Health)


	// Register user management and authentication endpoints.
	u := User{
		MasterDB:       masterDB,
		TokenGenerator: authenticator,
	}
	app.Handle("GET", "/v1/users", u.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("POST", "/v1/users", u.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("GET", "/v1/users/:id", u.Retrieve, mid.Authenticate(authenticator))
	app.Handle("PUT", "/v1/users/:id", u.Update, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("DELETE", "/v1/users/:id", u.Delete, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))



	// advertisers
	p := Advert{
		MasterDB: masterDB,
	}
	app.Handle("GET", "/v1/adverts", p.List, mid.Authenticate(authenticator))
	app.Handle("POST", "/v1/adverts", p.Create, mid.Authenticate(authenticator))
	app.Handle("GET", "/v1/adverts/:id", p.Retrieve, mid.Authenticate(authenticator))
	app.Handle("PUT", "/v1/adverts/:id", p.Update, mid.Authenticate(authenticator))
	app.Handle("DELETE", "/v1/adverts/:id", p.Delete, mid.Authenticate(authenticator))


	// This route is not authenticated
	app.Handle("GET", "/v1/users/token", u.Token)

	return app
}
