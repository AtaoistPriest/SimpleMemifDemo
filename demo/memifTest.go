/*
 *------------------------------------------------------------------
 * Copyright (c) 2020 Cisco and/or its affiliates.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *------------------------------------------------------------------
 */

package main

import (
	"fmt"
	"log"
	"me/memif"
	"time"
)

func Connected(i *memif.Interface) error {
	fmt.Println(i.GetName(), "Connected successfully")
	return nil
}

func DisConnected(i *memif.Interface) error {
	fmt.Println(i.GetName(), "DisConnected")
	return nil
}

func main() {
	name := "int0"
	socketName := "/run/vpp/memif.sock"

	fmt.Println("GoMemif: Responder")
	fmt.Println("-----------------------")

	socket, err := memif.NewSocket("forward", socketName)
	if err != nil {
		fmt.Println("Failed to create socket: ", err)
		return
	}

	args := &memif.Arguments{
		IsMaster:         false,   //slave
		ConnectedFunc:    Connected,
		DisconnectedFunc: DisConnected,
		Mode: 			  memif.InterfaceModeIp,
		Name:             name,
	}

	i, err := socket.NewInterface(args)
	if err != nil {
		fmt.Println("Failed to create interface on socket %s: %s", socket.GetFilename(), err)
		socket.Delete()
		return
	}

	// slave attempts to connect to control socket
	// to handle control communication call socket.StartPolling()
	if !i.IsMaster() {
		fmt.Println(name, ": Connecting to control socket...")
		for !i.IsConnecting() {
			err = i.RequestConnection()
			if err != nil {
				/* TODO: check for ECONNREFUSED errno
				 * if error is ECONNREFUSED it may simply mean that master
				 * interface is not up yet, use i.RequestConnection()
				 */
				fmt.Println("Faild to connect: ", err)
				socket.Delete()
				return
			}
		}
	}
	memifErrChan := make(chan error)

	socket.StartPolling(memifErrChan)
	for !i.IsConnected(){
		time.Sleep(time.Millisecond)
	}
	// allocate packet buffer
	pkt := make([]byte, 2048)
	// get rx queue
	rxq0, err := i.GetRxQueue(0)
	if err != nil {
		log.Print("get rx Queue : ", err)
		return
	}
	// get tx queue
	//txq0, err := i.GetTxQueue(0)
	//if err != nil {
	//	return
	//}
	log.Print("memif reader is ready!")
	for {
		pktLen, _ := rxq0.ReadPacket(pkt)
		if pktLen > 0{
			log.Print("data : ", pkt[:pktLen])
		}
		//	txq0.WritePacket(pkt[:pktLen])
		time.Sleep(time.Millisecond * 100)
	}
}
