package search

import "ash/gohunt/db"

// in memory representation of our search index, Inverted index
type Index map[string][]string

func (idx Index) Add (docs []db.CrawledUrl) {
	for _, doc := range docs {
		for _, token := range analyze(doc.Url + " " + doc.PageTitle + " " + doc.PageDescription + " " + doc.Headings) {
			ids := idx[token]
			if ids != nil && ids[len(ids) - 1] == doc.ID {
				continue
			}
			idx[token] = append(ids, doc.ID)
		}
	}
}