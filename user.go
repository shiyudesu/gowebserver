package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string //goroutine channel
	conn   net.Conn    //connection between user and client
	server *Server
}

func (u *User) sendMsg(msg string) {
	u.conn.Write([]byte(msg))
}

func (u *User) ListenChannel() {
	for {
		msg := <-u.C
		//u.conn.Write([]byte(msg + "\n"))
		u.sendMsg(msg + "\n")
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
	if msg == "who" {
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "online"
			u.sendMsg(onlineMsg)
		}
		u.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]

		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.sendMsg("This name has been used.\n")
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()
			u.Name = newName
			u.sendMsg("Name update:" + u.Name + "\n")
		}
	} else if len(msg) > 3 && msg[:3] == "to|" {
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			u.sendMsg("Wrong format, please use to|name|message\n")
			return
		}

		remoteUser, ok := u.server.OnlineMap[remoteName]
		if !ok {
			u.sendMsg("User Not Found.\n")
			return
		}

		content := strings.Split(msg, "|")[2]
		if content == "" {
			u.sendMsg("No Message.\n")
			return
		}
		remoteUser.sendMsg(u.Name + "said:" + content)
	} else {
		u.server.BroadCast(u, msg)
	}
}
