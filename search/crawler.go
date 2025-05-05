package search

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type CrawlData struct {
	Url          string
	Success      bool
	ResponseCode int
	CrawlData    ParsedBody
}

type ParsedBody struct {
	CrawlTime       time.Duration
	PageTitle       string
	PageDescription string
	Headings        string
	Links           Links
}

type Links struct {
	Internal []string
	External []string
}

func runCrawl(inputUrl string) CrawlData {
	resp, err := http.Get(inputUrl)
	baseUrl, _ := url.Parse(inputUrl)

	// Check if error or if response is empty
	if err != nil || resp == nil {
		fmt.Println("something went wrong fetching the body")
		return CrawlData{Url: inputUrl, Success: false, ResponseCode: 0, CrawlData: ParsedBody{}}
	}
	defer resp.Body.Close()

	// Check for 200 OK
	if resp.StatusCode != 200 {
		fmt.Println("not found status code 200")
		return CrawlData{Url: inputUrl, Success: false, ResponseCode: resp.StatusCode, CrawlData: ParsedBody{}}
	}

	// Check for html
	contentType := resp.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "text/html") {
		//Response is html
		data, err := parseBody(resp.Body, baseUrl)
		if err != nil {
			fmt.Println("something went wrong getting data from html body")
			return CrawlData{Url: inputUrl, Success: false, ResponseCode: resp.StatusCode, CrawlData: ParsedBody{}}
		}
		return CrawlData{Url: inputUrl, Success: true, ResponseCode: resp.StatusCode, CrawlData: data}
	} else {
		fmt.Println("not found html response")
		return CrawlData{Url: inputUrl, Success: false, ResponseCode: resp.StatusCode, CrawlData: ParsedBody{}}
	}

}

func parseBody(body io.Reader, baseUrl *url.URL) (ParsedBody, error) {
	doc, err := html.Parse(body)
	if err != nil {
		fmt.Println(err)
		fmt.Println("something went wrong parsing body")
		return ParsedBody{}, err
	}
	start := time.Now()

	// Get Links
	links := getLinks(doc, baseUrl)
	// Get Page Title and Description
	title, desc := getPageData(doc)
	// Get H1 Tags
	headings := getPageHeadings(doc)
	// Record timings
	end := time.Now()
	//Return the data
	return ParsedBody{
		CrawlTime:       end.Sub(start),
		PageTitle:       title,
		PageDescription: desc,
		Headings:        headings,
		Links:           links,
	}, nil
}

func getPageData(node *html.Node) (string, string) {
	if node == nil {
		return "", ""
	}
	// Find title and description
	title, desc := "", ""
	var findMetaAndTitle func(*html.Node)
	findMetaAndTitle = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "title" {
			// Check if empty
			if (node.FirstChild == nil) {
				title = ""
			} else {
				title = node.FirstChild.Data
			}
		} else if node.Type == html.ElementNode && node.Data == "meta" {
			var name, content string
			for _, attr := range node.Attr {
				if attr.Key == "name" {
					name = attr.Val
				} else if attr.Key == "content" {
					content = attr.Val
				}
			}
			if name == "description" {
				desc = content
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		findMetaAndTitle(child)
	}
	findMetaAndTitle(node)
	return title, desc
}

func getLinks(node *html.Node, baseUrl *url.URL) Links {
	links := Links{}
	if node == nil {
		return links
	}
	var findLinks func(*html.Node)
	findLinks = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					url, err := url.Parse(attr.Val)
					if err != nil || strings.HasPrefix(url.String(), "#") || strings.HasPrefix(url.String(), "mail") ||
						strings.HasPrefix(url.String(), "tel") || strings.HasPrefix(url.String(), "javascript") ||
						strings.HasSuffix(url.String(), ".pdf") || strings.HasSuffix(url.String(), ".md") {
						continue
					}
					if url.IsAbs() {
						if isSameHost(url.String(), baseUrl.String()) {
							links.Internal = append(links.Internal, url.String())
						} else {
							links.External = append(links.External, url.String())
						}
					} else {
						resolver := baseUrl.ResolveReference(url)
						links.Internal = append(links.Internal, resolver.String())
					}
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			findLinks(child)
		}
	}
	findLinks(node)
	return links
}

func isSameHost(absoluteUrl string, baseUrl string) bool {
	absUrl, err := url.Parse(absoluteUrl)
	if err != nil {
		return false
	}
	baseUrlParsed, err := url.Parse(baseUrl)
	if err != nil {
		return false
	}
	return absUrl.Host == baseUrlParsed.Host
}

func getPageHeadings(node *html.Node) string {
	if node == nil {
		return ""
	}
	var headings strings.Builder
	var findH1 func(*html.Node)
	findH1 = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "h1" {
			// check if node is empty
			if node.FirstChild != nil {
				headings.WriteString(node.FirstChild.Data)
				headings.WriteString(", ")
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			findH1(child)
		}
	}
	// remove the last comma
	findH1(node)
	return strings.TrimSuffix(headings.String(), ", ")
}