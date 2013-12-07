package starship

import (
	"fmt"
	"net"
	"strconv"

	"github.com/LSFN/lsfn"
)

type NebulaClient struct {
	starshipID string
	conn       *net.TCPConn
	gameID     string
}

// Perform a joining handshake with the Nebula
// We don't change the client struct's values until we know that the join has succeeded
func (client *NebulaClient) Join(host string, port int) bool {
	var err error
	var raddr *net.TCPAddr
	raddr, err = net.ResolveTCPAddr("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		fmt.Println(1, err)
		return false
	}
	fmt.Println("Address resolved")

	var conn *net.TCPConn
	conn, err = net.DialTCP("tcp", nil, raddr)
	if err != nil {
		fmt.Println(2, err)
		return false
	}
	fmt.Println("Connection made")

	// Receive the join info message
	joinInfo := new(lsfn.STSdown)
	err = lsfn.ReceiveSingleMessage(conn, joinInfo)
	if err != nil {
		conn.Close()
		fmt.Println(3, err)
		return false
	}
	currentGameID := joinInfo.GetJoinInfo().GetGameIDtoken()
	allowingNewClients := joinInfo.GetJoinInfo().GetAllowJoin()
	fmt.Println("Received join info:", currentGameID, allowingNewClients)

	// Determine whether to connect as a new client, rejoin or abandon the Nebula
	var join = lsfn.STSup_JoinRequest_JOIN
	var rejoin = lsfn.STSup_JoinRequest_REJOIN
	if client.gameID == currentGameID {
		lsfn.SendSingleMessage(client.conn, &lsfn.STSup{
			JoinRequest: &lsfn.STSup_JoinRequest{
				Type:        &rejoin,
				RejoinToken: &client.starshipID,
			},
		})
	} else {
		if allowingNewClients {
			lsfn.SendSingleMessage(client.conn, &lsfn.STSup{
				JoinRequest: &lsfn.STSup_JoinRequest{
					Type: &join,
				},
			})
		} else {
			conn.Close()
			fmt.Println(4, err)
			return false
		}
	}
	fmt.Println("Sent join request")

	// Receive the response
	joinResponse := new(lsfn.STSdown)
	err = lsfn.ReceiveSingleMessage(conn, joinResponse)
	if err != nil {
		conn.Close()
		fmt.Println(5, err)
		return false
	}
	fmt.Println("Received join response")

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
		fmt.Println(6, err)
		return false
	}
}
