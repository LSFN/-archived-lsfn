package lsfn

import (
	"net"

	"code.google.com/p/goprotobuf/proto"
)

func ReceiveSingleMessage(conn *net.TCPConn, message *proto.Message) error {
	// Read off the length of the message into a variant
	lengthVariant := new(Variant)
	singleByte := make([]byte, 1)
	for !lengthVariant.IsComplete() {
		bytes, err := conn.Read(singleByte)
		if err != nil {
			return err
		}
		if bytes == 1 {
			lengthVariant.ConnectByte(singleByte[0])
		}
	}

	// Receive a message of the stated length
	receiverSlice := make([]byte, lengthVariant.Uint64())
	var bytes int
	for bytes < len(receiverSlice) {
		x, err := conn.Read(receiverSlice[bytes:])
		if err != nil {
			return err
		}
		bytes += x
	}

	// Unmarshal the message into a protobuf structure
	err := proto.Unmarshal(receiverSlice, *message)
	if err != nil {
		message = nil
	}

	return err
}

func SendSingleMessage(conn *net.TCPConn, message *proto.Message) error {
	rawMessage, err := proto.Marshal(*message)
	if err != nil {
		return err
	}

	var bytes int
	variantLength := new(Variant)
	variantLength.FromUint64(uint64(len(rawMessage)))
	rawLength := variantLength.Bytes()
	for bytes < len(rawLength) {
		x, err := conn.Write(rawLength[bytes:])
		if err != nil {
			return err
		}
		bytes += x
	}

	bytes = 0
	for bytes < len(rawMessage) {
		x, err := conn.Write(rawMessage[bytes:])
		if err != nil {
			return err
		}
		bytes += x
	}

	return nil
}
