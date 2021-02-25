// Command text2 is a chromedp example demonstrating how to extract text from a
// specific element.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
)

type YCJobListing struct{
	Title string	`json:"posting_name"`
	Url string		`json:"posting_url"`
	Age string		`json:"posting_age"`
	IntAge int		`json:"posting_age_numerical"`
}

func main() {
	// Initialize logger
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	// create context
	log.Info("Initializing Context")
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	log.Info("Getting Job Listings")
	jobListings, _ := getJobListings(ctx)

	log.Info("Converting Job Listings to JSON")
	jobListingsString, _ := json.Marshal(jobListings)

	log.Info(fmt.Sprintf("Got Result: %v", string(jobListingsString)))

	filename := "output.json"
	ioutil.WriteFile(filename, jobListingsString, os.ModePerm)

}


func getJobListings(ctx context.Context) ([]YCJobListing, error){
	// Get Job Listings from a page on the YCombinator hacker news jobs page

	// Define output array
	jobListings := []YCJobListing{}

	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Navigate to page
	if err := chromedp.Run(ctx, chromedp.Navigate(`https://news.ycombinator.com/jobs`)); err != nil {
		return nil, fmt.Errorf("Error getting to yc link")
	}

	// Scrape Posting titles data
	var postingTitles []*cdp.Node
	if err := chromedp.Run(ctx, chromedp.Nodes(`.storylink`, &postingTitles)); err != nil {
		return nil, fmt.Errorf("Error getting to job posting object: %v", err)
	}
	
	// Scrape Posting dates data
	var postingDates []*cdp.Node
	if err := chromedp.Run(ctx, chromedp.Nodes(`.age a`, &postingDates)); err != nil {
		return nil, fmt.Errorf("Error getting jon posting dates: %v", err)
	}

	// Check if lengths of these two are the same
	if len(postingTitles) != len(postingDates) {
		log.Fatal("Length of posting titles and dates do not align!!")
	}

	// Parse data into a new struct and append to output array
	for i := 0; i < len(postingTitles); i++ {
		newListing := YCJobListing{
			Title: postingTitles[i].Children[0].NodeValue,
			Url: postingTitles[i].AttributeValue("href"),
			Age: postingDates[i].Children[0].NodeValue,
			IntAge: convertStringDateToInt(postingDates[i].Children[0].NodeValue),
		}

		jobListings = append(jobListings, newListing)
	}

	return jobListings, nil
}

func convertStringDateToInt(postingDate string) int{
	// Converts string like '8 days ago' to number of hours in 8 days, returns integer

	out := 0
	splitDate := strings.Split(postingDate, " ")

	if splitDate[1] == "hours" || splitDate[1] == "hour"{
		intDate, err := strconv.Atoi(splitDate[0])
		if err != nil{
			panic(err)
		}
		out = intDate
	}else if splitDate[1] == "day" || splitDate[1] == "days" {
		intDate, err := strconv.Atoi(splitDate[0])
		if err != nil{
			panic(err)
		}
		out = intDate * 24
	}

	return out
}