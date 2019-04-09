// 3 october 2014

package main

import (
	"os"
	"io"
	"archive/zip"
	"path/filepath"
	"strings"
	"strconv"
	"time"
	"golang.org/x/net/html"
)

import "C"

type Entry struct {
	Name	string
	Book		string
	ID		string
	Parent	string
	Order	int
	Date		time.Time
	MSHC	string
	File		string

	Children	[]*Entry
	Dest		string
}

var entries []*Entry

func parseEntry(r io.Reader, mshcname string, filename string) {
	var e *Entry

	e = new(Entry)
	e.MSHC = mshcname
	e.File = filename
	t := html.NewTokenizer(r)
	for {
		tt := t.Next()
		if tt == html.ErrorToken {
			err := t.Err()
			if err == io.EOF {
				break
			}
			panic(err)		// TODO
		}
		tok := t.Token()
		switch tok.Type {
		case html.StartTagToken, html.SelfClosingTagToken:
			if tok.Data != "meta" {
				break
			}

			var where *string
			var what string
			var order string
			var date string
			var err error

			for _, a := range tok.Attr {
				if a.Key == "name" {
					switch a.Val {
					case "Title":
						where = &e.Name
					case "Microsoft.Help.Book":
						where = &e.Book
					case "Microsoft.Help.Id":
						where = &e.ID
					case "Microsoft.Help.TocParent":
						where = &e.Parent
					case "Microsoft.Help.TocOrder":
						where = &order
					case "Microsoft.Help.TopicPublishDate":
						where = &date
					}
				} else if a.Key == "content" {
					what = a.Val
				}
			}
			if where != nil {
				*where = what
			}
			if where == &order {
				e.Order, err = strconv.Atoi(order)
				if err != nil {
					panic(err)		// TODO
				}
			} else if where == &date {
				e.Date, err = time.Parse(time.RFC1123, date)
				if err != nil {
					panic(err)		// TODO
				}
			}
		}
	}
	entries = append(entries, e)
}

func parseMSHC(mshcname string) {
	z, err := zip.OpenReader(mshcname)
	if err != nil {
		panic(err)
	}
	defer z.Close()
	for _, f := range z.File {
		if !strings.HasPrefix(f.Name, "ic") && filepath.Ext(f.Name) != ".htm" {
			continue
		}
		r, err := f.Open()
		if err != nil {
			panic(err)		// TODO
		}
		if strings.HasPrefix(f.Name, "ic") {
			addAsset(mshcname, f.Name, r)
		} else {
			parseEntry(r, mshcname, f.Name)
		}
		r.Close()
	}
}

func main() {
	for _, cab := range os.Args[2:] {
		parseMSHC(cab)
	}
	collectByID()
	assignChildren()
	sortChildren()
	buildDestinationFolder(os.Args[1])
	buildDevhelp(os.Args[1])
}
