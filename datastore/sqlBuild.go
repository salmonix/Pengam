package datastore

import (
	"fmt"
	"strings"
)

func CreateSQL(sql string, noOfDatasets int, columns []string) string {

	valsBlock := "("
	valsBlocks := make([]string, 0, noOfDatasets)
	for c := 1; c <= noOfDatasets*len(columns); c++ {
		if c%len(columns) == 0 {
			valsBlock += fmt.Sprintf("$%d)", c)
			valsBlocks = append(valsBlocks, valsBlock)
			valsBlock = "("
		} else {
			valsBlock += fmt.Sprintf("$%d,", c)
		}
	}
	vals := strings.Join(valsBlocks, ",")

	return sql + vals
}
