package utils

import (
	"ash/gohunt/search"
	"fmt"

	"github.com/robfig/cron"
)

func StartCronJobs() {
	c := cron.New()
	c.AddFunc("0 * * * *", search.RunEngine) // Run every hour
	c.AddFunc("15 * * * *", search.RunIndex) // Run every hour at 15 mins past
	c.Start()
	cronCount := len(c.Entries())
	fmt.Printf("setup %d cron jobs\n", cronCount)
}