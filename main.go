package main

import (
	"compress/gzip"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []URL    `xml:"url"`
}

type URL struct {
	Loc string `xml:"loc"`
}

type SitemapIndex struct {
	XMLName  xml.Name  `xml:"sitemapindex"`
	Sitemaps []Sitemap `xml:"sitemap"`
}

type Sitemap struct {
	Loc string `xml:"loc"`
}

type SitemapExtractor struct {
	urls map[string]bool
}

func NewSitemapExtractor() *SitemapExtractor {
	return &SitemapExtractor{
		urls: make(map[string]bool),
	}
}

func (se *SitemapExtractor) downloadSitemap(url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("User-Agent", "curl/8.7.1")
	req.Header.Set("Accept", "application/xml,text/xml,application/gzip,*/*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error downloading sitemap: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	var reader io.Reader = resp.Body

	if strings.HasSuffix(url, ".gz") || resp.Header.Get("Content-Type") == "application/x-gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error creating gzip reader: %v", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	return io.ReadAll(reader)
}

func (se *SitemapExtractor) parseSitemap(url string) error {
	content, err := se.downloadSitemap(url)
	if err != nil {
		return err
	}

	var sitemapIndex SitemapIndex
	if err := xml.Unmarshal(content, &sitemapIndex); err == nil && len(sitemapIndex.Sitemaps) > 0 {
		log.Printf("Processing sitemap index: %s", url)
		for _, sitemap := range sitemapIndex.Sitemaps {
			if err := se.parseSitemap(sitemap.Loc); err != nil {
				log.Printf("Error processing nested sitemap %s: %v", sitemap.Loc, err)
			}
		}
		return nil
	}

	var urlset URLSet
	if err := xml.Unmarshal(content, &urlset); err != nil {
		return fmt.Errorf("error parsing sitemap XML: %v", err)
	}

	log.Printf("Processing sitemap: %s", url)
	for _, url := range urlset.URLs {
		se.urls[url.Loc] = true
	}

	return nil
}

func (se *SitemapExtractor) ExtractURLs(sitemapURL string) []string {
	if err := se.parseSitemap(sitemapURL); err != nil {
		log.Printf("Error extracting URLs: %v", err)
	}

	var urls []string
	for url := range se.urls {
		urls = append(urls, url)
	}
	sort.Strings(urls)
	return urls
}

func (se *SitemapExtractor) ExportURLs(filename string, urls []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer file.Close()

	for _, url := range urls {
		if _, err := fmt.Fprintln(file, url); err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}

	log.Printf("Successfully exported %d URLs to %s", len(urls), filename)
	return nil
}

func main() {
	output := flag.String("o", "", "Output file path. If not specified, URLs will be printed to stdout")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Usage: sitemap [-o output.txt] URL")
		os.Exit(1)
	}

	sitemapURL := flag.Arg(0)
	extractor := NewSitemapExtractor()
	urls := extractor.ExtractURLs(sitemapURL)

	if *output != "" {
		if err := extractor.ExportURLs(*output, urls); err != nil {
			log.Fatalf("Error exporting URLs: %v", err)
		}
	} else {
		for _, url := range urls {
			fmt.Println(url)
		}
	}
}
