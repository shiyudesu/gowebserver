package main

import (
	"net"
)

type User struct {
	Name   string
	Addr   string
	C      chan string //goroutine channel
	conn   net.Conn    //connection between user and client
	server *Server
}

func (u *User) ListenChannel() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	go user.ListenChannel()

	return user
}

func (u *User) Online() {
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	u.server.BroadCast(u, "online")
}

func (u *User) Offline() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	u.server.BroadCast(u, "offline")
}

func (u *User) DoMessage(msg string) {
	u.server.BroadCast(u, msg)
}
