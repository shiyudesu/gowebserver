package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	IP   string
	Port int

	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	Message chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- sendMsg
}

func (s *Server) Handler(conn net.Conn) {
	//fmt.Println("connection successful")

	user := NewUser(conn)

	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()

	s.BroadCast(user, "online")

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				s.BroadCast(user, "offline")
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			msg := string(buf[:n-1])

			s.BroadCast(user, msg)
		}
	}()

	select {}
}

func (s *Server) ListenMessage() {
	for {
		msg := <-s.Message

		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		s.mapLock.Unlock()
	}
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		fmt.Println("listener err:", err)
		return
	}

	defer listener.Close()

	go s.ListenMessage()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("connect err", err)
			continue
		}

		go s.Handler(conn)
	}
}
