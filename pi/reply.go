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

import "fmt"

/*
 * Reply Codes:
 *
 * 1yz = Positive Preliminary reply.
 * 2yz = Positive Completion reply.
 * 3yz = Positive Intermediate reply.
 * 4yz = Transient Negative Completion reply.
 * 5yz = Permanent Negative Completion reply.
 */
type replyCode int

const (
	replyFileStatusOkay replyCode = 150

	replyOkay           replyCode = 200
	replySystemStatus   replyCode = 211
	replyFileStatus     replyCode = 213
	replyHello          replyCode = 220
	replyCloseDTP       replyCode = 226
	replyPASVOkay       replyCode = 227
	replyEPSVOkay       replyCode = 229
	replyLoggedIn       replyCode = 230
	replyFileActionOkay replyCode = 250
	replyPathnameOkay   replyCode = 257

	replyUserNameOkay      replyCode = 331
	replyFileActionPending replyCode = 350

	replyFailedOpenDTP replyCode = 425

	replyNotFoundCommand       replyCode = 500
	replyInvalidParameter      replyCode = 501
	replyNotSupportedCommand   replyCode = 502
	replyNotSupportedParameter replyCode = 504
	replyNotSupportedNetwork   replyCode = 522
	replyNotLoggedIn           replyCode = 530
	replyUnavailableFile       replyCode = 550
)

type reply struct {
	code      replyCode
	message   string
	multiline bool
}

func (r *reply) make() string {
	if r.multiline == true {
		return fmt.Sprintf("%v-%v\r\n%v END\r\n", r.code, r.message, r.code)
	}
	return fmt.Sprintf("%v %v\r\n", r.code, r.message)
}
