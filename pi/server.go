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

package pi

import (
	"bufio"
	"errors"
	"net"
	"strconv"

	"github.com/donamKim/ftp-server-go/file/driver"
)

var ErrServerClosed = errors.New("ftp: Server closed")

type Server struct {
	User        User
	Root        string
	PIPort      int
	PassivePort []int
}

type User struct {
	Name     string
	Password string
}

func (r *Server) ListenAndServe() error {
	l, err := r.listen()
	if err != nil {
		return err
	}
	return r.serve(l)
}

func (r *Server) listen() (*net.TCPListener, error) {
	laddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort("", strconv.Itoa(r.PIPort)))
	if err != nil {
		return nil, err
	}
	return net.ListenTCP("tcp", laddr)
}

func (r *Server) serve(l *net.TCPListener) error {
	for {
		v, err := l.AcceptTCP()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				continue
			}
			return err
		}
		c := r.newConn(v)
		go c.serve()
	}
}

func (r *Server) newConn(c *net.TCPConn) *conn {
	return &conn{
		manager:     &driver.Driver{},
		reader:      bufio.NewReader(c),
		writer:      bufio.NewWriter(c),
		addr:        c.LocalAddr().(*net.TCPAddr),
		user:        r.User,
		directory:   r.Root,
		passivePort: r.PassivePort,
	}
}
