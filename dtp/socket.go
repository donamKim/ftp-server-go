/*
 * FTP Server Go
 *
 * Copyright (C) 2019 Donam Kim. All rights reserved.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along
 * with this program; if not, write to the Free Software Foundation, Inc.,
 * 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
 */

package dtp

import (
	"errors"
	"net"
	"os"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type Socket struct {
	Port int
	Conn *net.TCPConn

	mutex    sync.Mutex
	errAsync error
}

func NewActive(host string, port int) (*Socket, error) {
	raddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return nil, err
	}

	socket := new(Socket)
	socket.Port = port
	socket.Conn = conn

	return socket, nil
}

func NewPassive(ports []int) (*Socket, error) {
	socket := new(Socket)
	for _, v := range ports {
		if err := socket.listenAndServe(v); err != nil {
			if isEADDRINUSE(err) == true {
				continue
			}
			return nil, err
		}
		socket.Port = v

		return socket, nil
	}

	return nil, errors.New("not found available passive port")
}

func (r *Socket) listenAndServe(port int) error {
	laddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort("", strconv.Itoa(port)))
	if err != nil {
		return err
	}
	l, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return err
	}
	if err := l.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return err
	}

	r.mutex.Lock()
	go func() {
		defer r.mutex.Unlock()
		if r.Conn, r.errAsync = l.AcceptTCP(); r.errAsync != nil {
			return
		}
		r.errAsync = l.Close()
	}()

	return nil
}

func isEADDRINUSE(err error) bool {
	errOp, ok := err.(*net.OpError)
	if !ok {
		return false
	}
	errSyscall, ok := errOp.Err.(*os.SyscallError)
	if !ok {
		return false
	}
	no, ok := errSyscall.Err.(syscall.Errno)
	if !ok {
		return false
	}
	if no == syscall.EADDRINUSE {
		return true
	}
	if runtime.GOOS == "windows" && no == 10048 {
		return true
	}

	return false
}

func (r *Socket) Read(p []byte) (n int, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.errAsync != nil {
		return 0, r.errAsync
	}
	if r.Conn == nil {
		return 0, errors.New("nil conn")
	}

	return r.Conn.Read(p)
}

func (r *Socket) Write(p []byte) (n int, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.errAsync != nil {
		return 0, r.errAsync
	}
	if r.Conn == nil {
		return 0, errors.New("nil conn")
	}

	return r.Conn.Write(p)
}

func (r *Socket) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.Conn == nil {
		return errors.New("nil conn")
	}

	return r.Conn.Close()
}
