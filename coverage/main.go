package main

import "flag"

func main() {
	flag.Parse()
	var analyser = gotoolcover{}
	j := analyser.analyseCoverage()
	alarmJson(j)

}

type gotoolcover struct{}

func (gotoolcover) analyseCoverage() *JSONData {

	return analyseGoToolsCover()
}
