package pearl

import (
	"fmt"
	"io"
	"net"

	"github.com/mmcloughlin/openssl"
	"github.com/mmcloughlin/pearl/log"
)

// Connection encapsulates a router connection.
type Connection struct {
	router  *Router
	conn    net.Conn
	tlsCtx  *TLSContext
	tlsConn *openssl.Conn

	logger log.Logger
}

// NewConnection constructs a connection
func NewConnection(r *Router, conn net.Conn, logger log.Logger) (*Connection, error) {
	tlsCtx, err := NewTLSContext(r.IdentityKey())
	if err != nil {
		return nil, err
	}

	tlsConn, err := tlsCtx.ServerConn(conn)
	if err != nil {
		return nil, err
	}

	return &Connection{
		router:  r,
		conn:    conn,
		tlsCtx:  tlsCtx,
		tlsConn: tlsConn,

		logger: logger.With("raddr", conn.RemoteAddr()),
	}, nil
}

// Handle handles the full lifecycle of the connection.
func (c *Connection) Handle() error {
	c.logger.Info("handle")

	// XXX read from conn
	buf := make([]byte, 5)
	_, err := io.ReadFull(c.tlsConn, buf)
	if err != nil {
		return err
	}

	fmt.Println(buf)

	return nil
}
