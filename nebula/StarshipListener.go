package nebula

import (
	"errors"
	"net"

	"code.google.com/p/uuid"

	"github.com/LSFN/lsfn"
)

type StarshipListener struct {
	conn     net.Conn
	Messages chan *lsfn.STSup
}

func NewStarshipListener(conn net.Conn) *StarshipListener {
	listener := new(StarshipListener)
	listener.conn = conn
	listener.Messages = make(chan *lsfn.STSup)
	return listener
}

func (listener *StarshipListener) Listen() {
	for {
		message := new(lsfn.STSup)
		err := lsfn.ReceiveSingleMessage(listener.conn, message)
		if err != nil {
			listener.conn.Close()
			close(listener.Messages)
			break
		}
		listener.Messages <- message
	}
}

func (listener *StarshipListener) SendMessage(downMessage *lsfn.STSdown) {
	err := lsfn.SendSingleMessage(listener.conn, downMessage)
	if err != nil {
		listener.Disconnect()
	}
}

func (listener *StarshipListener) Disconnect() {
	listener.conn.Close()
	close(listener.Messages)
}

// TODO possibly disconnect
func (listener *StarshipListener) Handshake(gameID string, allowJoin bool, rejoinIDs map[string]int) (string, error) {
	// Send the JoinInfo message
	err := lsfn.SendSingleMessage(listener.conn, &lsfn.STSdown{
		JoinInfo: &lsfn.STSdown_JoinInfo{
			AllowJoin:   &allowJoin,
			GameIDtoken: &gameID,
		},
	})
	if err != nil {
		return "", err
	}

	// Receive the JoinRequest message
	joinRequestMessage := new(lsfn.STSup)
	err = lsfn.ReceiveSingleMessage(listener.conn, joinRequestMessage)
	if err != nil {
		return "", err
	}
	if joinRequestMessage.JoinRequest == nil {
		return "", errors.New("Received message does not have JoinRequest field set")
	}

	// Send the JoinResponse message
	joinType := joinRequestMessage.GetJoinRequest().GetType()
	var joinAccept = lsfn.STSdown_JoinResponse_JOIN_ACCEPTED
	var joinReject = lsfn.STSdown_JoinResponse_JOIN_REJECTED
	var rejoinAccept = lsfn.STSdown_JoinResponse_REJOIN_ACCEPTED
	if joinType == lsfn.STSup_JoinRequest_JOIN {
		if allowJoin {
			id := uuid.New()
			err = lsfn.SendSingleMessage(listener.conn, &lsfn.STSdown{
				JoinResponse: &lsfn.STSdown_JoinResponse{
					Type:        &joinAccept,
					RejoinToken: &id,
				},
			})
			if err != nil {
				return "", err
			}
			return id, nil
		} else {
			err = lsfn.SendSingleMessage(listener.conn, &lsfn.STSdown{
				JoinResponse: &lsfn.STSdown_JoinResponse{
					Type: &joinReject,
				},
			})
			if err != nil {
				return "", err
			}
			return "", nil
		}
	} else {
		var successfulRejoin bool = false
		if joinRequestMessage.GetJoinRequest().RejoinToken != nil {
			rejoinID := joinRequestMessage.GetJoinRequest().GetRejoinToken()
			if _, ok := rejoinIDs[rejoinID]; ok {
				successfulRejoin = true
			}
		}
		if successfulRejoin {
			rejoinID := joinRequestMessage.GetJoinRequest().GetRejoinToken()
			err = lsfn.SendSingleMessage(listener.conn, &lsfn.STSdown{
				JoinResponse: &lsfn.STSdown_JoinResponse{
					Type:        &rejoinAccept,
					RejoinToken: &rejoinID,
				},
			})
			if err != nil {
				return "", err
			}
			return joinRequestMessage.GetJoinRequest().GetRejoinToken(), nil
		} else {
			err = lsfn.SendSingleMessage(listener.conn, &lsfn.STSdown{
				JoinResponse: &lsfn.STSdown_JoinResponse{
					Type: &joinReject,
				},
			})
			if err != nil {
				return "", err
			}
			return "", nil
		}
	}
}
