// This utility validates Time Zone dataset against osm database and tz shapefile,
// double check the results, sometimes false positives
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/TehPeGaSuS/xmasbot/xmas"
	"github.com/ringsaturn/tzf"
)

var tzFinder tzf.F

func init() {
	f, err := tzf.NewDefaultFinder()
	if err != nil {
		log.Fatal(err)
	}
	tzFinder = f
}

// Set target year
var target = func() time.Time {
	tmp := time.Now().UTC()
	if tmp.Month() == time.December && tmp.Day() < 26 {
		return time.Date(tmp.Year(), time.December, 25, 0, 0, 0, 0, time.UTC)
	}
	return time.Date(tmp.Year()+1, time.December, 25, 0, 0, 0, 0, time.UTC)
}()

var email *string
var nominatim *string

func main() {
	email = flag.String("email", "", "nominatim email")
	nominatim = flag.String("nominatim", "https://nominatim.openstreetmap.org", "nominatim server")
	flag.Parse()
	if *email == "" {
		fmt.Fprintf(os.Stderr, "%s", "provide email with -email flag\n")
		return
	}
	var zones xmas.TZS
	if err := json.Unmarshal(xmas.Zones, &zones); err != nil {
		log.Fatal(err)
	}
	//print target to be sure
	fmt.Println("Target:", target)
	sort.Sort(sort.Reverse(zones))
	for _, zone := range zones {
		fmt.Println("Zone:", zone.Offset)
		for _, country := range zone.Countries {
			if len(country.Cities) == 0 {
				remoteOffset, err := timeZone(country.Name, "")
				if err != nil {
					log.Println(country.Name, err)
				} else {
					if remoteOffset != zone.Offset {
						fmt.Printf("%s: Offset mismatch; Local: %v, Remote: %v\n",
							country.Name, zone.Offset, remoteOffset)
					}
				}
			}
			for _, city := range country.Cities {
				remoteOffset, err := timeZone(country.Name, city)
				if err != nil {
					log.Println(city+", "+country.Name, err)
				} else {
					if remoteOffset != zone.Offset {
						fmt.Printf("%s, %s: Offset mismatch; Local: %v, Remote: %v\n",
							city, country.Name, zone.Offset, remoteOffset)
					}
				}
			}
		}
	}
}

// Get Timezone Offset
func timeZone(country, city string) (float64, error) {
	var mapj xmas.NominatimResults
	var err error
	if country == "CAR" || country == "Congo" {
		mapj, err = xmas.NominatimFetcherLong(*email, *nominatim, country, "", "")
		if err != nil {
			return 0, err
		}
	} else {
		mapj, err = xmas.NominatimFetcher(*email, *nominatim, country+", "+city)
		if mapj != nil && len(mapj) == 0 {
			mapj, err = xmas.NominatimFetcher(*email, *nominatim, city+", "+country)
		}
		if err != nil {
			return 0, err
		}
	}
	if len(mapj) == 0 {
		return 0, fmt.Errorf("no results")
	}

	tzid := tzFinder.GetTimezoneName(mapj[0].Lon, mapj[0].Lat)
	if tzid == "" {
		return 0, fmt.Errorf("no timezone found")
	}
	zone, err := time.LoadLocation(tzid)
	if err != nil {
		return 0, err
	}
	offset := zoneOffset(target, zone)
	return float64(offset) / 60 / 60, nil
}

func zoneOffset(target time.Time, zone *time.Location) int {
	_, offset := time.Date(target.Year(), target.Month(), target.Day(),
		target.Hour(), target.Minute(), target.Second(),
		target.Nanosecond(), zone).Zone()
	return offset
}
