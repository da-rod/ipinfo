package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/oschwald/geoip2-golang"
)

const (
	defaultAddr  = ":8080"
	defaultMode  = "release"
	defaultLang  = "en"
	defaultAsnDB = "./dbip-asn-lite-2021-06.mmdb"
	defaultGeoDB = "./dbip-city-lite-2021-06.mmdb"
)

var (
	addr, asnDB, geoDB, lang *string
	asnReader, locReader     *geoip2.Reader
)

type as struct {
	Number uint
	Name   string
}

type location struct {
	Continent     string
	ContinentCode string
	Country       string
	CountryCode   string
	City          string
}

type ipinfo struct {
	AS       as
	Location location
}

func init() {
	// Lookup environment variables
	var a, m, aDB, gDB, l string
	if a = os.Getenv("IPINFO_ADDR"); a == "" {
		a = defaultAddr
	}
	if m = os.Getenv("IPINFO_MODE"); m == "" {
		m = defaultMode
	}
	if aDB = os.Getenv("IPINFO_DB_ASN"); aDB == "" {
		aDB = defaultAsnDB
	}
	if gDB = os.Getenv("IPINFO_DB_GEOIP"); gDB == "" {
		gDB = defaultGeoDB
	}
	if l = os.Getenv("IPINFO_LANG"); l == "" {
		l = defaultLang
	}

	// Parse arguments
	addr = flag.String("a", a, "Listening address:port")
	mode := flag.String("m", m, "Gin mode (available modes: debug, test, release)")
	asnDB = flag.String("db_asn", aDB, "ASN mmdb file")
	geoDB = flag.String("db_geoip", gDB, "GeoIP mmdb file")
	lang = flag.String("l", l, "Language used for names (available languages: de, en, es, fr, ja, pt-BR, ru, zh-CN)")
	flag.Parse()

	switch *lang {
	case "en", "de", "es", "fr", "ja", "pt-BR", "ru", "zh-CN":
	default:
		// Fallback to English
		*lang = "en"
	}

	// Set Gin mode
	gin.SetMode(*mode)

	// Load databases
	var err error
	asnReader, err = loadDB(*asnDB)
	if err != nil {
		panic(err)
	}
	locReader, err = loadDB(*geoDB)
	if err != nil {
		panic(err)
	}
}

func loadDB(file string) (*geoip2.Reader, error) {
	return geoip2.Open(file)
}

func unloadDB(db *geoip2.Reader) {
	db.Close()
}

func main() {
	// Setup router
	r := gin.Default()

	// ASN
	r.GET("/asn/reload", func(c *gin.Context) {
		newReader, err := loadDB(*asnDB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "failed to load database; using previous one...",
			})
		} else {
			unloadDB(asnReader)
			asnReader = newReader
			c.JSON(http.StatusOK, gin.H{"message": "asn database reloaded successfully"})
		}
	})

	r.GET("/asn/:ip", func(c *gin.Context) {
		if asn, err := getAS(c.Param("ip")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, asn)
		}
	})

	// GeoIP data
	r.GET("/geo/reload", func(c *gin.Context) {
		newReader, err := loadDB(*geoDB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "failed to load database; using previous one...",
			})
		} else {
			unloadDB(locReader)
			locReader = newReader
			c.JSON(http.StatusOK, gin.H{"message": "geoip database reloaded successfully"})
		}
	})

	r.GET("/geo/:ip", func(c *gin.Context) {
		if geo, err := getLocation(c.Param("ip")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, geo)
		}
	})

	// IP Info (ASN + GeoIP combined)
	r.GET("/ipinfo/:ip", func(c *gin.Context) {
		if ipdata, err := getIPInfo(c.Param("ip")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, ipdata)
		}
	})

	r.Run(*addr)
}

func getAS(ip string) (as, error) {
	ipaddr := net.ParseIP(ip)
	if ipaddr == nil {
		return as{}, fmt.Errorf("invalid ip address %q", ip)
	}
	data, err := asnReader.ASN(ipaddr)
	if err != nil {
		return as{}, err
	}
	return as{
		Number: data.AutonomousSystemNumber,
		Name:   data.AutonomousSystemOrganization,
	}, nil
}

func getLocation(ip string) (location, error) {
	ipaddr := net.ParseIP(ip)
	if ipaddr == nil {
		return location{}, fmt.Errorf("invalid ip address %q", ip)
	}
	geo, err := locReader.City(ipaddr)
	if err != nil {
		return location{}, err
	}
	return location{
		Continent:     geo.Continent.Names[*lang],
		ContinentCode: geo.Continent.Code,
		Country:       geo.Country.Names[*lang],
		CountryCode:   geo.Country.IsoCode,
		City:          geo.City.Names[*lang],
	}, nil
}

func getIPInfo(ip string) (ipinfo, error) {
	asData, _ := getAS(ip)
	loData, err := getLocation(ip)
	return ipinfo{
		asData,
		loData,
	}, err
}
