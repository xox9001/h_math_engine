package skiplist

import (
	"sync"
)

//类型、常量、结构体
const (
	SKIPLIST_MAX_LEVEL int = 25
	//SKIPLIST_P float32 = .5
	SKIPLIST_P int = 32767
)

type SkiplistNode struct {
	Value interface{}
	Forward []*SkiplistNode
	PreNode *SkiplistNode
	mutex sync.Mutex
}

type SkiplistType interface {
	Compare(value interface{},value2 interface{})(int)
	Dup(value interface{})(interface{},bool)
	Free(value interface{})(bool)
}

type Skiplist struct {
	Level int
	Len int64
	Type SkiplistType
	Header *SkiplistNode
	mutex sync.Mutex
}


type SkiplistIter struct {
	next *SkiplistNode
	mutex sync.Mutex
}



