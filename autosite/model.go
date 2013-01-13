/*
    Package autosite provides a simple infrastructure for running a
    personal website (off of the Google App Engine)
    
    Created by Ulf Möhring <hello@ulfmoehring.net>
*/

package autosite


type Model interface {
	Type() string
}

func FormKey(m Model, attr string) string {
	return m.Type() + "[" + attr + "]"
}
