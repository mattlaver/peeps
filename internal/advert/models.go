package advert

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type ContactDetails struct {
	Name string		`bson:"name" json:"name"`
	Email string	`bson:"email" json:"email"`
	Phone string	`bson:"phone" json:"phone"`
}

// Advert is .
type Advert struct {
	ID           bson.ObjectId `bson:"_id" json:"id"`                      // Unique identifier
	Advertiser   string        `bson:"advertiser" json:"advertiser"`       // Display name of the product.
	Size     	 string        `bson:"size" json:"size"`                   // Size of advertisement.
	Editions     []string      `bson:"editions" json:"editions"`           // Editions that the advertisement is printed
	Year         string        `bson:"year" json:"year"`                  // Year
	State        []string      `bson:"state" json:"state"`				   // State
	Contact		 ContactDetails	`bson:"contact" json:"contact"`				// Contact
	DateCreated  time.Time     `bson:"date_created" json:"date_created"`   // When the product was added.
	DateModified time.Time     `bson:"date_modified" json:"date_modified"` // When the product record was lost modified.
}

// NewAdvert is what we require from clients when adding a Advert.
type NewAdvert struct {
	Advertiser    string 		`json:"advertiser" validate:"required"`
	Contact 	  ContactDetails `json:"contact" validate:"required"`
	Editions      []string    	`json:"editions" validate:"required"`
	Year          string    	`json:"year" validate:"required"`
	State         []string      `json:"state" validate:"required"`
}

// UpdateAdvert defines what information may be provided to modify an
// existing Advert. All fields are optional so clients can send just the
// fields they want changed. It uses pointer fields so we can differentiate
// between a field that was not provided and a field that was provided as
// explicitly blank. Normally we do not want to use pointers to basic types but
// we make exceptions around marshalling/unmarshalling.
type UpdateAdvert struct {
	Advertiser	*string 	`json:"name"`
	Contact     *ContactDetails `json:"contact"`
	Size     	*string   	`json:"size"`
	Editions 	*[]string    	`json:"editions"`
	Year     	*string		`json:"year"`
	State       *[]string     `json:"state"`
}
