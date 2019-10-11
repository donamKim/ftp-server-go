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
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/donamKim/ftp-server-go/dtp"
)

type task interface {
	supported() bool
	requirePermission() bool
	parse(param string) error
	execute(conn *conn)
}

type taskAUTH struct{}

func (r *taskAUTH) supported() bool {
	return false
}

func (r *taskAUTH) requirePermission() bool {
	return false
}

func (r *taskAUTH) parse(param string) error {
	return nil
}

func (r *taskAUTH) execute(conn *conn) {}

type taskUSER struct {
	name string
}

func (r *taskUSER) supported() bool {
	return true
}

func (r *taskUSER) requirePermission() bool {
	return false
}

func (r *taskUSER) parse(param string) error {
	if len(param) == 0 {
		return errors.New("empty param")
	}
	r.name = param
	return nil
}

func (r *taskUSER) execute(conn *conn) {
	conn.requester = r.name
	conn.write(&reply{code: replyUserNameOkay, message: "User name okay, need password."})
}

type taskPASS struct {
	password string
}

func (r *taskPASS) supported() bool {
	return true
}

func (r *taskPASS) requirePermission() bool {
	return false
}

func (r *taskPASS) parse(param string) error {
	if len(param) == 0 {
		return errors.New("empty param")
	}
	r.password = param
	return nil
}

func (r *taskPASS) execute(conn *conn) {
	if conn.requester == conn.user.Name && r.password == conn.user.Password {
		conn.loggedIn = true
		conn.write(&reply{code: replyLoggedIn, message: "User logged in, proceed."})
	} else {
		conn.write(&reply{code: replyNotLoggedIn, message: "Not logged in."})
	}
}

type taskFEAT struct{}

func (r *taskFEAT) supported() bool {
	return true
}

func (r *taskFEAT) requirePermission() bool {
	return false
}

func (r *taskFEAT) parse(param string) error {
	return nil
}

func (r *taskFEAT) execute(conn *conn) {
	conn.write(&reply{code: replySystemStatus, message: "Extensions supported:\n UTF8\n", multiline: true})
}

type taskPWD struct{}

func (r *taskPWD) supported() bool {
	return true
}

func (r *taskPWD) requirePermission() bool {
	return false
}

func (r *taskPWD) parse(param string) error {
	return nil
}

func (r *taskPWD) execute(conn *conn) {
	conn.write(&reply{code: replyPathnameOkay, message: fmt.Sprintf("\"%v\" is the current directory", conn.directory)})
}

type taskTYPE struct {
	code string
}

func (r *taskTYPE) supported() bool {
	return true
}

func (r *taskTYPE) requirePermission() bool {
	return true
}

func (r *taskTYPE) parse(param string) error {
	if len(param) == 0 {
		return errors.New("empty param")
	}
	param = strings.ToUpper(param)
	s := strings.Split(param, " ")
	if s[0] != "A" && s[0] != "E" && s[0] != "I" && s[0] != "L" {
		return fmt.Errorf("invalid type code: %v", s[0])
	}
	r.code = s[0]

	return nil
}

func (r *taskTYPE) execute(conn *conn) {
	switch r.code {
	case "A":
		conn.write(&reply{code: replyOkay, message: "Representation type set to ASCII."})
	case "I":
		conn.write(&reply{code: replyOkay, message: "Representation type set to Image."})
	default:
		conn.write(&reply{code: replyNotSupportedCommand, message: "Command not implemented for that parameter."})
	}
}

type taskPASV struct{}

func (r *taskPASV) supported() bool {
	return true
}

func (r *taskPASV) requirePermission() bool {
	return true
}

func (r *taskPASV) parse(param string) error {
	return nil
}

func (r *taskPASV) execute(conn *conn) {
	socket, err := dtp.NewPassive(conn.passivePort)
	if err != nil {
		conn.write(&reply{code: replyFailedOpenDTP, message: fmt.Sprintf("Can't open data connection: %v", err)})
		return
	}
	conn.socket = socket

	p1 := socket.Port / 256
	p2 := socket.Port - (p1 * 256)
	h := strings.Split(conn.addr.IP.String(), ".")
	conn.write(&reply{code: replyPASVOkay, message: fmt.Sprintf("Entering Passive Mode (%v,%v,%v,%v,%v,%v).", h[0], h[1], h[2], h[3], p1, p2)})
}

type taskPORT struct {
	host string
	port int
}

func (r *taskPORT) supported() bool {
	return true
}

func (r *taskPORT) requirePermission() bool {
	return true
}

