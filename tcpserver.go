package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

type TcpServer struct {
	Cache       *LRUCache
	RedisClient *redis.Client
}

func NewTcpServer(redisClient *redis.Client, cache *LRUCache) *TcpServer {
	return &TcpServer{Cache: cache, RedisClient: redisClient}
}

func (tcpServer *TcpServer) Serve(address string) net.Listener {
	l, err := net.Listen("tcp4", address)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	handler := func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				break
			}
			go tcpServer.handleRequest(conn)
		}
	}
	go handler()
	return l
}

func (tcpServer *TcpServer) handleRequest(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		commands, err := readArray(reader)
		if err != nil {
			_ = writeError(conn, "can't read message")
			_ = conn.Close()
			break
		}

		err = tcpServer.execCommands(conn, commands)
		if err != nil {
			_ = conn.Close()
			break
		}
	}
}

func (tcpServer *TcpServer) execCommands(conn net.Conn, commands []string) error {
	if len(commands) == 0 {
		return writeError(conn, "no commands found")
	}
	command := strings.ToLower(commands[0])
	if command == "ping" {
		return writeSimpleString(conn, "PONG")
	}
	if command == "get" {
		return tcpServer.execGet(conn, commands)
	}

	return writeError(conn, fmt.Sprintf("command %s not supported", command))
}

func (tcpServer *TcpServer) execGet(conn net.Conn, commands []string) error {
	if len(commands) < 2 {
		return writeError(conn, "get command must have key param")
	}

	key := commands[1]
	value, ok := tcpServer.Cache.Get(key)
	if ok {
		message := value.(string)
		return writeBulkString(conn, &message)
	}

	redisVal, err := tcpServer.RedisClient.Get(key).Result()
	if err == redis.Nil {
		return writeBulkString(conn, nil)
	}
	if err != nil {
		return writeError(conn, "can't execute command")
	}

	tcpServer.Cache.Add(key, redisVal)

	return writeBulkString(conn, &redisVal)
}

func readArray(reader *bufio.Reader) ([]string, error) {
	netData, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	lineNumbers, ok := readInteger(netData, '*')
	if !ok {
		return nil, errors.New("not an array")
	}

	var commands []string
	for i := 0; i < lineNumbers; i++ {
		command, err := readBulkString(reader)
		if err != nil {
			return nil, err
		}
		commands = append(commands, command)
	}
	return commands, nil
}

func readBulkString(reader *bufio.Reader) (string, error) {
	netData, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	bytesNumber, ok := readInteger(netData, '$')
	if !ok {
		return "", errors.New("not bulk string")
	}
	//bytesNumber could be -1 in case string is nil. But it shouldn't be nil for keys
	if bytesNumber == -1 {
		return "", errors.New("nil string not supported")
	}

	bytes := make([]byte, bytesNumber+2)

	_, err = io.ReadFull(reader, bytes)
	if err != nil {
		return "", err
	}
	return string(bytes[0 : len(bytes)-2]), nil
}

func readInteger(data string, delim byte) (int, bool) {
	line := strings.TrimSpace(data)
	if len(line) < 2 || line[0] != delim {
		return 0, false
	}
	number, err := strconv.Atoi(line[1:])
	if err != nil {
		return 0, false
	}
	return number, true
}

func writeError(conn net.Conn, message string) error {
	data := fmt.Sprintf("-Error %s\r\n", message)
	_, err := conn.Write([]byte(data))
	return err
}

func writeBulkString(conn net.Conn, message *string) error {
	var data string
	if message != nil {
		bytesNumber := len(*message)
		data = fmt.Sprintf("$%d\r\n%s\r\n", bytesNumber, *message)
	} else {
		data = "$-1\r\n"
	}

	_, err := conn.Write([]byte(data))
	return err
}

func writeSimpleString(conn net.Conn, message string) error {
	data := fmt.Sprintf("+%s\r\n", message)
	_, err := conn.Write([]byte(data))
	return err
}
