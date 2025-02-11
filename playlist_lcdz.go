package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

func NewPlaylistLCDZ() (*Playlist, error) {
	urlSource := "https://livecode.demozoo.org/type/Byte_Jam.html"
	resp, err := http.Get(urlSource)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LCDZ page responded with [%s] rather than 200", resp.Status)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	p := NewPlaylist()
	items, err := findAllEntries(doc)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("LCDZ page responded, but no lua files found")
	}

	p.items = items
	p.order = ORDER_RANDOM

	return p, nil
}

/*
findAllEntries is now doing something a bit more sophisticated...
func findAllLuaLinks(n *html.Node) []string {
	var links []string
	re := regexp.MustCompile(`.lua$`)
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, element := range n.Attr {
			if element.Key == "href" {
				if re.MatchString(element.Val) {
					links = append(links, fmt.Sprintf("https://livecode.demozoo.org%s", element.Val))
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		newLinks := findAllLuaLinks(c)
		links = append(links, newLinks...)
	}

	return links
}
*/

func findAllEntries(n *html.Node) ([]PlaylistItem, error) {
	detailsNodes := getAllMatchingBelow(n, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "details"
	})

	var items []PlaylistItem
	for _, detailsNode := range detailsNodes {
		newItems, err := parseDetailsNode(detailsNode)
		if err != nil {
			return nil, err
		}
		items = append(items, newItems...)
	}

	return items, nil
}

func parseDetailsNode(n *html.Node) ([]PlaylistItem, error) {
	entryCardNodes := getAllMatchingBelow(n, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "div" && nodeContainsClass(n, "entry-card")
	})

	var items []PlaylistItem
	for _, entryCardNode := range entryCardNodes {
		author, link, err := parseEntryCardNode(entryCardNode)
		if err != nil {
			//return nil, err
			continue // Some fail - that's ok for now
		}
		items = append(items, PlaylistItem{
			location: link,
			author:   author,
		})
	}

	return items, nil
}

// parseEntryCardNode returns (author, link, error)
func parseEntryCardNode(n *html.Node) (string, string, error) {
	elName := getFirstMatchingBelow(n, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "div" && nodeContainsClass(n, "entry-name")
	})

	if elName == nil {
		return "", "", fmt.Errorf("no entry name found")
	}

	elInfo := getFirstMatchingBelow(n, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "table" && nodeContainsClass(n, "entry-info")
	})

	if elInfo == nil {
		return "", "", fmt.Errorf("no entry info found")
	}

	elLink := getFirstMatchingBelow(elInfo, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "a"
	})

	if elLink == nil {
		return "", "", fmt.Errorf("no entry link found")
	}

	href := getValueFromAttribute(elLink, "href")
	if href == "" {
		return "", "", fmt.Errorf("no entry link href found")
	}

	re := regexp.MustCompile(`.lua$`)
	if !re.MatchString(href) {
		return "", "", fmt.Errorf("link [%s] is not a lua file", href)
	}

	link := fmt.Sprintf("https://livecode.demozoo.org%s", href)

	return elName.FirstChild.Data, link, nil
}

type elMatcher func(*html.Node) bool

func getFirstMatchingBelow(n *html.Node, fnMatcher elMatcher) *html.Node {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		childMatch := getFirstMatchingSelfOrBelow(child, fnMatcher)
		if childMatch != nil {
			return childMatch
		}
	}
	return nil
}

func getAllMatchingBelow(n *html.Node, fnMatcher elMatcher) []*html.Node {
	nodes := []*html.Node{}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		childMatches := getAllMatchingSelfOrBelow(child, fnMatcher)
		nodes = append(nodes, childMatches...)
	}
	return nodes
}

// getFirstMatchingSelfOrBelow is recursive
// return may be nil if nothing can be found
func getFirstMatchingSelfOrBelow(n *html.Node, fnMatcher elMatcher) *html.Node {
	if fnMatcher(n) {
		return n // this one is a match!
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		childMatch := getFirstMatchingSelfOrBelow(child, fnMatcher)
		if childMatch != nil {
			return childMatch
		}
	}
	return nil
}

// getAllMatchingSelfOrBelow is recursive
// return may be empty if nothing can be found
func getAllMatchingSelfOrBelow(n *html.Node, fnMatcher elMatcher) []*html.Node {
	nodes := []*html.Node{}
	if fnMatcher(n) {
		nodes = append(nodes, n)
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		childMatches := getAllMatchingSelfOrBelow(child, fnMatcher)
		nodes = append(nodes, childMatches...)
	}
	return nodes
}

func nodeContainsClass(n *html.Node, className string) bool {
	for _, a := range n.Attr {
		// not sure strings.Contains is a super safe way, but it'll do us for now
		if a.Key == "class" && strings.Contains(a.Val, className) {
			return true
		}
	}
	return false
}

func getValueFromAttribute(n *html.Node, findKey string) string {
	for _, a := range n.Attr {
		if a.Key == findKey {
			return a.Val
		}
	}
	return ""
}