func (r *taskPORT) parse(param string) (err error) {
	s := strings.Split(param, ",")
	if len(s) != 6 {
		return errors.New("invalid parameter format")
	}
	p1, err := strconv.Atoi(s[4])
	if err != nil {
		return err
	}
	p2, err := strconv.Atoi(s[5])
	if err != nil {
		return err
	}

	r.port = (p1 * 256) + p2
	r.host = s[0] + "." + s[1] + "." + s[2] + "." + s[3]

	return nil
}

func (r *taskPORT) execute(conn *conn) {
	socket, err := dtp.NewActive(r.host, r.port)
	if err != nil {
		conn.write(&reply{code: replyFailedOpenDTP, message: fmt.Sprintf("Can't open data connection: %v", err)})
		return
	}
	conn.socket = socket

	conn.write(&reply{code: replyOkay, message: "Connection established on active mode."})
}

type taskEPSV struct{}

func (r *taskEPSV) supported() bool {
	return true
}

func (r *taskEPSV) requirePermission() bool {
	return true
}

func (r *taskEPSV) parse(param string) error {
	return nil
}

func (r *taskEPSV) execute(conn *conn) {
	socket, err := dtp.NewPassive(conn.passivePort)
	if err != nil {
		conn.write(&reply{code: replyFailedOpenDTP, message: fmt.Sprintf("Can't open data connection: %v", err)})
		return
	}
	conn.socket = socket

	conn.write(&reply{code: replyEPSVOkay, message: fmt.Sprintf("Entering Extended Passive Mode (|||%d|)", socket.Port)})
}

type taskEPRT struct {
	family int
	host   string
	port   int
}

func (r *taskEPRT) supported() bool {
	return true
}

func (r *taskEPRT) requirePermission() bool {
	return true
}

func (r *taskEPRT) parse(param string) (err error) {
	s := strings.Split(param, "|")
	if len(s) != 5 {
		return errors.New("invalid parameter format")
	}
	if r.family, err = strconv.Atoi(s[1]); err != nil {
		return err
	}
	r.host = s[2]
	if r.port, err = strconv.Atoi(s[3]); err != nil {
		return err
	}

	return nil
}

func (r *taskEPRT) execute(conn *conn) {
	if r.family != 1 && r.family != 2 {
		conn.write(&reply{code: replyNotSupportedNetwork, message: "Network protocol not supported, use (1,2)"})
		return
	}
	socket, err := dtp.NewActive(r.host, r.port)
	if err != nil {
		conn.write(&reply{code: replyFailedOpenDTP, message: fmt.Sprintf("Can't open data connection: %v", err)})
		return
	}
	conn.socket = socket

	conn.write(&reply{code: replyOkay, message: "Connection established on active mode."})
}

type taskLIST struct {
	path string
}

func (r *taskLIST) supported() bool {
	return true
}

func (r *taskLIST) requirePermission() bool {
	return true
}

func (r *taskLIST) parse(param string) error {
	r.path = param
	return nil
}

func (r *taskLIST) execute(conn *conn) {
	info, err := conn.manager.Stat(conn.buildPath(r.path))
	if err != nil {
		conn.write(&reply{code: replyUnavailableFile, message: fmt.Sprintf("File unavailable: %v", err)})
		return
	}

	var buf bytes.Buffer
	if info.IsDir() == true {
		list, err := conn.manager.List(conn.buildPath(r.path))
		if err != nil {
			conn.write(&reply{code: replyUnavailableFile, message: fmt.Sprintf("File unavailable: %v", err)})
			return
		}
		for _, v := range list {
			buf.Write(v.Encode())
		}
	} else {
		buf.Write(info.Encode())
	}

	conn.write(&reply{code: replyFileStatusOkay, message: "File status okay; about to open data connection."})
	conn.writeSocket(bytes.NewReader(buf.Bytes()))
	conn.write(&reply{code: replyCloseDTP, message: "Closing data connection."})
}

type taskCWD struct {
	path string
}

func (r *taskCWD) supported() bool {
	return true
}

func (r *taskCWD) requirePermission() bool {
	return true
}

func (r *taskCWD) parse(param string) error {
	r.path = param
	return nil
}

func (r *taskCWD) execute(conn *conn) {
	conn.directory = conn.buildPath(r.path)
	conn.write(&reply{code: replyFileActionOkay, message: "Requested file action okay, completed."})
}

type taskRETR struct {
	path string
}

func (r *taskRETR) supported() bool {
	return true
}

func (r *taskRETR) requirePermission() bool {
	return true
}

func (r *taskRETR) parse(param string) error {
	r.path = param
	return nil
}

func (r *taskRETR) execute(conn *conn) {
	f, err := conn.manager.Get(conn.buildPath(r.path))
	if err != nil {
		conn.write(&reply{code: replyUnavailableFile, message: fmt.Sprintf("File unavailable: %v", err)})
		return
	}

	conn.write(&reply{code: replyFileStatusOkay, message: "File status okay; about to open data connection."})
	conn.writeSocket(f)
	conn.write(&reply{code: replyCloseDTP, message: "Closing data connection."})
}

