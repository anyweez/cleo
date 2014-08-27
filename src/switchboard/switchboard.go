package switchboard

import yamux "github.com/hashicorp/yamux"
import "net"

type SwitchboardClient struct {
	Connection *net.TCPConn
	Session    *yamux.Session
}

type SwitchboardServer struct {
	Connection *net.TCPConn
	Session    *yamux.Session
	Listener   *net.TCPListener
}

/**
 * Create a new Switchboard client. The client should establish a connection
 * to the desired network endpoint so that creating channels on the connection
 * is fast. The client won't exist if the connection can't be established.
 */
func NewClient(network string, address *net.TCPAddr) (SwitchboardClient, error) {
	sbc := SwitchboardClient{}
	// TODO: there must be a better way to take care of the net.Dial call...
	cerr, serr := error(nil), error(nil)
	sbc.Connection, cerr = net.DialTCP(network, nil, address)

	if cerr != nil {
		return sbc, cerr
	}

	sbc.Session, serr = yamux.Client(sbc.Connection, nil)

	if serr != nil {
		return sbc, serr
	}

	return sbc, nil
}

func NewServer(network string, address *net.TCPAddr) (SwitchboardServer, error) {
	sbs := SwitchboardServer{}
	cerr, aerr, serr := error(nil), error(nil), error(nil)
	sbs.Listener, cerr = net.ListenTCP(network, address)

	if cerr != nil {
		return sbs, cerr
	}

	sbs.Connection, aerr = sbs.Listener.AcceptTCP()

	if aerr != nil {
		return sbs, aerr
	}

	sbs.Session, serr = yamux.Server(sbs.Connection, nil)

	if serr != nil {
		return sbs, serr
	}

	return sbs, nil
}

/**
 * This method creates a new stream on the Client's connection.
 */
func (c *SwitchboardClient) GetStream() (*net.Conn, error) {
	stream, err := c.Session.Open()

	return &stream, err
}

/**
 * This method listens on the Server's port until an inbound connection
 * arrives. When that happens, return a new stream.
 */
func (s *SwitchboardServer) GetStream() (*net.Conn, error) {
	stream, err := s.Session.Accept()

	return &stream, err
}
