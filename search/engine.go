package search

import (
	"ash/gohunt/db"
	"fmt"
	"time"
)

func RunEngine() {
	fmt.Println("started search engine crawl...")
	defer fmt.Println("search engine crawl has finished")
	
	// Get Settings
	settings := &db.SearchSettings{}
	err := settings.Get()
	if err != nil {
		fmt.Println("something went wrong getting the settings")
		return
	}
	
	// Search the search engine is on
	if !settings.SearchOn {
		fmt.Println("search is turned off")
		return
	}
	
	crawl := &db.CrawledUrl{}
	nextUrls, err := crawl.GetNextCrawlUrls(int(settings.Amount))
	if err != nil {
		fmt.Println("something went wrong getting next urls")
		return
	}

	newUrls := []db.CrawledUrl{}
	testedTime := time.Now()
	for _, next := range nextUrls {
		result := runCrawl(next.Url)
		if !result.Success {
			// Update row in database with the failed crawl
			err := next.UpdatedUrl(db.CrawledUrl{
				ID:              next.ID,
				Url:             next.Url,
				Success:         false,
				CrawlDuration:   result.CrawlData.CrawlTime,
				ResponseCode:    result.ResponseCode,
				PageTitle:       result.CrawlData.PageTitle,
				PageDescription: result.CrawlData.PageDescription,
				Headings:        result.CrawlData.Headings,
				LastTested:      &testedTime,
			})
			if err != nil {
				fmt.Println("something went wrong updating a failed url")
			}
			continue
		}
		// Update a successful row in database
		err := next.UpdatedUrl(db.CrawledUrl{
			ID:              next.ID,
			Url:             next.Url,
			Success:         result.Success,
			CrawlDuration:   result.CrawlData.CrawlTime,
			ResponseCode:    result.ResponseCode,
			PageTitle:       result.CrawlData.PageTitle,
			PageDescription: result.CrawlData.PageDescription,
			Headings:        result.CrawlData.Headings,
			LastTested:      &testedTime,
		})
		if err != nil {
			fmt.Printf("something went wrong updating %v /n", next.Url)
		}

		// Push the newly found external urls to an array
		for _, newUrl := range result.CrawlData.Links.External {
			newUrls = append(newUrls, db.CrawledUrl{Url: newUrl})
		}
	}

	// Check if we should add the newly found urls to the database
	if !settings.AddNew {
		fmt.Printf("Adding new urls to database is disabled")
		return
	}

	// Insert new urls
	for _, newUrl := range newUrls {
		err := newUrl.Save()
		if err != nil {
			fmt.Println("something went wrong adding the new url to the database")
		}
	}
	fmt.Printf("\n Added %d new urls to the database", len(newUrls))
}

func RunIndex() {
	fmt.Println("started search indexing...")
	defer fmt.Println("search indexing has finished")
	crawled := &db.CrawledUrl{}
	notIndexed, err := crawled.GetNotIndex()
	if err != nil {
		return
	}

	idx := make(Index)
	idx.Add(notIndexed)
	searchIndex := &db.SearchIndex{}
	err = searchIndex.Save(idx, notIndexed)
	if err != nil {
		return
	}
	err = crawled.SetIndexedTrue(notIndexed)
	if err != nil {
		return
	}
}