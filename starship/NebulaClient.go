package starship

import (
	"net"

	"code.google.com/p/goprotobuf/proto"

	"github.com/LSFN/lsfn"
)

type NebulaClient struct {
	starshipID string
	conn       *TCPConn
	gameID     string
}

// Perform a joining handshake with the Nebula
// We don't change the client struct's values until we know that the join has succeeded
func (client *NebulaClient) Join(host string, port int) bool {
	var err error
	var conn *net.TCPConn
	conn, err = net.Dial("tcp", host+":"+string(port))
	if err != nil {
		return false
	}

	// Receive the join info message
	joinInfo := new(lsfn.STSdown)
	err = lsfn.ReceiveSingleMessage(conn, joinInfo)
	if err != nil {
		conn.Close()
		return false
	}
	currentGameID := joinInfo.GetJoinInfo().GetGameIDtoken()
	allowingNewClients := joinInfo.GetJoinInfo().GetAllowJoin()

	// Determine whether to connect as a new client, rejoin or abandon the Nebula
	if client.gameID == currentGameID {
		lsfn.SendSingleMessage(&lsfn.STSup{
			JoinRequest: &lsfn.STSup_JoinRequest{
				Type: &lsfn.STSup_JoinRequest_REJOIN,
				RejoinToken: &client.starshipID
			},
		})
	} else {
		if allowingNewClients {
			lsfn.SendSingleMessage(&lsfn.STSup{
				JoinRequest: &lsfn.STSup_JoinRequest{
					Type: &lsfn.STSup_JoinRequest_JOIN,
				},
			})
		} else {
			conn.Close()
			return false
		}
	}

	// Receive the response
	joinResponse := new(lsfn.STSdown)
	err = lsfn.ReceiveSingleMessage(conn, joinResponse)
	if err != nil {
		conn.Close()
		return false
	}
	switch joinResponse.GetJoinResponse().GetType() {
		case lsfn.STSdown_JoinResponse_JOIN_ACCEPTED:
			client.gameID = currentGameID
			client.starshipID = joinResponse.GetJoinResponse().GetRejoinToken()
			client.conn = conn
			return true
		case lsfn.STSdown_JoinResponse_REJOIN_ACCEPTED:
			client.conn = conn
			return true
		case lsfn.STSdown_JoinResponse_JOIN_REJECTED:
			conn.Close()
			return false
	}
}
