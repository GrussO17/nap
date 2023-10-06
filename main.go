package main

import (
	"flag"
	"fmt"
	"net"
	"github.com/fatih/color"
)

type printer func(format string, a ...interface{})

func injector(out_pipe chan<- []byte, laddr string, log printer) {
	list, err := net.Listen("tcp", laddr)
	if err != nil {
		color.HiRed("Fialed to bind to: ", laddr)
		return
	}
	conn, err := list.Accept()
	if err != nil {
		color.HiRed("Fialed to accept")
		return
	}
	go reader(out_pipe, conn, log)
}

func reader(out_pipe chan<- []byte, conn net.Conn, log printer) {
	for {
		bytes := make([]byte, 2048)
		_, err := conn.Read(bytes)
		if err != nil {
			color.HiRed("Failed to read")
			return
		}
		log(string(bytes))
		out_pipe <- bytes
	}
}

func writer(in_pipe <-chan []byte, conn net.Conn) {
	for {
		conn.Write(<-in_pipe)
	}
}

func main() {
	var forward_ip = flag.String("f", "127.0.0.1", "Address to forward traffic to")
	var forward_port = flag.Int("p", 8888, "Port to forward traffic to")
	var local_port = flag.Int("l", 9999, "Port to listen for connections on")
	var server_inject_port = flag.Int("forward-inject", 9000, "Port that will have traffic forwarded to the forward address")
	var client_inject_port = flag.Int("client-inject", 9001, "Port that will have traffic forwarded to the client (connected) address")
	var local_ip = flag.String("local-interface", "127.0.0.1", "Interface to ")
	flag.Parse()
	forward_addr := *forward_ip + ":" + fmt.Sprint(*forward_port)
	reverse_addr := *local_ip + ":" + fmt.Sprint(*local_port)
	forward_inject_addr := *local_ip + ":" + fmt.Sprint(*server_inject_port)
	reverse_inject_addr := *local_ip + ":" + fmt.Sprint(*client_inject_port)

	color.HiCyan("Guide to the colors")
	color.HiCyan("------------------------------------")
	color.HiRed("Errors will be printed in Red")
	color.Green("Client(%s) -> Target(%s)", reverse_addr, forward_addr)
	color.Magenta("Target(%s) -> Client(%s)", forward_addr, reverse_addr)
	color.Yellow("Injector(%s) -> Target(%s)", forward_inject_addr, forward_addr)
	color.Blue("Injector(%s) -> Client(%s)", reverse_inject_addr, reverse_addr)
	color.HiCyan("------------------------------------")

	forward_pipe := make(chan []byte, 4096)
	reverse_pipe := make(chan []byte, 4096)

	forward_conn, err := net.Dial("tcp", forward_addr)
	if err != nil {
		color.HiRed("Failed to connect to %s", forward_addr)
		return
	}
	color.HiRed("Waiting for client to connect")
	list, err := net.Listen("tcp", reverse_addr)
	if err != nil {
		color.Red("Fialed to bind to: ", reverse_addr)
		return
	}
	reverse_conn, err := list.Accept()
	if err != nil {
		fmt.Println("Fialed to accept")
		return
	}
	color.HiRed("Starting")
	go injector(forward_pipe, forward_inject_addr, color.Yellow)
	go injector(reverse_pipe, reverse_inject_addr, color.Blue)
	go reader(forward_pipe, reverse_conn, color.Green)
	go reader(reverse_pipe, forward_conn, color.Magenta)
	go writer(forward_pipe, forward_conn)
	go writer(reverse_pipe, reverse_conn)
	select {} // fancy way of waiting forever
}
