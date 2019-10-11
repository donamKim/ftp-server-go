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
	"fmt"
	"io"
	"log"
	"net"
	"path/filepath"
	"strings"

	"github.com/donamKim/ftp-server-go/dtp"
	"github.com/donamKim/ftp-server-go/file"
)

type conn struct {
	socket      *dtp.Socket
	reader      *bufio.Reader
	writer      *bufio.Writer
	manager     file.Manager
	addr        *net.TCPAddr
	user        User
	requester   string
	directory   string
	passivePort []int
	loggedIn    bool
	rnfr        string
}

func (r *conn) serve() {
	r.write(&reply{code: replyHello, message: "Service ready for new user."})

	for {
		cmd, err := r.read()
		if err != nil {
			if err != io.EOF {
				log.Printf("reader error: %v", err)
			}
			break
		}

		task := commands[cmd.fn]
		if task == nil {
			log.Printf("not found command: %v", cmd.fn)
			r.write(&reply{code: replyNotFoundCommand, message: "This command is not found."})
		} else if task.supported() == false {
			log.Printf("not supported command: %v", cmd.fn)
			r.write(&reply{code: replyNotSupportedParameter, message: "This command is not supported."})
		} else if task.requirePermission() == true && r.loggedIn == false {
			log.Printf("not permission to command: %v", cmd.fn)
			r.write(&reply{code: replyNotLoggedIn, message: "Not permission to this command."})
		} else if err := task.parse(cmd.param); err != nil {
			log.Printf("invalid parameter: fn=%v, err=%v", cmd.fn, err)
			r.write(&reply{code: replyInvalidParameter, message: fmt.Sprintf("Invalid parameter: %v", err)})
		} else {
			log.Printf("execute command: %v", cmd.fn)
			task.execute(r)
		}
	}
}

func (r *conn) write(reply *reply) {
	if _, err := r.writer.WriteString(reply.make()); err != nil {
		log.Printf("failed to write relpy: code=%v, message=%v, err=%v", reply.code, reply.message, err)
	} else {
		log.Printf("write reply: code=%v, message=%v", reply.code, reply.message)
	}
	if err := r.writer.Flush(); err != nil {
		log.Printf("failed to flush write buffer: %v", err)
	}
}

func (r *conn) read() (*cmd, error) {
	cmd, err := r.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	return newCMD(cmd), nil
}

func (r *conn) writeSocket(f io.Reader) {
	if r.socket == nil {
		return
	}

	if _, err := io.Copy(r.socket, f); err != nil {
		log.Printf("failed to write socket: %v", err)
	}
	if err := r.socket.Close(); err != nil {
		log.Printf("failed to close socket: %v", err)
	}
	r.socket = nil
}

func (r *conn) buildPath(path string) string {
	if len(path) == 0 {
		return r.directory
	}
	if path[0:1] == string(filepath.Separator) {
		return path
	}
	if path == ".." {
		s := strings.Split(r.directory, string(filepath.Separator))
		return string(filepath.Separator) + filepath.Join(s[:len(s)-1]...)
	}

	return r.directory + string(filepath.Separator) + path
}
