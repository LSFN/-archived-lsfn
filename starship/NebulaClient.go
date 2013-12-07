package starship

import (
	"fmt"
	"net"

	"github.com/LSFN/lsfn"
)

type NebulaClient struct {
	starshipID string
	conn       net.Conn
	gameID     string
}

// Perform a joining handshake with the Nebula
// We don't change the client struct's values until we know that the join has succeeded
func (client *NebulaClient) Join(host string, port int) bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return false
	}
	client.conn = conn

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
	var join = lsfn.STSup_JoinRequest_JOIN
	var rejoin = lsfn.STSup_JoinRequest_REJOIN
	if client.gameID == currentGameID {
		err = lsfn.SendSingleMessage(client.conn, &lsfn.STSup{
			JoinRequest: &lsfn.STSup_JoinRequest{
				Type:        &rejoin,
				RejoinToken: &client.starshipID,
			},
		})
		if err != nil {
			return false
		}
	} else {
		if allowingNewClients {
			err = lsfn.SendSingleMessage(client.conn, &lsfn.STSup{
				JoinRequest: &lsfn.STSup_JoinRequest{
					Type: &join,
				},
			})
			if err != nil {
				return false
			}
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
	default:
		//lsfn.STSdown_JoinResponse_JOIN_REJECTED:
		conn.Close()
		return false
	}
}
