package documents

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// Scanner talks to a ClamAV daemon (clamd) over TCP using the INSTREAM
// protocol. It has no third-party dependencies. When no address is configured
// scanning is disabled and everything is reported clean (dev default).
type Scanner struct {
	addr string
}

func NewScanner(addr string) *Scanner {
	return &Scanner{addr: addr}
}

// Enabled reports whether a clamd address is configured.
func (s *Scanner) Enabled() bool { return s.addr != "" }

// Scan streams r to clamd and returns "clean", "infected", or "error".
func (s *Scanner) Scan(ctx context.Context, r io.Reader) (string, error) {
	if s.addr == "" {
		return "clean", nil // scanning disabled
	}

	d := net.Dialer{Timeout: 5 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", s.addr)
	if err != nil {
		return "error", fmt.Errorf("dial clamd: %w", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(60 * time.Second))

	if _, err := conn.Write([]byte("zINSTREAM\x00")); err != nil {
		return "error", err
	}

	buf := make([]byte, 32*1024)
	for {
		n, rerr := r.Read(buf)
		if n > 0 {
			var sz [4]byte
			binary.BigEndian.PutUint32(sz[:], uint32(n))
			if _, err := conn.Write(sz[:]); err != nil {
				return "error", err
			}
			if _, err := conn.Write(buf[:n]); err != nil {
				return "error", err
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return "error", rerr
		}
	}
	// A zero-length chunk terminates the stream.
	if _, err := conn.Write([]byte{0, 0, 0, 0}); err != nil {
		return "error", err
	}

	resp, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil && err != io.EOF {
		return "error", err
	}
	resp = strings.TrimSpace(resp)
	switch {
	case strings.Contains(resp, "FOUND"):
		return "infected", nil
	case strings.HasSuffix(resp, "OK"):
		return "clean", nil
	default:
		return "error", fmt.Errorf("clamav response: %q", resp)
	}
}
