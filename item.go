package main

import (
	"bytes"
	"strings"
)

// Item represents base structure of elements like link, video and so on.
type Item struct {
	ID          string   `json:"id"`
	userID      string   `json:"userId"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	URL         string   `json:"url"`
}

// Link link type structure.
type Link struct {
	Item
}

// Video video type structure.
type Video struct {
	Item
}

// ParseLink extract information about link from []string, depend on simple rules.
// - if element is [video], then item type is video;
// - if element starts from http(s), element represents url of item;
// - if element starts from #, element is a tag;
// - if element is not identified as a type, a tag or a url, then it's a part of description.
func ParseLink(args []string) (interface{}, error) {
	url := ""
	tags := []string{}
	var description bytes.Buffer
	linkType := "link"
	argsLen := len(args)
	for i := 0; i < len(args); i++ {
		switch {
		case strings.HasPrefix(args[i], "http://"), strings.HasPrefix(args[i], "https://"):
			url = args[i]
		case strings.HasPrefix(args[i], "#"):
			tags = append(tags, args[i][1:])
		case strings.HasPrefix(args[i], "[") && strings.HasSuffix(args[i], "]"):
			linkType = args[i][1 : argsLen-1]
		default:
			if description.Len() > 0 {
				description.WriteString(" ")
			}
			description.WriteString(args[i])
		}

	}

	switch linkType {
	case "video":
		item := &Video{}
		item.URL = url
		item.Description = description.String()
		item.Tags = tags

		return item, nil
	default:
		item := &Link{}
		item.URL = url
		item.Description = description.String()
		item.Tags = tags

		return item, nil
	}

	return nil, nil
}
