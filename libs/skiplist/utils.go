package skiplist

//结构体方法
func (list *Skiplist)SkiplistGetIter()(*SkiplistIter){

	return &SkiplistIter{
		next: list.Header.Forward[0],
	}
}


func (iter *SkiplistIter)SkiplistNext()(*SkiplistNode){

	var curr *SkiplistNode = iter.next

	if curr != nil {
		iter.next = curr.Forward[0]
	}

	return curr
}

//释放链表
func (list *Skiplist)Release(){
	//TODO: 暂时处理
	list.Header = nil
}

func (iter *SkiplistIter)Release(){
	iter.next = nil
}


//创建节点
func CreateSkiplistNode(level int,value interface{}) (*SkiplistNode){

	return &SkiplistNode{
		Value: value,
		PreNode: nil,
		Forward: make([]*SkiplistNode,level,level*2),
	}
}

//创建跳跃表
func NewSkiplist(list_type SkiplistType)(*Skiplist){

	return &Skiplist{
		Len:1,
		Header: CreateSkiplistNode(SKIPLIST_MAX_LEVEL,nil),
		Level:1,
		Type: list_type,
	}

}

//插入节点
func (list *Skiplist)InsertNode(value interface{}) bool {

	list.mutex.Lock()
	defer list.mutex.Unlock()

	update := make([]*SkiplistNode, SKIPLIST_MAX_LEVEL, SKIPLIST_MAX_LEVEL*2)
	node := list.Header

	for i := list.Level - 1; i >= 0; i-- {

		for {

			if node.Forward[i] != nil && list.Type.Compare(node.Forward[i].Value, value) <= 0 {
					node = node.Forward[i]
			} else {
				break;
			}
		}

		update[i] = node
	}

	//如果价格存在但是订单数据一致，则不再插入
	if (node.Value != nil && list.Type.Compare(node.Value,value) == 0 ) {
		//TODO:判断重复数据
		return false
	}

	level := SkiplistGenLevel()

	if level > list.Level {

		for i:=list.Level;i< level;i++{
			update[i] = list.Header
		}
		list.Level = level
	}

	node = CreateSkiplistNode(level,value)

	for i:=0;i< level;i++{
		node.Forward[i] = update[i].Forward[i]
		update[i].Forward[i] = node
	}

	//node.PreNode = update[0]
	list.Len+=1

	return true
}

func (list *Skiplist)FindNode(value interface{})(*SkiplistNode){

	list.mutex.Lock()
	defer list.mutex.Unlock()

	node := list.Header

	for i:=list.Level-1;i>=0;i--{
		for {
			if node.Forward[i] != nil && list.Type.Compare(node.Value,value) <=0 {
				node = node.Forward[i]
			}else{
				break
			}
		}
	}

	if node.Value != nil && list.Type.Compare(node.Value,value) == 0 {
		return node
	}

	return nil
}

func (list *Skiplist)DeleteNode(value *SkiplistNode)(bool){

	list.mutex.Lock()

	update := make([]*SkiplistNode,SKIPLIST_MAX_LEVEL,SKIPLIST_MAX_LEVEL*2)
	node := list.Header


	for i := list.Level - 1; i >= 0; i-- {

		for {

			if node.Forward[i] != nil && list.Type.Compare(node.Forward[i].Value, value.Value) < 0 {
				node = node.Forward[i]
			} else {
				break;
			}
		}

		update[i] = node
	}

	for i:=0;i<list.Level;i++{

		if update[i].Forward[i] == value {
			update[i].Forward[i] = value.Forward[i]
		}

	}

	for {
		if list.Level > 1 && list.Header.Forward[list.Level-1] == nil{
			list.Level -= 1
		}else{
			break
		}
	}

	list.Type.Free(value.Value)
	value = nil
	list.Len -= 1
	list.mutex.Unlock()

	return true
}