package csdn

// articleURL builds the canonical url for an article given its author username
// and numeric id.
func articleURL(username, id string) string {
	return blogHost + "/" + username + "/article/details/" + id
}

// userURL builds the canonical url for a profile given its username.
func userURL(username string) string {
	return blogHost + "/" + username
}
