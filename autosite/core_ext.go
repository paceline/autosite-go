/*
    Package autosite provides a simple infrastructure for running a
    personal website (off of the Google App Engine)
    
    Created by Ulf Möhring <hello@ulfmoehring.net>
*/

package autosite

import (
	"html/template"
	"strconv"
    "time"
)

// Look for a string in string array and return position
func LookFor(haystack []string, needle string) int {
	for i := 0; i < len(haystack); i++ {
		if haystack[i] == needle {
			return i
		}
	}
	return len(haystack)
}

// Helper: Return time in proper format
func formatTime(t time.Time) string {
	elapsed := time.Since(t).Seconds()
	if elapsed < 60 {
		return "A few seconds ago"
	}
	if elapsed >= 60 && elapsed < 29030400 {
		var t string
		var u string	
		switch {
			case elapsed < 3600:
				t = strconv.FormatFloat(elapsed/60, 'f', 0, 64)
				u = "minute"
			case elapsed >= 3600 && elapsed < 86400:
				t = strconv.FormatFloat(elapsed/60/60, 'f', 0, 64)
				u = "hour"
			case elapsed >= 86400 && elapsed < 604800:
				t = strconv.FormatFloat(elapsed/60/60/24, 'f', 0, 64)
				u = "day"
			case elapsed >= 604800 && elapsed < 2419200:
				t = strconv.FormatFloat(elapsed/60/60/24/7, 'f', 0, 64)
				u = "week"
			case elapsed >= 2419200:
				t = strconv.FormatFloat(elapsed/60/60/24/7/4, 'f', 0, 64)
				u = "month"
		}
		if t == "1" {
			return "About " + t + " " + u + " ago"
		}
		return "About " + t + " " + u + "s ago"
	}
	return "More than a year ago"
}

// Helper: View wrapper for html safe
func htmlSafe(text string) template.HTML {
	return template.HTML(text)
}
