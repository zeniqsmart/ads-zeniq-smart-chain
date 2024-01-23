package b

import (
	"github.com/zeniqsmart/ads-zeniq-smart-chain/indextree/b/cppbtree"
)

type Enumerator = cppbtree.Enumerator

type Tree = cppbtree.Tree

func TreeNew() *Tree {
	return cppbtree.TreeNew()
}
