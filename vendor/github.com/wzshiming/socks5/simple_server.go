package socks5

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"
)

// SimpleServer is a simplified server, which can be configured as easily as client.
type SimpleServer struct {
	Server
	Listener net.Listener
	Network  string
	Address  string
	Username string
	Password string
}

// NewServer creates a new NewSimpleServer
func NewSimpleServer(addr string) (*SimpleServer, error) {
	s := &SimpleServer{}
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "socks5", "socks5h":
	default:
		return nil, fmt.Errorf("unsupported protocol '%s'", u.Scheme)
	}
	host := u.Host
	port := u.Port()
	if port == "" {
		port = "1080"
		hostname := u.Hostname()
		host = net.JoinHostPort(hostname, port)
	}
	if u.User != nil {
		s.Username = u.User.Username()
		s.Password, _ = u.User.Password()
		s.Authentication = UserAuth(s.Username, s.Password)
	}

	s.Address = host
	s.Network = "tcp"
	s.ListenBindReuseTimeout = time.Second / 2
	return s, nil
}

// Run the server
func (s *SimpleServer) Run(ctx context.Context) error {
	var listenConfig net.ListenConfig
	if s.Listener == nil {
		listener, err := listenConfig.Listen(ctx, s.Network, s.Address)
		if err != nil {
			return err
		}
		s.Listener = listener
	}
	s.Address = s.Listener.Addr().String()
	return s.Serve(s.Listener)
}

// Start the server
func (s *SimpleServer) Start(ctx context.Context) error {
	var listenConfig net.ListenConfig
	if s.Listener == nil {
		listener, err := listenConfig.Listen(ctx, s.Network, s.Address)
		if err != nil {
			return err
		}
		s.Listener = listener
	}
	s.Address = s.Listener.Addr().String()
	go s.Serve(s.Listener)
	return nil
}

// Close closes the listener
func (s *SimpleServer) Close() error {
	if s.Listener == nil {
		return nil
	}
	return s.Listener.Close()
}

// ProxyURL returns the URL of the proxy
func (s *SimpleServer) ProxyURL() string {
	u := url.URL{
		Scheme: "socks5",
		Host:   s.Address,
	}
	if s.Username != "" {
		u.User = url.UserPassword(s.Username, s.Password)
	}
	return u.String()
}
