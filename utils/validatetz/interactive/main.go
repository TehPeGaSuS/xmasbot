// Interactive utility for querying location info
// Reads lines from stdin prints to stdout
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/TehPeGaSuS/xmasbot/xmas"
	"github.com/ringsaturn/tzf"
	"github.com/ringsaturn/tzf/convert"
)

var tzFinder tzf.F

var email *string
var nominatim *string
var ext *string

// Set target year
var target = func() time.Time {
	tmp := time.Now().UTC()
	if tmp.Month() == time.December && tmp.Day() < 26 {
		return time.Date(tmp.Year(), time.December, 25, 0, 0, 0, 0, time.UTC)
	}
	return time.Date(tmp.Year()+1, time.December, 25, 0, 0, 0, 0, time.UTC)
}()

func main() {
	ext = flag.String("ext", "", "external geojson")
	email = flag.String("email", "", "nominatim email")
	nominatim = flag.String("nominatim", "https://nominatim.openstreetmap.org", "nominatim server")
	flag.Parse()
	if *email == "" {
		fmt.Fprintf(os.Stderr, "%s", "provide email with -email flag\n")
		return
	}
	if *ext != "" {
		f, err := os.OpenFile(*ext, os.O_RDONLY, 0655)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		data, err := io.ReadAll(f)
		if err != nil {
			panic(err)
		}
		var boundary convert.BoundaryFile
		if err := json.Unmarshal(data, &boundary); err != nil {
			panic(err)
		}
		finder, err := tzf.NewFinderFromRawJSON(&boundary)
		if err != nil {
			panic(err)
		}
		tzFinder = finder
	} else {
		finder, err := tzf.NewDefaultFinder()
		if err != nil {
			panic(err)
		}
		tzFinder = finder
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		result, err := locationInfo(scanner.Text())
		if err == nil {
			fmt.Printf("%s\n", result)
		} else {
			fmt.Printf("%s, Error: %s\n", scanner.Text(), err)
		}
	}
}

func locationInfo(location string) (string, error) {
	mapj, err := xmas.NominatimFetcher(*email, *nominatim, location)
	if err != nil {
		return "", err
	}

	if len(mapj) == 0 {
		return "", errors.New("status not OK")
	}
	now := time.Now()
	tzid := tzFinder.GetTimezoneName(mapj[0].Lon, mapj[0].Lat)
	lookup := time.Now().Sub(now)
	if tzid == "" {
		return "", errors.New("no timezone found")
	}
	zone, err := time.LoadLocation(tzid)
	if err != nil {
		return "", err
	}
	offset := zoneOffset(target, zone)
	return fmt.Sprintf("%s, Offset %v, zone: %s, time now: %s, tz dur %s", mapj[0].DisplayName, float64(offset)/60/60, zone, time.Now().In(zone), lookup), nil
}

func zoneOffset(target time.Time, zone *time.Location) int {
	_, offset := time.Date(target.Year(), target.Month(), target.Day(),
		target.Hour(), target.Minute(), target.Second(),
		target.Nanosecond(), zone).Zone()
	return offset
}
