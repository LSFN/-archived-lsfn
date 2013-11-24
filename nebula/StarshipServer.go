package nebula

import (
	"fmt"
	"net"

	"code.google.com/p/uuid"

	"lsfn/common"
)

type StarshipServer struct {
	unjoinedClients  map[*StarshipListener]int
	orphanShipIDs    map[string]int
	joinedClients    map[string]*StarshipListener
	networkChannels  map[string]chan *common.STSup
	listenConnection *net.TCPListener
	gameID           string
	allowJoin        bool
}

func (s *StarshipServer) handleConnectingStarship(conn *net.TCPConn) {
	conn.SetKeepAlive(true)
	starship := &StarshipListener{conn: conn}
	s.unjoinedClients[starship] = 1
	shipID := starship.Handshake(s.gameID, s.allowJoin, s.orphanShipIDs)
	if shipID != "" {
		delete(s.unjoinedClients, starship)
		s.joinedClients[shipID] = starship
		fmt.Println("Client with id " + shipID + " joined successfully")
		go starship.Listen()
	} else {
		delete(s.unjoinedClients, starship)
		starship.Disconnect()
	}
}

func (s *StarshipServer) Listen() {
	var err error
	conn, err := net.Listen("tcp", ":39461")
	s.listenConnection = conn.(*net.TCPListener)
	if err != nil {
		return
	}
	s.gameID = uuid.New()
	s.allowJoin = true
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

func (s *StarshipServer) processIncomingMessages() {

}
