package treemux

import (
	"fmt"
	"net/url"
	"strings"
)

type node struct {
	path string

	priority int

	// The list of static children to check.
	staticIndices []byte
	staticChild   []*node

	// If none of the above match, check the wildcard children
	wildcardChild *node

	// If none of the above match, then we use the catch-all, if applicable.
	catchAllChild *node

	// Data for the node is below.

	addSlash   bool
	isCatchAll bool
	// If this node is the end of the URL, then call the handler, if applicable.
	leafValue interface{}

	// The names of the parameters to apply.
	leafWildcardNames []string
}

func (n *node) sortStaticChild(i int) {
	for i > 0 && n.staticChild[i].priority > n.staticChild[i-1].priority {
		n.staticChild[i], n.staticChild[i-1] = n.staticChild[i-1], n.staticChild[i]
		n.staticIndices[i], n.staticIndices[i-1] = n.staticIndices[i-1], n.staticIndices[i]
		i -= 1
	}
}

func (n *node) setValue(v interface{}) {
	if n.leafValue != nil {
		panic(fmt.Errorf("treemux: duplicate value for path %s", n.path))
	}
	n.leafValue = v
}

func (n *node) addPath(path string, wildcards []string) *node {
	leaf := len(path) == 0
	if leaf {
		if wildcards != nil {
			// Make sure the current wildcards are the same as the old ones.
			// If not then we have an ambiguous path.
			if n.leafWildcardNames != nil {
				if len(n.leafWildcardNames) != len(wildcards) {
					// This should never happen.
					panic("Reached leaf node with differing wildcard array length. Please report this as a bug.")
				}

				for i := 0; i < len(wildcards); i++ {
					if n.leafWildcardNames[i] != wildcards[i] {
						panic(fmt.Sprintf("Wildcards %v are ambiguous with wildcards %v",
							n.leafWildcardNames, wildcards))
					}
				}
			} else {
				// No wildcards yet, so just add the existing set.
				n.leafWildcardNames = wildcards
			}
		}

		return n
	}

	c := path[0]
	nextSlash := strings.Index(path, "/")
	var thisToken string
	var tokenEnd int

	if c == '/' {
		thisToken = "/"
		tokenEnd = 1
	} else if nextSlash == -1 {
		thisToken = path
		tokenEnd = len(path)
	} else {
		thisToken = path[0:nextSlash]
		tokenEnd = nextSlash
	}
	remainingPath := path[tokenEnd:]

	if c == '*' {
		// Token starts with a *, so it's a catch-all
		thisToken = thisToken[1:]
		if n.catchAllChild == nil {
			n.catchAllChild = &node{path: thisToken, isCatchAll: true}
		}

		if path[1:] != n.catchAllChild.path {
			panic(fmt.Sprintf("Catch-all name in %s doesn't match %s",
				path, n.catchAllChild.path))
		}

		if nextSlash != -1 {
			panic("/ after catch-all found in " + path)
		}

		if wildcards == nil {
			wildcards = []string{thisToken}
		} else {
			wildcards = append(wildcards, thisToken)
		}
		n.catchAllChild.leafWildcardNames = wildcards

		return n.catchAllChild
	} else if c == ':' {
		// Token starts with a :
		thisToken = thisToken[1:]

		if wildcards == nil {
			wildcards = []string{thisToken}
		} else {
			wildcards = append(wildcards, thisToken)
		}

		if n.wildcardChild == nil {
			n.wildcardChild = &node{path: "wildcard"}
		}

		return n.wildcardChild.addPath(remainingPath, wildcards)

	} else {
		if strings.ContainsAny(thisToken, ":*") {
			panic("* or : in middle of path component " + path)
		}

		// Do we have an existing node that starts with the same letter?
		for i, index := range n.staticIndices {
			if c == index {
				// Yes. Split it based on the common prefix of the existing
				// node and the new one.
				child, prefixSplit := n.splitCommonPrefix(i, thisToken)
				child.priority++
				n.sortStaticChild(i)
				return child.addPath(path[prefixSplit:], wildcards)
			}
		}

		// No existing node starting with this letter, so create it.
		child := &node{path: thisToken}

		if n.staticIndices == nil {
			n.staticIndices = []byte{c}
			n.staticChild = []*node{child}
		} else {
			n.staticIndices = append(n.staticIndices, c)
			n.staticChild = append(n.staticChild, child)
		}
		return child.addPath(remainingPath, wildcards)
	}
}

