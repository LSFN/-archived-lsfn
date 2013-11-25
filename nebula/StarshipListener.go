package nebula

import (
	"net"

	"code.google.com/p/goprotobuf/proto"
	"code.google.com/p/uuid"

	"lsfn/common"
)

type StarshipListener struct {
	conn     *net.TCPConn
	Messages chan *common.STSup
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

func (listener *StarshipListener) receiveSingleMessage() (*common.STSup, error) {
	// Read off the length of the message into a variant
	lengthVariant := new(common.Variant)
	singleByte := make([]byte, 1)
	for !lengthVariant.isComplete() {
		bytes, err := listener.conn.Read(singleByte)
		if err != nil {
			return nil, err
		}
		if bytes == 1 {
			lengthVariant.ConnectByte(singleByte[0])
		}
	}

	// Receive a message of the stated length
	receiverSlice := make([]byte, lengthVariant.Uint64)
	var bytes int
	for bytes < len(receiverSlice) {
		x, err := listener.conn.Read(receiverSlice[bytes:])
		if err != nil {
			return nil, err
		}
		bytes += x
	}

	// Unmarshal the message into a protobuf structure
	result := new(common.STSup)
	err := proto.Unmarshal(receiverSlice, result)
	if err != nil {
		result = nil
	}

	return result, err
}

func (listener *StarshipListener) SendMessage(downMessage *common.STSdown) {
	err := listener.sendSingleMessage(downMessage)
	if err != nil {
		listener.Disconnect()
	}
}

func (listener *StarshipListener) sendSingleMessage(downMessage *common.STSdown) error {
	rawMessage, err := proto.Marshal(downMessage)
	if err != nil {
		return err
	}

	var bytes int
	variantLength := new(common.Variant)
	variantLength.FromUint64(len(rawMessage))
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
func (listener *StarshipListener) Handshake(gameID string, allowJoin bool, rejoinIDs map[string]int) string {
	// Send the JoinInfo message
	joinInfoMessage := &common.STSdown{
		JoinInfo: &common.STSdown_JoinInfo{
			AllowJoin:   &allowJoin,
			GameIDtoken: &gameID,
		},
	}
	err := listener.sendSingleMessage(joinInfoMessage)
	// if joins are not allowed then when allowJoin is false, handshakes will end here
	if err != nil {
		return ""
	}

	// Receive the JoinRequest message
	joinRequestMessage, err := listener.receiveSingleMessage()
	if err != nil {
		return ""
	}
	if joinRequestMessage.JoinRequest == nil {
		return ""
	}

	// Send the JoinResponse message
	joinType := joinRequestMessage.GetJoinRequest().GetType()
	var joinAccept = common.STSdown_JoinResponse_JOIN_ACCEPTED
	var joinReject = common.STSdown_JoinResponse_JOIN_REJECTED
	var rejoinAccept = common.STSdown_JoinResponse_REJOIN_ACCEPTED
	if joinType == common.STSup_JoinRequest_JOIN {
		if allowJoin {
			id := uuid.New()
			listener.sendSingleMessage(&common.STSdown{
				JoinResponse: &common.STSdown_JoinResponse{
					Type:        &joinAccept,
					RejoinToken: &id,
				},
			})
			return id
		} else {
			listener.sendSingleMessage(&common.STSdown{
				JoinResponse: &common.STSdown_JoinResponse{
					Type: &joinReject,
				},
			})
			return ""
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
			listener.sendSingleMessage(&common.STSdown{
				JoinResponse: &common.STSdown_JoinResponse{
					Type:        &rejoinAccept,
					RejoinToken: &rejoinID,
				},
			})
			return joinRequestMessage.GetJoinRequest().GetRejoinToken()
		} else {
			listener.sendSingleMessage(&common.STSdown{
				JoinResponse: &common.STSdown_JoinResponse{
					Type: &joinReject,
				},
			})
			return ""
		}
	}
}
