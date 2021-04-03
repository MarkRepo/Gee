package gee

import (
	"log"
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node       // roots key eg, roots['GET'] roots['POST']
	handlers map[string]HandlerFunc // handlers key eg, handlers['GET-/p/:lang/doc'], handlers['POST-/p/book']
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

// parsePattern only one * is allowed
func parsePattern(pattern string) []string {
	items := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, item := range items {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

func (r *router) addRoute(method, pattern string, handler HandlerFunc) {
	log.Printf("Route %4s - %s", method, pattern)
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	if _, ok := r.roots[method]; !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

// 解析了:和*两种匹配符的参数，返回一个 map
func (r *router) getRoute(method, path string) (*node, map[string]string) {
	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}

	searchParts := parsePattern(path)
	n := root.search(searchParts, 0)
	if n == nil {
		return nil, nil
	}

	params := make(map[string]string)
	parts := parsePattern(n.pattern)
	for i, part := range parts {
		if part[0] == ':' && len(part) > 1 {
			params[part[1:]] = searchParts[i]
		}
		if part[0] == '*' && len(part) > 1 {
			params[part[1:]] = strings.Join(searchParts[i:], "/")
			break
		}
	}
	return n, params
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	c.Next()
}
