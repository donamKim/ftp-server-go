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

import "strings"

type cmd struct {
	fn    string
	param string
}

func newCMD(s string) *cmd {
	v := strings.Split(strings.TrimRight(s, "\r\n"), " ")
	cmd := new(cmd)
	cmd.fn = strings.ToUpper(v[0])
	if len(v) == 1 {
		cmd.param = ""
	} else {
		cmd.param = strings.Join(v[1:], " ")
	}

	return cmd
}

var commands = map[string]task{
	"AUTH": new(taskAUTH),
	"USER": new(taskUSER),
	"PASS": new(taskPASS),
	"FEAT": new(taskFEAT),
	"PWD":  new(taskPWD),
	"TYPE": new(taskTYPE),
	"PASV": new(taskPASV),
	"PORT": new(taskPORT),
	"EPSV": new(taskEPSV),
	"EPRT": new(taskEPRT),
	"LIST": new(taskLIST),
	"CWD":  new(taskCWD),
	"RETR": new(taskRETR),
	"STOR": new(taskSTOR),
	"DELE": new(taskDELE),
	"RMD":  new(taskRMD),
	"RNFR": new(taskRNFR),
	"RNTO": new(taskRNTO),
	"SIZE": new(taskSIZE),
}
