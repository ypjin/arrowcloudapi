package utils

import (
	"strings"
)

// Link : HTTP link header
type Link struct {
	// URL : url part of header
	URL string
	// Rel : prev or next
	Rel string
}

// Links : multiple link
type Links []*Link

// Prev returns the URL indicated by "prev" rel.
func (l Links) Prev() string {
	prev := ""
	for _, link := range l {
		if strings.ToLower(link.Rel) == "prev" {
			prev = link.URL
			break
		}
	}
	return prev
}

// Next returns the URL indicated by "next" rel.
func (l Links) Next() string {
	next := ""
	for _, link := range l {
		if link.Rel == "next" {
			next = link.URL
			break
		}
	}
	return next
}

// ParseLink parses the raw link header to Links
func ParseLink(raw string) Links {
	links := Links{}

	for _, l := range strings.Split(raw, ",") {
		link := parseSingleLink(l)
		if link != nil {
			links = append(links, link)
		}
	}

	return links
}

func parseSingleLink(raw string) *Link {
	link := &Link{}

	for _, str := range strings.Split(raw, ";") {
		str = strings.TrimSpace(str)
		if strings.HasPrefix(str, "<") && strings.HasSuffix(str, ">") {
			str = strings.Trim(str, "<>")
			link.URL = str
			continue
		}

		parts := strings.SplitN(str, "=", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "rel" {
			continue
		}

		link.Rel = strings.ToLower(strings.Trim(parts[1], "\""))
	}

	if len(link.URL) == 0 || len(link.Rel) == 0 {
		link = nil
	}

	return link
}
