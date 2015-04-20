package kademlia

import (
	"container/list"
)

// define a Bucket
type Bucket struct {
	Contacts *list.List
}

// construct function
func BuildBucket() *Bucket {
	ptr := new(Bucket)
	ptr.Contacts = list.New()
	return ptr
}

func (B Bucket) FindContact(contact *Contact) (res bool, node *list.Element) {
	res = false
	for el := B.Contacts.Front(); el != nil; el = el.Next() {
		if contact == el.Value.(*Contact) {
			res = true
			node = el
			return
		}
	}
	return
}

func (B Bucket) FindById(nodeId ID) (*Contact, error) {
	for el := B.Contacts.Front(); el != nil; el = el.Next() {
		if nodeId == el.Value.(*Contact).NodeID {
			return el.Value.(*Contact), nil
		}
	}
	return nil, &NotFoundError{nodeId, "Not found"}
}

func (B Bucket) IsFull() bool {
	if B.Contacts.Len() == k {
		return true
	}
	return false
}
