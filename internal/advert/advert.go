package advert

import (
	"context"
	"fmt"
	"time"

	"github.com/mattlaver/peeps/internal/platform/db"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const advertsCollection = "adverts"

var (
	// ErrNotFound abstracts the mgo not found error.
	ErrNotFound = errors.New("Entity not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
)


// List retrieves a list of existing products from the database.
func List(ctx context.Context, dbConn *db.DB) ([]Advert, error) {
	ctx, span := trace.StartSpan(ctx, "internal.product.List")
	defer span.End()

	p := []Advert{}

	f := func(collection *mgo.Collection) error {
		return collection.Find(nil).All(&p)
	}
	if err := dbConn.Execute(ctx, advertsCollection, f); err != nil {
		return nil, errors.Wrap(err, "db.adverts.find()")
	}

	return p, nil
}

// Retrieve gets the specified product from the database.
func Retrieve(ctx context.Context, dbConn *db.DB, id string) (*Advert, error) {
	ctx, span := trace.StartSpan(ctx, "internal.product.Retrieve")
	defer span.End()

	if !bson.IsObjectIdHex(id) {
		return nil, ErrInvalidID
	}

	q := bson.M{"_id": bson.ObjectIdHex(id)}

	var p *Advert
	f := func(collection *mgo.Collection) error {
		return collection.Find(q).One(&p)
	}
	if err := dbConn.Execute(ctx, advertsCollection, f); err != nil {
		if err == mgo.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, fmt.Sprintf("db.adverts.find(%s)", db.Query(q)))
	}

	return p, nil
}

// Create inserts a new product into the database.
func Create(ctx context.Context, dbConn *db.DB, cp *NewAdvert, now time.Time) (*Advert, error) {
	ctx, span := trace.StartSpan(ctx, "internal.adverts.Create")
	defer span.End()

	// Mongo truncates times to milliseconds when storing. We and do the same
	// here so the value we return is consistent with what we store.
	now = now.Truncate(time.Millisecond)

	p := Advert{
		ID:           	bson.NewObjectId(),
		Advertiser:     cp.Advertiser,
		Editions:       cp.Editions,
		Year:     		cp.Year,
		State:    		cp.State,
		DateCreated:  	now,
		DateModified: 	now,
	}

	f := func(collection *mgo.Collection) error {
		return collection.Insert(&p)
	}
	if err := dbConn.Execute(ctx, advertsCollection, f); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("db.adverts.insert(%s)", db.Query(&p)))
	}

	return &p, nil
}

// Update replaces a product document in the database.
func Update(ctx context.Context, dbConn *db.DB, id string, upd UpdateAdvert, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.advert.Update")
	defer span.End()

	if !bson.IsObjectIdHex(id) {
		return ErrInvalidID
	}

	fields := make(bson.M)

	if upd.Advertiser != nil {
		fields["advertiser"] = *upd.Advertiser
	}
	if upd.Editions != nil {
		fields["editions"] = *upd.Editions
	}
	if upd.Year != nil {
		fields["year"] = *upd.Year
	}
	if upd.State != nil {
		fields["state"] = *upd.State
	}

	// If there's nothing to update we can quit early.
	if len(fields) == 0 {
		return nil
	}

	fields["date_modified"] = now

	m := bson.M{"$set": fields}
	q := bson.M{"_id": bson.ObjectIdHex(id)}

	f := func(collection *mgo.Collection) error {
		return collection.Update(q, m)
	}
	if err := dbConn.Execute(ctx, advertsCollection, f); err != nil {
		if err == mgo.ErrNotFound {
			return ErrNotFound
		}
		return errors.Wrap(err, fmt.Sprintf("db.adverts.update(%s, %s)", db.Query(q), db.Query(m)))
	}

	return nil
}

// Delete removes a product from the database.
func Delete(ctx context.Context, dbConn *db.DB, id string) error {
	ctx, span := trace.StartSpan(ctx, "internal.advert.Delete")
	defer span.End()

	if !bson.IsObjectIdHex(id) {
		return ErrInvalidID
	}

	q := bson.M{"_id": bson.ObjectIdHex(id)}

	f := func(collection *mgo.Collection) error {
		return collection.Remove(q)
	}
	if err := dbConn.Execute(ctx, advertsCollection, f); err != nil {
		if err == mgo.ErrNotFound {
			return ErrNotFound
		}
		return errors.Wrap(err, fmt.Sprintf("db.adverts.remove(%v)", q))
	}

	return nil
}