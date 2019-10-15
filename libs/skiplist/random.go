package skiplist

import "math/rand"

func SkiplistGenLevel() (int){
	var level int = 1

	for {
		r := rand.Int()
		if ( r & 65535 ) < SKIPLIST_P {
			level +=1
		}else{
			break
		}
	}

	if level > SKIPLIST_MAX_LEVEL {
		level = SKIPLIST_MAX_LEVEL
	}

	return level
}

