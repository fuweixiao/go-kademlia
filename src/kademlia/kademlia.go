package kademlia

// Contains the core kademlia type. In addition to core state, this type serves
// as a receiver for the RPC methods, which is required by that package.

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

const (
	alpha = 3
	b     = 8 * IDBytes
	k     = 20
)

// Kademlia type. You can put whatever state you need in this.
type Kademlia struct {
	NodeID        ID
	SelfContact   Contact
	Buckets       []Bucket
	UpdateChannel chan *Contact
	Data          map[ID][]byte
}

func NewKademlia(laddr string) *Kademlia {
	// TODO: Initialize other state here as you add functionality.
	k := new(Kademlia)
	k.NodeID = NewRandomID()
	k.Buckets = make([]Bucket, b, b)
	k.UpdateChannel = make(chan *Contact)
	k.Data = make(map[ID][]byte)
	for i := 0; i < b; i++ {
		k.Buckets[i] = *(BuildBucket())
	}

	// Set up RPC server
	// NOTE: KademliaCore is just a wrapper around Kademlia. This type includes
	// the RPC functions.
	rpc.Register(&KademliaCore{k})
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal("Listen: ", err)
	}
	// Run RPC server forever.
	go http.Serve(l, nil)

	// Add self contact
	hostname, port, _ := net.SplitHostPort(l.Addr().String())
	port_int, _ := strconv.Atoi(port)
	ipAddrStrings, err := net.LookupHost(hostname)
	var host net.IP
	for i := 0; i < len(ipAddrStrings); i++ {
		host = net.ParseIP(ipAddrStrings[i])
		if host.To4() != nil {
			break
		}
	}
	k.SelfContact = Contact{k.NodeID, host, uint16(port_int)}

	// Run a go routine to update KBuckets
	go k.UpdateBucket(k.UpdateChannel)

	return k
}

type NotFoundError struct {
	id  ID
	msg string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%x %s", e.id, e.msg)
}

func (k *Kademlia) FindContact(nodeId ID) (*Contact, error) {
	// TODO: Search through contacts, find specified ID
	// Find contact with provided ID
	if nodeId == k.SelfContact.NodeID {
		return &k.SelfContact, nil
	}
	dist := nodeId.Xor(k.SelfContact.NodeID)
	BucketId := dist.PrefixLen()
	return k.Buckets[BucketId].FindById(nodeId)
}

// This is the function to perform the RPC
func (k *Kademlia) DoPing(host net.IP, port uint16) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	pong, err := k.InternalDoPing(host, port, true)
	if err != nil {
		return "ERROR"
	} else {
		return "OK: " + pong.MsgID.AsString()
	}
}

func (k *Kademlia) DoStore(contact *Contact, key ID, value []byte) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"

	//build store request
	req := StoreRequest{k.SelfContact, NewRandomID(), key, value}
	var res StoreResult

	//DialHTTP and call RPC
	client, err := rpc.DialHTTP("tcp", contact.Host.String()+":"+strconv.FormatInt(int64(contact.Port), 10))
	if err != nil {
		log.Fatal("DialHTTP: ", err)
	}
	err = client.Call("KademliaCore.Store", req, &res)
	if err != nil {
		log.Fatal("Call: ", err)
		return "ERROR"
	} else {
		return "OK: " + res.MsgID.AsString()
	}
}

func (k *Kademlia) DoFindNode(contact *Contact, searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	req := FindNodeRequest{k.SelfContact, NewRandomID(), searchKey}
	var res FindNodeResult
	client, err := rpc.DialHTTP("tcp", contact.Host.String()+":"+strconv.FormatInt(int64(contact.Port), 10))
	if err != nil {
		log.Fatal("DialHTTP: ", err)
	}
	err = client.Call("KademliaCore.FindNode", req, &res)
	if err != nil {
		log.Fatal("Call: ", err)
		return "ERROR"
	} else {
		return "OK: " + res.MsgID.AsString()
	}
}

func (k *Kademlia) DoFindValue(contact *Contact, searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	req := FindValueRequest{k.SelfContact, NewRandomID(), searchKey}
	var res FindValueResult
	client, err := rpc.DialHTTP("tcp", contact.Host.String()+":"+strconv.FormatInt(int64(contact.Port), 10))
	if err != nil {
		log.Fatal("DialHTTP: ", err)
	}
	err = client.Call("KademliaCore.FindValue", req, &res)
	if err != nil {
		log.Fatal("Call: ", err)
		return "ERROR"
	} else {
		return "OK: " + res.MsgID.AsString()
	}
}

func (k *Kademlia) LocalFindValue(searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	if value, ok := k.Data[searchKey]; ok {
		return "OK: " + string(value)
	} else {
		return "ERROR"
	}
}

func (k *Kademlia) DoIterativeFindNode(id ID) string {
	// For project 2!
	return "ERR: Not implemented"
}
func (k *Kademlia) DoIterativeStore(key ID, value []byte) string {
	// For project 2!
	return "ERR: Not implemented"
}
func (k *Kademlia) DoIterativeFindValue(key ID) string {
	// For project 2!
	return "ERR: Not implemented"
}

func (k *Kademlia) InternalDoPing(host net.IP, port uint16, update bool) (PongMessage, error) {
	// real function to Do ping

	// build ping messsage
	ping := PingMessage{k.SelfContact, NewRandomID()}
	var pong PongMessage

	//dial client
	client, err := rpc.DialHTTP("tcp", host.String()+":"+strconv.FormatInt(int64(port), 10))
	if err != nil {
		log.Fatal("DialHTTP: ", err)
	}
	err = client.Call("KademliaCore.Ping", ping, &pong)
	if err != nil {
		log.Fatal("Call: ", err)
	} else {
		if update {
			k.UpdateChannel <- &pong.Sender // do update
		} else {
			// do nothing
		}
	}
	return pong, nil
}

func (k *Kademlia) UpdateBucket(UpdateChannel chan *Contact) {
	for {
		select {
		case contact := <-UpdateChannel:
			BucketIdx := k.NodeID.Xor(contact.NodeID).PrefixLen()
			if BucketIdx == 160 {
				break
			}
			Contacts := k.Buckets[BucketIdx].Contacts
			if flag, node := k.Buckets[BucketIdx].FindContact(contact); flag {
				Contacts.MoveToBack(node)
			} else if k.Buckets[BucketIdx].IsFull() {
				FirstContact := Contacts.Front().Value.(*Contact)
				_, err := k.InternalDoPing(FirstContact.Host, FirstContact.Port, false)
				if err != nil {
					Contacts.Remove(Contacts.Front())
					Contacts.PushBack(contact)
				} else {
					Contacts.MoveToBack(Contacts.Front())
				}
			} else {
				k.Buckets[BucketIdx].Contacts.PushBack(contact)
			}
		}
	}
}
