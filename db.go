package main

import (
	"net/http"
	"strings"

	"github.com/oschwald/geoip2-golang/v2"
)

const (
	defaultAsnDB = "./dbip-asn-lite.mmdb"
	defaultGeoDB = "./dbip-city-lite.mmdb"
)

var (
	asnDB,
	geoDB *string

	asnReader,
	locReader *geoip2.Reader
)

func loadDB(file string) (*geoip2.Reader, error) {
	return geoip2.Open(file)
}

func unloadDB(db *geoip2.Reader) error {
	return db.Close()
}

func reloadDB(w http.ResponseWriter, r *http.Request) {
	var (
		db *string
		rd *geoip2.Reader
		isASN,
		isGeo bool
	)
	if strings.Contains(r.RequestURI, "/asn/") {
		db = asnDB
		rd = asnReader
		isASN = true
	} else if strings.Contains(r.RequestURI, "/geo/") {
		db = geoDB
		rd = locReader
		isGeo = true
	} else {
		renderJSON(w, http.StatusInternalServerError, jsonMessage{
			Message: "unable to identify db to reload; skipping...",
		})
	}
	newReader, err := loadDB(*db)
	if err == nil {
		err = unloadDB(rd)
	}
	if err != nil {
		renderJSON(w, http.StatusInternalServerError, jsonMessage{
			Message: "failed to load database; using previous one...",
			Err:     err,
		})
		return
	}
	if isASN {
		asnReader = newReader
	}
	if isGeo {
		locReader = newReader
	}
	renderJSON(w, http.StatusOK, jsonMessage{
		Message: "database reloaded successfully",
	})
}
