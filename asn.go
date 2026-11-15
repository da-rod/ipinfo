package main

import (
	"fmt"
	"net/http"
	"net/netip"
)

type AS struct {
	Number uint   `json:"number"`
	Name   string `json:"name"`
}

func (as *AS) Get(ip string) error {
	ipaddr, err := netip.ParseAddr(ip)
	if err != nil {
		return fmt.Errorf("invalid ip address %q", ip)
	}
	asn, err := asnReader.ASN(ipaddr)
	if err != nil {
		return err
	}
	if !asn.HasData() {
		return nil
	}
	as.Number = asn.AutonomousSystemNumber
	as.Name = asn.AutonomousSystemOrganization
	return nil
}

func asLookup(w http.ResponseWriter, r *http.Request) {
	as := &AS{}
	if err := as.Get(r.PathValue("ip")); err != nil {
		renderJSON(w, http.StatusInternalServerError, struct {
			Error error `json:"error"`
		}{err})
		return
	}
	renderJSON(w, http.StatusOK, as)
}
