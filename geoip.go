package main

import (
	"fmt"
	"net/http"
	"net/netip"
)

type Location struct {
	Continent     string `json:"continent"`
	ContinentCode string `json:"continent_code"`
	Country       string `json:"country"`
	CountryCode   string `json:"country_code"`
	City          string `json:"city"`
}

func (loc *Location) Get(ip string) error {
	ipaddr, err := netip.ParseAddr(ip)
	if err != nil {
		return fmt.Errorf("invalid ip address %q", ip)
	}
	geo, err := locReader.City(ipaddr)
	if err != nil {
		return err
	}
	if !geo.HasData() {
		return nil
	}
	loc.Continent = geo.Continent.Names.English
	loc.ContinentCode = geo.Continent.Code
	loc.Country = geo.Country.Names.English
	loc.CountryCode = geo.Country.ISOCode
	loc.City = geo.City.Names.English
	return nil
}

func geoLookup(w http.ResponseWriter, r *http.Request) {
	loc := &Location{}
	if err := loc.Get(r.PathValue("ip")); err != nil {
		renderJSON(w, http.StatusInternalServerError, struct {
			Error error `json:"error"`
		}{err})
		return
	}
	renderJSON(w, http.StatusOK, loc)
}
