package write

import "log"

func Zip(records []map[string][]byte, file string) {
	log.Println(len(records), "records zip to file", file)
}
