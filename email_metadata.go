// email-metadata will walk a directory of mime encoded email files
// and dump (some) envelope metadata to a csv
package main

import (
	"encoding/csv"
	"io/fs"
	"log"
	"net/mail"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/jhillyerd/enmime"
)

var receivedRe = regexp.MustCompile(`.* id ([^ ]+) for ([^ ;]+); (.*)`)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <dir>", os.Args[0])
	}

	walkDir := os.Args[1]

	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	w.Write([]string{"id", "path", "date", "dst", "to", "from", "subject", "cc"})

	err := filepath.Walk(walkDir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		msg, err := enmime.ReadEnvelope(f)
		if err != nil {
			log.Printf("read email header for %s err %s", path, err)
			return nil
		}

		id := path
		received := msg.GetHeader("received")
		match := receivedRe.FindStringSubmatch(received)
		if len(match) > 1 {
			id = match[1]
		}
		var dst string
		if len(match) > 2 {
			dst = match[2]
		}
		date := msg.GetHeader("date")
		if len(match) > 3 {
			dateStr := match[3]

			ts, err := mail.ParseDate(dateStr)
			if err == nil {
				date = ts.Format(time.RFC3339)
			}
		}

		from := msg.GetHeader("from")
		to := msg.GetHeader("to")
		subject := msg.GetHeader("subject")
		cc := msg.GetHeader("cc")

		mail.ParseAddressList(to)

		w.Write([]string{id, path, date, dst, to, from, subject, cc})
		w.Flush()
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
