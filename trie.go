package gee

import (
	"fmt"
	"strings"
)

// Trie树
type node struct {
	pattern     string  // 待匹配的路由
	part        string  // 路由的一部分
	children    []*node // 子节点
	isEnd       bool    // 判断该节点是否是终结点
	childHasEnd bool    // 判断子节点中是否包含终终结点
	isWild      bool    // 是否精确匹配，包含 '*' 或 ':' 时为 true
	middlewares []HandlerFunc
}

// 查找子节点中第一个终结点，用于插入时路由冲突
func (n *node) findFirstEndChild() *node {
	for _, child := range n.children {
		if child.isEnd {
			return child
		}
	}
	return nil
}

// 第一个匹配的节点，用于插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 所有匹配的节点，用于查询
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		n.isEnd = true
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}

	if len(parts) == height+1 && n.childHasEnd && child.isWild {
		endNode := n.findFirstEndChild()
		panic(fmt.Sprintf("The new path '%s' is conflict with path '%s'", pattern, endNode.pattern))
	}

	if len(parts) == height+1 && !n.childHasEnd {
		n.childHasEnd = true
	}

	child.insert(pattern, parts, height+1)
}

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
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}

	return nil
}
