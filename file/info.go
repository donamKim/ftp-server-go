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

package file

import (
	"bytes"
	"fmt"
	"os"
	"syscall"
)

type Info struct {
	os.FileInfo
	Uid uint32
	Gid uint32
}

func NewInfo(info os.FileInfo) *Info {
	v := new(Info)
	v.FileInfo = info
	if stat := v.Sys().(*syscall.Stat_t); stat != nil {
		v.Uid = stat.Uid
		v.Gid = stat.Gid
	}

	return v
}

func (r *Info) Encode() []byte {
	var buf bytes.Buffer
	if r.IsDir() == true {
		buf.WriteString(fmt.Sprintf("Type=%v;", "dir"))
	} else {
		buf.WriteString(fmt.Sprintf("Type=%v;", "file"))
		buf.WriteString(fmt.Sprintf("Size=%v;", r.Size()))
	}
	buf.WriteString(fmt.Sprintf("UNIX.owner=%v;", r.Uid))
	buf.WriteString(fmt.Sprintf("UNIX.group=%v;", r.Gid))
	buf.WriteString(fmt.Sprintf("Modify=%v;", r.ModTime().Format("20060102150405")))
	buf.WriteString(fmt.Sprintf("Perm=%v;", r.Mode().String()))
	buf.WriteString(fmt.Sprintf(" %v\r\n", r.Name()))

	return buf.Bytes()
}
