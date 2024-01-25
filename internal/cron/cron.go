package cron

import (
	"log"
	"time"

	"github.com/redds-be/reddlinks/internal/env"
	"github.com/redds-be/reddlinks/internal/http"
	"github.com/redds-be/reddlinks/internal/utils"
)

// StartJobs starts the jobs that needs to run periodically.
func StartJobs(conf utils.Configuration, envVars env.Env) {
	// Periodically clean the database
	go func(duration time.Duration) {
		for {
			err := conf.CollectGarbage()
			if err != nil {
				log.Println("Could not collect garbage:", err)
			}
			time.Sleep(duration)
		}
	}(time.Duration(envVars.TimeBetweenCleanups) * time.Minute)

	// Generate a new token every x time
	go func(duration time.Duration) {
		for {
			http.Token = utils.RandomToken()
			time.Sleep(duration)
		}
	}(3 * time.Hour) //nolint:gomnd
}
