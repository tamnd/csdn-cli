package csdn

import "net/url"

// articleURL builds the canonical url for an article given its author username
// and numeric id.
func articleURL(username, id string) string {
	return blogHost + "/" + username + "/article/details/" + id
}

// userURL builds the canonical url for a profile given its username.
func userURL(username string) string {
	return blogHost + "/" + username
}

// searchURL builds the human search url for a query (the page a reader visits,
// distinct from the so.csdn.net JSON api the client calls).
func searchURL(q string) string {
	return soHost + "/so/search?q=" + url.QueryEscape(q)
}
