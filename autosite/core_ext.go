/*
    Package autosite provides a simple infrastructure for running a
    personal website (off of the Google App Engine)
    
    Created by Ulf Möhring <hello@ulfmoehring.net>
*/

package autosite

// Look for a string in string array and return position
func LookFor(haystack []string, needle string) int {
	for i := 0; i < len(haystack); i++ {
		if haystack[i] == needle {
			return i
		}
	}
	return len(haystack)
}
