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

package driver

import (
	"io"
	"os"
	"path/filepath"

	"github.com/donamKim/ftp-server-go/file"
)

type Driver struct{}

func (r *Driver) Stat(path string) (*file.Info, error) {
	f, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	return file.NewInfo(f), nil
}

func (r *Driver) List(path string) ([]*file.Info, error) {
	list := make([]*file.Info, 0)
	err := filepath.Walk(path, func(p string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if p == path {
			return nil
		}

		info := file.NewInfo(f)
		list = append(list, info)
		if info.IsDir() == true {
			return filepath.SkipDir
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *Driver) Get(path string) (io.Reader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	return f, nil
}

func (r *Driver) Put(path string, reader io.Reader) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, reader); err != nil {
		return err
	}

	return nil
}

func (r *Driver) Remove(path string) error {
	return os.Remove(path)
}

func (r *Driver) Rename(old string, new string) error {
	return os.Rename(old, new)
}
