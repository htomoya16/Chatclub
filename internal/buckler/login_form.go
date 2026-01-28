package buckler

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

func findLoginCallbackForm(body []byte) (string, map[string]string) {
	if len(body) == 0 {
		return "", nil
	}
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return "", nil
	}

	var form *html.Node
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "form" {
			if action := attr(n, "action"); action != "" && strings.Contains(action, "/login/callback") {
				form = n
				return
			}
		}
		for c := n.FirstChild; c != nil && form == nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	if form == nil {
		return "", nil
	}

	action := attr(form, "action")
	fields := map[string]string{}
	var collect func(n *html.Node)
	collect = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "input" {
			if strings.ToLower(attr(n, "type")) == "hidden" {
				name := attr(n, "name")
				if name != "" {
					fields[name] = attr(n, "value")
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			collect(c)
		}
	}
	collect(form)

	if len(fields) == 0 {
		return "", nil
	}
	return action, fields
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}
