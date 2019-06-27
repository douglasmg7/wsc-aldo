package main

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/douglasmg7/money"
)

type Product struct {
	Code                 string      `db:"code"`
	Brand                string      `db:"brand"`
	Category             string      `db:"category"`
	Description          string      `db:"description"`
	Unit                 string      `db:"unit"`
	Multiple             int         `db:"multiple"`
	DealerPrice          money.Money `db:"dealer_price"`
	SuggestionPrice      money.Money `db:"suggestion_price"`
	TechnicalDescription string      `db:"technical_description"`
	Availability         bool        `db:"availability"`
	Length               int         `db:"length"` // mm.
	Width                int         `db:"width"`  // mm.
	Height               int         `db:"height"` // mm.
	Weight               int         `db:"weight"` // grams.
	PictureLink          string      `db:"picture_link"`
	WarrantyPeriod       int         `db:"warranty_period"` // Days.
	RMAProcedure         string      `db:"rma_procedure"`
	CreatedAt            time.Time   `db:"created_at"`
	ChangedAt            time.Time   `db:"changed_at"`
	Changed              bool        `db:"changed"`
	New                  bool        `db:"new"`
	Removed              bool        `db:"removed"`
}

func init_() {
	p := Product{}
	var fieldsName []string
	var fieldsNameDb []string
	var fieldsInterface []interface{}

	val := reflect.ValueOf(&p).Elem()
	for i := 0; i < val.NumField(); i++ {
		fieldType := val.Type().Field(i)
		fieldsName = append(fieldsName, fieldType.Name)
		// fmt.Println(fieldType.Name)
		fieldsNameDb = append(fieldsNameDb, fieldType.Tag.Get("db"))
		// fmt.Println(fieldType.Tag.Get("db"))

		// v := val.Field(i).Addr().Interface().(*string)

		// v := val.Field(i).Addr().Interface()
		// *(v.(*string)) = "asdf"

		fieldsInterface = append(fieldsInterface, val.Field(i).Addr().Interface())
	}
	var buffer bytes.Buffer
	buffer.WriteString("SELECT ")
	buffer.WriteString(strings.Join(fieldsNameDb, ", "))
	buffer.WriteString(" FROM ")
	buffer.WriteString("product ")
	buffer.WriteString("WHERE code=?")
	fmt.Println(buffer.String())
	// fmt.Println(fieldsNameDb)
	// log.Println(p)
	log.Fatal("Fim")
}

func (p *Product) Find(Id string) error {
	var fieldsName []string
	var fieldsNameDb []string
	var fieldsInterface []interface{}

	val := reflect.ValueOf(p).Elem()
	for i := 0; i < val.NumField(); i++ {
		fieldType := val.Type().Field(i)
		fieldsName = append(fieldsName, fieldType.Name)
		// fmt.Println(fieldType.Name)
		fieldsNameDb = append(fieldsNameDb, fieldType.Tag.Get("db"))
		// fmt.Println(fieldType.Tag.Get("db"))

		// v := val.Field(i).Addr().Interface().(*string)

		// v := val.Field(i).Addr().Interface()
		// *(v.(*string)) = "asdf"

		fieldsInterface = append(fieldsInterface, val.Field(i).Addr().Interface())
	}
	var buffer bytes.Buffer
	buffer.WriteString("SELECT ")
	buffer.WriteString(strings.Join(fieldsNameDb, ", "))
	buffer.WriteString(" FROM ")
	buffer.WriteString("product ")
	buffer.WriteString("WHERE code=?")
	// fmt.Println(buffer.String())
	// fmt.Println(fieldsNameDb)
	// log.Println(p)

	err := db.QueryRow(buffer.String(), Id).Scan(fieldsInterface...)
	return err
}

// func init() {
// p := Product{}
// val := reflect.ValueOf(&p).Elem()
// for i := 0; i < val.NumField(); i++ {
// fieldType := val.Type().Field(i)
// fmt.Println(fieldType.Name)
// fmt.Println(fieldType.Tag.Get("db"))
// if i == 0 {
// // v := val.Field(i).Addr().Interface().(*string)
// v := val.Field(i).Addr().Interface()
// *(v.(*string)) = "asdf"
// }
// }
// log.Println(p)
// log.Fatal()
// }

func (p *Product) Find_old(Id string) error {
	err := db.QueryRow(`
		SELECT 
			code, 
			brand, 
			category, 
			description, 
			unit,
			multiple,
			dealer_price,
			suggestion_price,
			technical_description,
			availability, 
			length,
			width,
			height,
			weight,
			picture_link,
			warranty_period,
			rma_procedure,
			created_at,
			changed_at,
			changed,
			new,
			removed
		FROM 
			product 
		WHERE 
			code = ?`, Id).
		Scan(
			&p.Code,
			&p.Brand,
			&p.Category,
			&p.Description,
			&p.Unit,
			&p.Multiple,
			&p.DealerPrice,
			&p.SuggestionPrice,
			&p.TechnicalDescription,
			&p.Availability,
			&p.Length,
			&p.Width,
			&p.Height,
			&p.Weight,
			&p.PictureLink,
			&p.WarrantyPeriod,
			&p.RMAProcedure,
			&p.CreatedAt,
			&p.ChangedAt,
			&p.Changed,
			&p.New,
			&p.Removed)
	return err
}

func (p *Product) Save() error {
	stmt, err := db.Prepare(`
		INSERT INTO product(
			code, 
			brand, 
			category, 
			description, 
			unit,
			multiple,
			dealer_price,
			suggestion_price,
			technical_description,
			availability, 
			length,
			width,
			height,
			weight,
			picture_link,
			warranty_period,
			rma_procedure,
			created_at,
			changed_at,
			changed,
			new,
			removed
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		p.Code,
		p.Brand,
		p.Category,
		p.Description,
		p.Unit,
		p.Multiple,
		p.DealerPrice,
		p.SuggestionPrice,
		p.TechnicalDescription,
		p.Availability,
		p.Length,
		p.Width,
		p.Height,
		p.Weight,
		p.PictureLink,
		p.WarrantyPeriod,
		p.RMAProcedure,
		p.CreatedAt,
		p.ChangedAt,
		p.Changed,
		p.New,
		p.Removed)
	if err != nil {
		return err
	}
	return err
}

func (p *Product) Diff(pn *Product) bool {
	if p.Code != pn.Code {
		return true
	}
	if p.Brand != pn.Brand {
		return true
	}
	if p.Category != pn.Category {
		return true
	}
	if p.Description != pn.Description {
		return true
	}
	if p.Unit != pn.Unit {
		return true
	}
	if p.Multiple != pn.Multiple {
		return true
	}
	if p.DealerPrice != pn.DealerPrice {
		return true
	}
	if p.SuggestionPrice != pn.SuggestionPrice {
		return true
	}
	if p.TechnicalDescription != pn.TechnicalDescription {
		return true
	}
	if p.Availability != pn.Availability {
		return true
	}
	if p.Length != pn.Length {
		return true
	}
	if p.Width != pn.Width {
		return true
	}
	if p.Height != pn.Height {
		return true
	}
	if p.Weight != pn.Weight {
		return true
	}
	if p.PictureLink != pn.PictureLink {
		return true
	}
	if p.WarrantyPeriod != pn.WarrantyPeriod {
		return true
	}
	if p.RMAProcedure != pn.RMAProcedure {
		return true
	}
	return false
}
