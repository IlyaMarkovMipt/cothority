package handlers

import "gopkg.in/dedis/onet.v1"

// Announce is used to pass a message to all children.
type Announce struct {
	Message string
}

// StructAnnounce just contains Announce and the data necessary to identify and
// process the message in the onet framework.
type StructAnnounce struct {
	*onet.TreeNode
	Announce
}

// Reply returns the count of all children.
type Reply struct {
	ChildrenCount int
}

// StructReply just contains Reply and the data necessary to identify and
// process the message in the onet framework.
type StructReply struct {
	*onet.TreeNode
	Reply
}
