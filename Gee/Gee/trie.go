package gee

import "strings"

// 所谓动态路由，即一条路由规则可以匹配某一类型而非某一条固定的路由.
// 实现动态路由最常用的数据结构，被称为前缀树(Trie树),每一个节点的所有的子节点都拥有相同的前缀
// 我们实现的动态路由具有以下功能：
// 1. 参数匹配:。例如 /p/:lang/doc，可以匹配 /p/c/doc 和 /p/go/doc。
// 2. 通配*。例如 /static/*filepath，可以匹配/static/fav.ico，
// 也可以匹配/static/js/jQuery.js，这种模式常用于静态服务器，能够递归地匹配子路径

type node struct {
	pattern  string  // 待匹配路由
	part     string  // 路由中的一部分
	children []*node // 所有子节点
	isWild   bool    // 是否通配，part含有 : 或 * 时为 true
}

// 第一个匹配成功的节点，用于插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 所有匹配成功的节点，用于查询
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// 递归查找每一层的节点，如果没有匹配到当前part的节点，则新建一个，有一点需要注意，/p/:lang/doc只有在第三层节点，即doc节点，
// pattern才会设置为/p/:lang/doc。p和:lang节点的pattern属性皆为空。因此，当匹配结束时，
// 我们可以使用n.pattern == ""来判断路由规则是否匹配成功
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{
			part:   part,
			isWild: part[0] == '*' || part[0] == ':',
		}
		n.children = append(n.children, child)
	}

	child.insert(pattern, parts, height+1)
}

// 查询功能，同样也是递归查询每一层的节点，退出规则是: 匹配到了*，匹配失败，或者匹配到了第len(parts)层节点
func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)
	for _, child := range children {
		if result := child.search(parts, height+1); result != nil {
			return result
		}
	}

	return nil
}
