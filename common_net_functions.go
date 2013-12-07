package lsfn

import (
	"io"
	"net"

	"code.google.com/p/goprotobuf/proto"
)

func ReceiveSingleMessage(conn net.Conn, message proto.Message) error {
	// Read off the length of the message into a variant
	lengthVariant := NewVariant()
	singleByte := make([]byte, 1)
	for !lengthVariant.IsComplete() {
		numBytes, err := conn.Read(singleByte)
		if err != nil {
			return err
		}
		if numBytes == 1 {
			lengthVariant.ConnectByte(singleByte[0])
		}
	}

	// Receive a message of the stated length
	receiverSlice := make([]byte, lengthVariant.Uint64())
	var numBytes int
	for numBytes < len(receiverSlice) {
		x, err := conn.Read(receiverSlice[numBytes:])
		if err != nil {
			return err
		}
		numBytes += x
	}

	// Unmarshal the message into a protobuf structure
	err := proto.Unmarshal(receiverSlice, message)
	if err != nil {
		message = nil
	}

	return err
}

func SendSingleMessage(conn net.Conn, message proto.Message) error {
	rawMessage, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	variantLength := NewVariant()
	variantLength.FromUint64(uint64(len(rawMessage)))
	rawLength := variantLength.Bytes()
	err = write(conn, rawLength)
	if err != nil {
		return err
	}

	err = write(conn, rawMessage)
	if err != nil {
		return err
	}

	return nil
}

// read is used to ensure that the given number of bytes
// are read if possible, even if multiple calls to Read
// are required.
func read(r io.Reader, i int) ([]byte, error) {
	out := make([]byte, i)
	in := out[:]
	for i > 0 {
		if n, err := r.Read(in); err != nil {
			return nil, err
		} else {
			in = in[n:]
			i -= n
		}
	}
	return out, nil
}

// write is used to ensure that the given data is written
// if possible, even if multiple calls to Write are
// required.
func write(w io.Writer, data []byte) error {
	i := len(data)
	for i > 0 {
		if n, err := w.Write(data); err != nil {
			return err
		} else {
			data = data[n:]
			i -= n
		}
	}
	return nil
}
