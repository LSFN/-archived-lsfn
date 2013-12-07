package nebula

import (
	"errors"
	"fmt"
	"net"

	"code.google.com/p/goprotobuf/proto"
	"code.google.com/p/uuid"

	"github.com/LSFN/lsfn"
)

type StarshipListener struct {
	conn     *net.TCPConn
	Messages chan *lsfn.STSup
}

func NewStarshipListener(conn *net.TCPConn) *StarshipListener {
	listener := new(StarshipListener)
	listener.conn = conn
	listener.Messages = make(chan *lsfn.STSup)
	return listener
}

func (listener *StarshipListener) Listen() {
	for {
		message, err := listener.receiveSingleMessage()
		if err != nil {
			listener.conn.Close()
			close(listener.Messages)
			break
		}
		listener.Messages <- message
	}
}

func (listener *StarshipListener) receiveSingleMessage() (*lsfn.STSup, error) {
	// Read off the length of the message into a variant
	lengthVariant := lsfn.NewVariant()
	singleByte := make([]byte, 1)
	fmt.Println(1)
	for !lengthVariant.IsComplete() {
		bytes, err := listener.conn.Read(singleByte)
		fmt.Println(bytes, singleByte)
		if err != nil {
			return nil, err
		}
		if bytes == 1 {
			lengthVariant.ConnectByte(singleByte[0])
		}
	}
	fmt.Println(2)

	// Receive a message of the stated length
	receiverSlice := make([]byte, lengthVariant.Uint64())
	var bytes int
	for bytes < len(receiverSlice) {
		x, err := listener.conn.Read(receiverSlice[bytes:])
		if err != nil {
			return nil, err
		}
		bytes += x
	}
	fmt.Println(3)

	// Unmarshal the message into a protobuf structure
	result := new(lsfn.STSup)
	err := proto.Unmarshal(receiverSlice, result)
	if err != nil {
		result = nil
	}
	fmt.Println(4)

	return result, err
}

func (listener *StarshipListener) SendMessage(downMessage *lsfn.STSdown) {
	err := listener.sendSingleMessage(downMessage)
	if err != nil {
		listener.Disconnect()
	}
}

func (listener *StarshipListener) sendSingleMessage(downMessage *lsfn.STSdown) error {
	rawMessage, err := proto.Marshal(downMessage)
	if err != nil {
		return err
	}

	var bytes int
	variantLength := lsfn.NewVariant()
	variantLength.FromUint64(uint64(len(rawMessage)))
	rawLength := variantLength.Bytes()
	for bytes < len(rawLength) {
		x, err := listener.conn.Write(rawLength[bytes:])
		if err != nil {
			return err
		}
		bytes += x
	}

	bytes = 0
	for bytes < len(rawMessage) {
		x, err := listener.conn.Write(rawMessage[bytes:])
		if err != nil {
			return err
		}
		bytes += x
	}

	return nil
}

func (listener *StarshipListener) Disconnect() {
	listener.conn.Close()
	close(listener.Messages)
}

// TODO possibly disconnect
func (listener *StarshipListener) Handshake(gameID string, allowJoin bool, rejoinIDs map[string]int) (string, error) {
	// Send the JoinInfo message
	joinInfoMessage := &lsfn.STSdown{
		JoinInfo: &lsfn.STSdown_JoinInfo{
			AllowJoin:   &allowJoin,
			GameIDtoken: &gameID,
		},
	}
	err := listener.sendSingleMessage(joinInfoMessage)
	fmt.Println("Sent join info")
	// if joins are not allowed then when allowJoin is false, handshakes will end here
	if err != nil {
		return "", err
	}

	// Receive the JoinRequest message
	joinRequestMessage, err := listener.receiveSingleMessage()
	if err != nil {
		return "", err
	}
	if joinRequestMessage.JoinRequest == nil {
		return "", errors.New("Received message does not have JoinRequest field set")
	}
	fmt.Println("Received join request")

	// Send the JoinResponse message
	joinType := joinRequestMessage.GetJoinRequest().GetType()
	var joinAccept = lsfn.STSdown_JoinResponse_JOIN_ACCEPTED
	var joinReject = lsfn.STSdown_JoinResponse_JOIN_REJECTED
	var rejoinAccept = lsfn.STSdown_JoinResponse_REJOIN_ACCEPTED
	if joinType == lsfn.STSup_JoinRequest_JOIN {
		fmt.Println("Client is trying to join")
		if allowJoin {
			fmt.Println("Client is being allowed to join")
			id := uuid.New()
			listener.sendSingleMessage(&lsfn.STSdown{
				JoinResponse: &lsfn.STSdown_JoinResponse{
					Type:        &joinAccept,
					RejoinToken: &id,
				},
			})
			fmt.Println("Sent join response")
			return id, nil
		} else {
			fmt.Println("Client in not allowed to join")
			listener.sendSingleMessage(&lsfn.STSdown{
				JoinResponse: &lsfn.STSdown_JoinResponse{
					Type: &joinReject,
				},
			})
			fmt.Println("Sent join response")
			return "", nil
		}
	} else {
		fmt.Println("Client is trying to rejoin")
		var successfulRejoin bool = false
		if joinRequestMessage.GetJoinRequest().RejoinToken != nil {
			rejoinID := joinRequestMessage.GetJoinRequest().GetRejoinToken()
			if _, ok := rejoinIDs[rejoinID]; ok {
				successfulRejoin = true
			}
		}
		if successfulRejoin {
			fmt.Println("Client rejoined successfully")
			rejoinID := joinRequestMessage.GetJoinRequest().GetRejoinToken()
			listener.sendSingleMessage(&lsfn.STSdown{
				JoinResponse: &lsfn.STSdown_JoinResponse{
					Type:        &rejoinAccept,
					RejoinToken: &rejoinID,
				},
			})
			fmt.Println("Sent join response")
			return joinRequestMessage.GetJoinRequest().GetRejoinToken(), nil
		} else {
			fmt.Println("Client failed to rejoin")
			listener.sendSingleMessage(&lsfn.STSdown{
				JoinResponse: &lsfn.STSdown_JoinResponse{
					Type: &joinReject,
				},
			})
			fmt.Println("Sent join response")
			return "", nil
		}
	}
}
