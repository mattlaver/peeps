package handlers

import (
	"context"
	"net/http"

	"github.com/mattlaver/peeps/internal/advert"
	"github.com/mattlaver/peeps/internal/platform/db"
	"github.com/mattlaver/peeps/internal/platform/web"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// Advert represents the Advert API method handler set.
type Advert struct {
	MasterDB *db.DB

	// ADD OTHER STATE LIKE THE LOGGER IF NEEDED.
}


// List returns all the existing Adverts in the system.
func (p *Advert) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Advert.List")
	defer span.End()

	dbConn := p.MasterDB.Copy()
	defer dbConn.Close()

	Adverts, err := advert.List(ctx, dbConn)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, Adverts, http.StatusOK)
}

// Retrieve returns the specified Advert from the system.
func (p *Advert) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Advert.Retrieve")
	defer span.End()

	dbConn := p.MasterDB.Copy()
	defer dbConn.Close()

	prod, err := advert.Retrieve(ctx, dbConn, params["id"])
	if err != nil {
		switch err {
		case advert.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case advert.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "ID: %s", params["id"])
		}
	}

	return web.Respond(ctx, w, prod, http.StatusOK)
}

// Create inserts a new Advert into the system.
func (p *Advert) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Advert.Create")
	defer span.End()

	dbConn := p.MasterDB.Copy()
	defer dbConn.Close()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	var np advert.NewAdvert
	if err := web.Decode(r, &np); err != nil {
		return errors.Wrap(err, "")
	}

	nUsr, err := advert.Create(ctx, dbConn, &np, v.Now)
	if err != nil {
		return errors.Wrapf(err, "Advert: %+v", &np)
	}

	return web.Respond(ctx, w, nUsr, http.StatusCreated)
}

// Update updates the specified Advert in the system.
func (p *Advert) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Advert.Update")
	defer span.End()

	dbConn := p.MasterDB.Copy()
	defer dbConn.Close()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	var up advert.UpdateAdvert
	if err := web.Decode(r, &up); err != nil {
		return errors.Wrap(err, "")
	}

	err := advert.Update(ctx, dbConn, params["id"], up, v.Now)
	if err != nil {
		switch err {
		case advert.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case advert.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "ID: %s Update: %+v", params["id"], up)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Delete removes the specified Advert from the system.
func (p *Advert) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Advert.Delete")
	defer span.End()

	dbConn := p.MasterDB.Copy()
	defer dbConn.Close()

	err := advert.Delete(ctx, dbConn, params["id"])
	if err != nil {
		switch err {
		case advert.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case advert.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "Id: %s", params["id"])
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
