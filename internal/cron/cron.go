//    reddlinks, a simple link shortener written in Go.
//    Copyright (C) 2024 redd
//
//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU General Public License as published by
//    the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.
//
//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU General Public License for more details.
//
//    You should have received a copy of the GNU General Public License
//    along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Package cron is used to start "cron" jobs.
package cron

import (
	"log"
	"time"

	"github.com/redds-be/reddlinks/internal/env"
	"github.com/redds-be/reddlinks/internal/utils"
)

// StartJobs starts the jobs that needs to run periodically.
//
// This functions starts a go routine that starts an infinite loop
// that will call [utils.CollectGarbage] after a sleep period defined in the env variables.
// [utils.CollectGarbage] is called before sleeping to ensure
// garbage collection is done earch time the program is started.
// Previously, there were two functions called in this function,
// As of now, only [utils.CollectGarbage] making its usefulness questionable.
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
}
