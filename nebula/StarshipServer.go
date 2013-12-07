package nebula

import (
	"fmt"
	"net"

	"code.google.com/p/uuid"

	"github.com/LSFN/lsfn"
)

type StarshipServer struct {
	isListening      bool
	unjoinedClients  map[*StarshipListener]int
	orphanShipIDs    map[string]int
	joinedClients    map[string]*StarshipListener
	networkChannels  map[string]chan *lsfn.STSup
	listenConnection net.Listener
	gameID           string
	allowJoin        bool
}

func NewStarshipServer() *StarshipServer {
	s := new(StarshipServer)
	s.unjoinedClients = make(map[*StarshipListener]int)
	s.orphanShipIDs = make(map[string]int)
	s.joinedClients = make(map[string]*StarshipListener)
	s.networkChannels = make(map[string]chan *lsfn.STSup)
	return s
}

func (s *StarshipServer) handleConnectingStarship(conn net.Conn) {
	fmt.Println("New client is joining")
	starship := NewStarshipListener(conn)
	s.unjoinedClients[starship] = 1
	shipID, err := starship.Handshake(s.gameID, s.allowJoin, s.orphanShipIDs)
	if shipID == "" {
		if err != nil {
			fmt.Println(err)
		}
		delete(s.unjoinedClients, starship)
		starship.Disconnect()
	} else {
		delete(s.unjoinedClients, starship)
		s.joinedClients[shipID] = starship
		fmt.Println("Client with id " + shipID + " joined successfully")
		go starship.Listen()
	}
}

func (s *StarshipServer) Listen() {
	conn, err := net.Listen("tcp", ":39461")
	s.listenConnection = conn
	if err != nil {
		return
	}
	s.gameID = uuid.New()
	s.allowJoin = true
	s.isListening = true
	for {
		conn, err := s.listenConnection.Accept()
		if err != nil {
			s.shutDown()
			break
		}
		s.handleConnectingStarship(conn.(*net.TCPConn))
	}
}

func (s *StarshipServer) shutDown() {
	s.listenConnection.Close()
	for client := range s.unjoinedClients {
		client.Disconnect()
	}

}

func (s *StarshipServer) Listening() bool {
	return s.isListening
}

func (s *StarshipServer) processIncomingMessages() {

}
