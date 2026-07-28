// This utility tests TZ string splitting
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"

	"github.com/TehPeGaSuS/xmasbot/xmas"
)

func main() {
	var zones xmas.TZS
	err := json.Unmarshal(xmas.Zones, &zones)
	if err != nil {
		log.Fatal(err)
	}

	sort.Sort(sort.Reverse(zones))

	for _, k := range zones {
		fmt.Println("**********************")
		fmt.Println(k.Format(200, true))
	}
}
