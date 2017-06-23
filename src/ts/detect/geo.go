package detect

import (
	"github.com/abh/geoip"
	"os"
	"path/filepath"
)

type (
	Geo struct {
		*geoip.GeoIP
	}
	GeoRecord struct {
		geoip.GeoIPRecord `bson:"GeoIPRecord,inline"`
	}
)

const (
	GEO_CITY_DB_FILE string = "/geo/GeoIPCity.dat"
	GEO_ORG_DB_FILE  string = "/geo/GeoIPOrg.dat"
)

func NewGeo() (*Geo, error) {
	curDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, err
	}

	gip, err := geoip.Open(filepath.Dir(curDir) + GEO_CITY_DB_FILE)
	if err != nil {
		return nil, err
	}

	g := &Geo{}
	g.GeoIP = gip
	return g, nil
}

func (s *Geo) GetRecord(ip string) *GeoRecord {
	r := s.GeoIP.GetRecord(ip)
	if r == nil {
		return nil
	}

	g := &GeoRecord{}
	g.GeoIPRecord = *r
	return g
}
