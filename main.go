package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)


const (
	CONN_TYPE = "tcp"

	CMD_PREFIX = "/"
	CMD_JOIN   = CMD_PREFIX + "join"
	CMD_DIAL   = CMD_PREFIX + "dial"
	CMD_LEAVE  = CMD_PREFIX + "leave"
	CMD_QUIT   = CMD_PREFIX + "quit"
	CMD_INIT   = CMD_PREFIX + "init"
)


type Client struct {
	name string
	chatRoom *ChatRoom
	incoming chan *Message
	outgoing chan string
	conn net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

func (client *Client) Quit() {
	client.conn.Close()
}

func NewClient(chatRoom *ChatRoom, conn net.Conn) *Client {
	client := &Client {
		name:     "",
		chatRoom: chatRoom,
		incoming: make(chan *Message),
		outgoing: make(chan string),
		conn:     conn,
		reader:   nil,
		writer:   nil,
	}
	client.reader = bufio.NewReader(conn)
	client.writer = bufio.NewWriter(conn)
	client.Listen()
	return client
}

// Reads and parse the string received, if it has some prefix than it does some function. if not, it prints it
func (client *Client) Read() {
	go func() {
		for message := range client.incoming {
			switch {
			default:
				fmt.Print(message.String())
			case strings.HasPrefix(message.text, CMD_INIT):
				client.name = strings.TrimSpace(strings.TrimPrefix(message.text, CMD_INIT+" "))
			case strings.HasPrefix(message.text, CMD_JOIN):
				client.chatRoom.Broadcast(strings.TrimSpace(strings.TrimPrefix(message.text, CMD_JOIN+" ")) + "\n")
			case strings.HasPrefix(message.text, CMD_DIAL):
				port := strings.TrimSpace(strings.TrimPrefix(message.text, CMD_DIAL+" "))
				if client.chatRoom.connPort == port {
					break
				}
				// i check if the client connecting is already present in my clients. if it is not, the i add it.
				for _, x := range client.chatRoom.clients {
					conn := x.conn.LocalAddr().String()
					presentPort := strings.Split(conn, ":")[1]
					if port != (":" + presentPort) {
						conn, err := net.Dial(CONN_TYPE, port)
						if err != nil {
							fmt.Println(err)
						}
						client := NewClient(x.chatRoom, conn)
						x.chatRoom.Join(client)
						break
					}
				}
			}
		}
	}()
}

// sends the message to all the clients connected
func (chatRoom *ChatRoom) Broadcast(message string) {
	for _, client := range chatRoom.clients {
		client.outgoing <- message
	}
}

// if a message is inserted in the outgoing field, than it is printed
func (client *Client) Write() {
	for message := range client.outgoing {

		_, err := client.writer.WriteString(message)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = client.writer.Flush()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

// starts all the go-routines
func (client *Client) Listen() {
	go client.Read()
	go client.Write()
	go client.ClientRead(client.conn)
	go client.ClientWrite(client.conn)
}

func NewMessage(time time.Time, client *Client, text string) *Message {
	return &Message{
		time:   time,
		client: client,
		text:   text,
	}
}

type ChatRoom struct {
	myName string
	clients  []*Client
	incoming chan *Message
	connPort string
	join chan *Client
}

func NewChatRoom(connPort string) *ChatRoom {

	chatRoom := &ChatRoom{
		myName: "",
		clients: make([]*Client, 0),
		incoming: make(chan *Message),
		connPort: connPort,
		join: make(chan *Client),
	}

	return chatRoom
}

// add client to the chatroom
func (chatRoom *ChatRoom) Join(client *Client) {
	chatRoom.clients = append(chatRoom.clients, client)
}

type Message struct {
	time time.Time
	client *Client
	text string
}

// format the string to send
func (message *Message) String() string {
	return fmt.Sprintf("%s - %s: %s\n", message.time.Format(time.Kitchen), message.client.name, message.text)
}

// Reads all incoming messages and put them in the incoming field.
func (client *Client)ClientRead(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		message := NewMessage(time.Now(), client, strings.TrimSuffix(str, "\n"))
		client.incoming <- message
	}
}

// Reads from Stdin, and outputs to the socket.
func (client *Client)ClientWrite(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	client.chatRoom.Broadcast(CMD_INIT + " " + client.chatRoom.myName)

	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		client.chatRoom.Broadcast(str)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Checks if the params are correct
	if len(os.Args) <= 1 {
		log.Println("required server port as parameter 1 and port to connect to as parameter 2(optional)")
		os.Exit(1)
	}

	// Takes connection port and creates the listener, assuming the connection port given is correct (:xxxx)
	connPort := os.Args[1]
	chatRoom := NewChatRoom(connPort)

	listener, err := net.Listen(CONN_TYPE, "localhost" + connPort)
	if err != nil {
		log.Println("Error", err)
		os.Exit(1)
	}
	defer listener.Close()
	log.Println("listening on port", "localhost" + connPort)
	
	reader := bufio.NewReader(os.Stdin)
	
	// Before doing anything i wait for the user name 
	fmt.Print("Enter name: ")
	text, err := reader.ReadString('\n')
	if err != nil {
		log.Println(err)
	}

	chatRoom.myName = text

	// If the user putted a server to connect to i connect to it and broadcast to everyone the port, to let everyone know there is a new member
	if len(os.Args) >= 3 {

		conn, err := net.Dial(CONN_TYPE, os.Args[2])
		if err != nil {
			fmt.Println(err)
		}
		client := NewClient(chatRoom, conn)
		chatRoom.Join(client)
		client.outgoing <- CMD_JOIN + " " + CMD_DIAL + " " + connPort + "\n"
	}

	// Waits for a new client to connect.
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error: ", err)
			continue
		}
		chatRoom.Join(NewClient(chatRoom, conn))
	}
}