/*-
 * Copyright 2017 Square Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package starttls

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
)

func dumpAuthTLSFromFTP(dialer Dialer, address string, config *tls.Config) (*tls.ConnectionState, error) {
	c, err := dialer.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	conn := c.(*net.TCPConn)
	status, err := readFTP(conn)
	if err != nil {
		return nil, err
	}
	if status != 220 {
		return nil, fmt.Errorf("FTP server responded with status %d, was expecting 220", status)
	}

	fmt.Fprintf(conn, "AUTH TLS\r\n")
	status, err = readFTP(conn)
	if err != nil {
		return nil, err
	}
	if status != 234 {
		return nil, fmt.Errorf("FTP server responded with status %d, was expecting 234", status)
	}

	tlsConn := tls.Client(conn, config)
	err = tlsConn.Handshake()
	if err != nil {
		return nil, err
	}

	state := tlsConn.ConnectionState()
	return &state, nil
}

func readFTP(conn *net.TCPConn) (int, error) {
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}
	if len(response) <= 3 {
		return 0, fmt.Errorf("Error parsing ftp protocol: Status code too short: '%s'", response)
	}
	return strconv.Atoi(response[:3])
}