func (n *node) splitCommonPrefix(existingNodeIndex int, path string) (*node, int) {
	childNode := n.staticChild[existingNodeIndex]

	if strings.HasPrefix(path, childNode.path) {
		// No split needs to be done. Rather, the new path shares the entire
		// prefix with the existing node, so the new node is just a child of
		// the existing one. Or the new path is the same as the existing path,
		// which means that we just move on to the next token. Either way,
		// this return accomplishes that
		return childNode, len(childNode.path)
	}

	var i int
	// Find the length of the common prefix of the child node and the new path.
	for i = range childNode.path {
		if i == len(path) {
			break
		}
		if path[i] != childNode.path[i] {
			break
		}
	}

	commonPrefix := path[0:i]
	childNode.path = childNode.path[i:]

	// Create a new intermediary node in the place of the existing node, with
	// the existing node as a child.
	newNode := &node{
		path:     commonPrefix,
		priority: childNode.priority,
		// Index is the first letter of the non-common part of the path.
		staticIndices: []byte{childNode.path[0]},
		staticChild:   []*node{childNode},
	}
	n.staticChild[existingNodeIndex] = newNode

	return newNode, i
}

func (n *node) search(path string) (found *node, params []string) {
	// if test != nil {
	// 	test.Logf("Searching for %s in %s", path, n.dumpTree("", ""))
	// }
	pathLen := len(path)
	if pathLen == 0 {
		if n.leafValue == nil {
			return nil, nil
		} else {
			return n, nil
		}
	}

	// First see if this matches a static token.
	firstChar := path[0]
	for i, staticIndex := range n.staticIndices {
		if staticIndex == firstChar {
			child := n.staticChild[i]
			childPathLen := len(child.path)
			if pathLen >= childPathLen && child.path == path[:childPathLen] {
				nextPath := path[childPathLen:]
				found, params = child.search(nextPath)
			}
			break
		}
	}

	if found != nil {
		return
	}

	if n.wildcardChild != nil {
		// Didn't find a static token, so check for a wildcard.
		nextSlash := 0
		for nextSlash < pathLen && path[nextSlash] != '/' {
			nextSlash++
		}

		thisToken := path[0:nextSlash]
		nextToken := path[nextSlash:]

		if len(thisToken) > 0 { // Don't match on empty tokens.
			found, params = n.wildcardChild.search(nextToken)
			if found != nil {
				unescaped, err := url.QueryUnescape(thisToken)
				if err != nil {
					unescaped = thisToken
				}

				if params == nil {
					params = []string{unescaped}
				} else {
					params = append(params, unescaped)
				}

				return
			}
		}
	}

	catchAllChild := n.catchAllChild
	if catchAllChild != nil {
		// Hit the catchall, so just assign the whole remaining path.
		unescaped, err := url.QueryUnescape(path)
		if err != nil {
			unescaped = path
		}

		return catchAllChild, []string{unescaped}
	}

	return nil, nil
}

func (n *node) dumpTree(prefix, nodeType string) string {
	line := fmt.Sprintf("%s %02d %s%s [%d] %v wildcards %v\n", prefix, n.priority, nodeType, n.path,
		len(n.staticChild), n.leafValue, n.leafWildcardNames)
	prefix += "  "
	for _, node := range n.staticChild {
		line += node.dumpTree(prefix, "")
	}
	if n.wildcardChild != nil {
		line += n.wildcardChild.dumpTree(prefix, ":")
	}
	if n.catchAllChild != nil {
		line += n.catchAllChild.dumpTree(prefix, "*")
	}
	return line
}
