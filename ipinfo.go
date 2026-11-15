package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
)

type IPInfo struct {
	IP        netip.Addr `json:"ip"`
	Hostnames []string   `json:"hostnames"`
	AS        *AS        `json:"as"`
	Location  *Location  `json:"location"`
}

func (ii *IPInfo) Get(ip string) error {
	var err1, err2 error
	ii.IP, err1 = netip.ParseAddr(ip)
	if err1 != nil {
		return fmt.Errorf("invalid ip address %q", ip)
	}
	ii.Hostnames, _ = net.LookupAddr(ip)
	ii.AS = &AS{}
	err1 = ii.AS.Get(ip)
	ii.Location = &Location{}
	err2 = ii.Location.Get(ip)
	return errors.Join(err1, err2)
}

func ipInfoLookup(w http.ResponseWriter, r *http.Request) {
	ii := &IPInfo{}
	if err := ii.Get(r.PathValue("ip")); err != nil {
		renderJSON(w, http.StatusInternalServerError, struct {
			Error error `json:"error"`
		}{err})
		return
	}
	renderJSON(w, http.StatusOK, ii)
}
