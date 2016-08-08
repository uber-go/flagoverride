// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package flags

import (
	"fmt"
	"time"
)

func ExampleParseArgs() {
	type Logging struct {
		Interval int
		Path     string
	}
	type Socket struct {
		ReadTimeout  time.Duration `yaml:"read_timeout"`
		WriteTimeout time.Duration
	}

	type TCP struct {
		ReadTimeout time.Duration
		Socket
	}

	type Network struct {
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
		TCP
	}

	type Cfg struct {
		Logging
		Network
	}

	// this is just an example, normally one would use packages like yaml to
	// populate the struct rather than manually create it.
	c := &Cfg{
		Logging: Logging{Interval: 3, Path: "/tmp"},
		Network: Network{
			TCP: TCP{
				ReadTimeout: time.Duration(10) * time.Millisecond,
				Socket: Socket{
					ReadTimeout: time.Duration(10) * time.Millisecond,
				},
			},
		},
	}

	fmt.Printf("loaded config 'network.tcp.socket.read_timeout' is %s\n", c.Network.TCP.Socket.ReadTimeout)
	args := []string{"--logging.interval", "2", "--network.tcp.socket.read_timeout", "50ms"}
	ParseArgs(c, args)
	fmt.Printf("after override 'network.tcp.socket.read_timeout' is %s\n", c.Network.TCP.Socket.ReadTimeout)
	// Output:
	// loaded config 'network.tcp.socket.read_timeout' is 10ms
	// after override 'network.tcp.socket.read_timeout' is 50ms
}
