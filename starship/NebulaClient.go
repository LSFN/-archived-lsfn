package starship

import (
	"net"

	"code.google.com/p/goprotobuf/proto"

	"lsfn/common"
)

type NebulaClient struct {
	starshipID string
	conn       *TCPConn
}

func (client *NebulaClient) Join(host string, port int) {

}
