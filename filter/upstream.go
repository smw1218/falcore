package filter

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

type Upstream struct {
	// The upstream host to connect to
	Host string
	// The port on the upstream host
	Port int
	// Default 60 seconds
	Timeout time.Duration
	// Will ignore https on the incoming request and always upstream http
	ForceHttp bool
	// Ping URL Path-only for checking upness
	PingPath string

	transport *http.Transport
	host      string
	tcpaddr   *net.TCPAddr
	tcpconn   *net.TCPConn
}

func NewUpstream(host string, port int, forceHttp bool) *Upstream {
	u := new(Upstream)
	u.Host = host
	u.Port = port
	u.ForceHttp = forceHttp
	ips, err := net.LookupIP(host)
	var ip net.IP = nil
	for i := range ips {
		ip = ips[i].To4()
		if ip != nil {
			break
		}
	}
	if err == nil && ip != nil {
		u.tcpaddr = new(net.TCPAddr)
		u.tcpaddr.Port = port
		u.tcpaddr.IP = ip
	} else {
		Warn("Can't get IP addr for %v: %v", host, err)
	}
	u.Timeout = 60e9
	u.host = fmt.Sprintf("%v:%v", u.Host, u.Port)

	u.transport = new(http.Transport)

	u.transport.Dial = func(n, addr string) (c net.Conn, err error) {
		Fine("Dialing connection to %v", u.tcpaddr)
		var ctcp *net.TCPConn
		ctcp, err = net.DialTCP("tcp4", nil, u.tcpaddr)
		if ctcp != nil {
			u.tcpconn = ctcp
			u.tcpconn.SetDeadline(time.Now().Add(u.Timeout))
		}
		if err != nil {
			Error("Dial Failed: %v", err)
		}
		return ctcp, err
	}
	u.transport.MaxIdleConnsPerHost = 15
	return u
}

// Alter the number of connections to multiplex with
func (u *Upstream) SetPoolSize(size int) {
	u.transport.MaxIdleConnsPerHost = size
}

func (u *Upstream) FilterRequest(request *Request) (res *http.Response) {
	var err error
	req := request.HttpRequest

	// Force the upstream to use http 
	if u.ForceHttp || req.URL.Scheme == "" {
		req.URL.Scheme = "http"
		req.URL.Host = req.Host
	}
	before := time.Now()
	req.Header.Set("Connection", "Keep-Alive")
	if u.tcpconn != nil {
		u.tcpconn.SetDeadline(time.Now().Add(u.Timeout))
	}
	res, err = u.transport.RoundTrip(req)
	diff := TimeDiff(before, time.Now())
	if err != nil {
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			Error("%s Upstream Timeout error: %v", request.ID, err)
			res = SimpleResponse(req, 504, nil, "Gateway Timeout\n")
			request.CurrentStage.Status = 2 // Fail
		} else {
			Error("%s Upstream error: %v", request.ID, err)
			res = SimpleResponse(req, 502, nil, "Bad Gateway\n")
			request.CurrentStage.Status = 2 // Fail
		}
	}
	Debug("%s [%s] [%s] %s s=%d Time=%.4f", request.ID, req.Method, u.host, req.URL, res.StatusCode, diff)
	return
}

func (u *Upstream) ping() (up bool, ok bool) {
	if u.PingPath != "" {
		// the url must be syntactically valid for this to work but the host will be ignored because we
		// are overriding the connection always
		request, err := http.NewRequest("GET", "http://localhost"+u.PingPath, nil)
		request.Header.Set("Connection", "Keep-Alive") // not sure if this should be here for a ping
		if err != nil {
			Error("Bad Ping request: %v", err)
			return false, true
		}
		if u.tcpconn != nil {
			u.tcpconn.SetDeadline(time.Now().Add(u.Timeout))
		}
		res, err := u.transport.RoundTrip(request)

		if err != nil {
			Error("Failed Ping to %v:%v: %v", u.Host, u.Port, err)
			return false, true
		} else {
			res.Body.Close()
		}
		if res.StatusCode == 200 {
			return true, true
		}
		Error("Failed Ping to %v:%v: %v", u.Host, u.Port, res.Status)
		// bad status
		return false, true
	}
	return false, false
}
