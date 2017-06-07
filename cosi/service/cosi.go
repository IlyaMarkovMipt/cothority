package service

import (
	"errors"
	"fmt"
	"hash"
	"time"

	"github.com/satori/go.uuid"
	"gopkg.in/dedis/cothority.v2/cosi/protocol"
	"gopkg.in/dedis/kyber.v1"
	khash "gopkg.in/dedis/kyber.v1/util/hash"
	"gopkg.in/dedis/onet.v2"
	"gopkg.in/dedis/onet.v2/log"
	"gopkg.in/dedis/onet.v2/network"
)

// This file contains all the code to run a CoSi service. It is used to reply to
// client request for signing something using CoSi.
// As a prototype, it just signs and returns. It would be very easy to write an
// updated version that chains all signatures for example.

// ServiceName is the name to refer to the CoSi service
const ServiceName = "CoSi"

func init() {
	onet.RegisterNewService(ServiceName, newCoSiService)
	network.RegisterMessage(&SignatureRequest{})
	network.RegisterMessage(&SignatureResponse{})
}

type Suite interface {
	kyber.Group
	Hash() hash.Hash
}

// CoSi is the service that handles collective signing operations
type CoSi struct {
	*onet.ServiceProcessor
	suite Suite
}

// SignatureRequest is what the Cosi service is expected to receive from clients.
type SignatureRequest struct {
	Message []byte
	Roster  *onet.Roster
}

// SignatureResponse is what the Cosi service will reply to clients.
type SignatureResponse struct {
	Hash      []byte
	Signature []byte
}

// SignatureRequest treats external request to this service.
func (cs *CoSi) SignatureRequest(req *SignatureRequest) (network.Message, onet.ClientError) {
	if req.Roster.ID.IsNil() {
		req.Roster.ID = onet.RosterID(uuid.NewV4())
	}

	_, root := req.Roster.Search(cs.ServerIdentity().ID)
	tree := req.Roster.GenerateNaryTreeWithRoot(2, root)
	tni := cs.NewTreeNodeInstance(tree, tree.Root, cosi.Name)
	pi, err := cosi.NewProtocol(tni)
	if err != nil {
		return nil, onet.NewClientErrorCode(4100, "Couldn't make new protocol: "+err.Error())
	}
	cs.RegisterProtocolInstance(pi)
	pcosi := pi.(*cosi.CoSi)
	pcosi.SigningMessage(req.Message)
	h, err := khash.Bytes(cs.suite.Hash(), req.Message)
	if err != nil {
		return nil, onet.NewClientErrorCode(4101, "Couldn't hash message: "+err.Error())
	}
	response := make(chan []byte)
	pcosi.RegisterSignatureHook(func(sig []byte) {
		response <- sig
	})
	log.Lvl3("Cosi Service starting up root protocol")
	go pi.Dispatch()
	go pi.Start()
	sig := <-response
	if log.DebugVisible() > 1 {
		fmt.Printf("%s: Signed a message.\n", time.Now().Format("Mon Jan 2 15:04:05 -0700 MST 2006"))
	}
	return &SignatureResponse{
		Hash:      h,
		Signature: sig,
	}, nil
}

// NewProtocol is called on all nodes of a Tree (except the root, since it is
// the one starting the protocol) so it's the Service that will be called to
// generate the PI on all others node.
func (cs *CoSi) NewProtocol(tn *onet.TreeNodeInstance, conf *onet.GenericConfig) (onet.ProtocolInstance, error) {
	log.Lvl3("Cosi Service received New Protocol event")
	pi, err := cosi.NewProtocol(tn)
	return pi, err
}

func newCoSiService(c *onet.Context, s interface{}) (onet.Service, error) {
	suite, ok := s.(Suite)
	if !ok {
		return nil, errors.New("cosi: invalid suite given")
	}
	service := &CoSi{
		ServiceProcessor: onet.NewServiceProcessor(c, suite),
		suite:            suite,
	}
	err := service.RegisterHandler(service.SignatureRequest)
	if err != nil {
		return nil, err
	}
	return service, nil
}