type taskSTOR struct {
	path string
}

func (r *taskSTOR) supported() bool {
	return true
}

func (r *taskSTOR) requirePermission() bool {
	return true
}

func (r *taskSTOR) parse(param string) error {
	r.path = param
	return nil
}

func (r *taskSTOR) execute(conn *conn) {
	conn.write(&reply{code: replyFileStatusOkay, message: "File status okay; about to open data connection."})
	if err := conn.manager.Put(conn.buildPath(r.path), conn.socket); err != nil {
		conn.write(&reply{code: replyUnavailableFile, message: fmt.Sprintf("File unavailable: %v", err)})
		return
	}
	conn.write(&reply{code: replyCloseDTP, message: "Closing data connection."})
}

type taskDELE struct {
	path string
}

func (r *taskDELE) supported() bool {
	return true
}

func (r *taskDELE) requirePermission() bool {
	return true
}

func (r *taskDELE) parse(param string) error {
	r.path = param
	return nil
}

func (r *taskDELE) execute(conn *conn) {
	info, err := conn.manager.Stat(conn.buildPath(r.path))
	if err != nil {
		conn.write(&reply{code: replyUnavailableFile, message: fmt.Sprintf("File unavailable: %v", err)})
		return
	}
	if info.IsDir() == true {
		conn.write(&reply{code: replyUnavailableFile, message: "File unavailable: is directory"})
		return
	}

	if err := conn.manager.Remove(conn.buildPath(r.path)); err != nil {
		conn.write(&reply{code: replyUnavailableFile, message: fmt.Sprintf("File unavailable: %v", err)})
		return
	}
	conn.write(&reply{code: replyFileActionOkay, message: "Requested file action okay, completed."})
}

type taskRMD struct {
	path string
}

func (r *taskRMD) supported() bool {
	return true
}

func (r *taskRMD) requirePermission() bool {
	return true
}

func (r *taskRMD) parse(param string) error {
	r.path = param
	return nil
}

func (r *taskRMD) execute(conn *conn) {
	info, err := conn.manager.Stat(conn.buildPath(r.path))
	if err != nil {
		conn.write(&reply{code: replyUnavailableFile, message: fmt.Sprintf("File unavailable: %v", err)})
		return
	}
	if info.IsDir() == false {
		conn.write(&reply{code: replyUnavailableFile, message: "File unavailable: is not directory"})
		return
	}

	if err := conn.manager.Remove(conn.buildPath(r.path)); err != nil {
		conn.write(&reply{code: replyUnavailableFile, message: fmt.Sprintf("File unavailable: %v", err)})
		return
	}
	conn.write(&reply{code: replyFileActionOkay, message: "Requested file action okay, completed."})
}

type taskRNFR struct {
	path string
}

func (r *taskRNFR) supported() bool {
	return true
}

func (r *taskRNFR) requirePermission() bool {
	return true
}

func (r *taskRNFR) parse(param string) error {
	r.path = param
	return nil
}

func (r *taskRNFR) execute(conn *conn) {
	conn.rnfr = conn.buildPath(r.path)
	conn.write(&reply{code: replyFileActionPending, message: "Requested file action pending further information."})
}

type taskRNTO struct {
	path string
}

func (r *taskRNTO) supported() bool {
	return true
}

func (r *taskRNTO) requirePermission() bool {
	return true
}

func (r *taskRNTO) parse(param string) error {
	r.path = param
	return nil
}

func (r *taskRNTO) execute(conn *conn) {
	if err := conn.manager.Rename(conn.rnfr, conn.buildPath(r.path)); err != nil {
		conn.write(&reply{code: replyUnavailableFile, message: fmt.Sprintf("File unavailable: %v", err)})
		return
	}
	conn.rnfr = ""
	conn.write(&reply{code: replyFileActionOkay, message: "Requested file action okay, completed."})
}

type taskSIZE struct {
	path string
}

func (r *taskSIZE) supported() bool {
	return true
}

func (r *taskSIZE) requirePermission() bool {
	return true
}

func (r *taskSIZE) parse(param string) error {
	r.path = param
	return nil
}

func (r *taskSIZE) execute(conn *conn) {
	info, err := conn.manager.Stat(conn.buildPath(r.path))
	if err != nil {
		conn.write(&reply{code: replyUnavailableFile, message: fmt.Sprintf("File unavailable: %v", err)})
		return
	}
	conn.write(&reply{code: replyFileStatus, message: strconv.Itoa(int(info.Size()))})
}
