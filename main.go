package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

const (
	CONN_PORT = ":3333"
	CONN_TYPE = "tcp"

	CMD_PREFIX = "/"
	CMD_CREATE = CMD_PREFIX + "create"
	CMD_LIST   = CMD_PREFIX + "list"
	CMD_JOIN   = CMD_PREFIX + "join"
	CMD_LEAVE  = CMD_PREFIX + "leave"
	CMD_HELP   = CMD_PREFIX + "help"
	CMD_NAME   = CMD_PREFIX + "name"
	CMD_QUIT   = CMD_PREFIX + "quit"

	MSG_CONNECT = "Welcome to the server! Type \"/help\" to get a list of commands.\n"
	MSG_FULL    = "Server is full. Please try reconnecting later."
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

func NewClient(conn net.Conn) *Client {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	stdReader := bufio.NewReader(os.Stdin)
    log.Print("Enter name: ")
    text, err := stdReader.ReadString('\n')
	if err != nil {
		log.Println(err)
	}

	client := &Client {
		name:     text,
		chatRoom: nil,
		incoming: make(chan *Message),
		outgoing: make(chan string),
		conn:     conn,
		reader:   reader,
		writer:   writer,
	}
	client.Listen()
	return client
}

func (client *Client) Listen() {
	go client.Read()
	go client.Write()
}

func (client *Client) Read() {
	for {
		str, err := client.reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			break
		}
		client.incoming <- &Message{
			time: time.Now(),
			client: client,
			text: str,
		}
	}
	log.Println("client reader closed")
	close(client.incoming)
}

func (client *Client) Write() {
	for str := range client.outgoing {
		_, err := client.writer.WriteString(str)
		if err != nil {
			log.Println(err)
			break
		}
		err = client.writer.Flush()
		if err != nil {
			log.Println(err)
			break
		}
	}
	log.Println("client writer closed")
}

func (client *Client) Quit() {
	client.conn.Close()
}

type ChatRoom struct {
	clients  []*Client
	incoming chan *Message
	messages []*Message
	join chan *Client
}

func NewChatRoom() *ChatRoom {

	chatRoom := &ChatRoom{
		clients: make([]*Client, 0),
		incoming: make(chan *Message),
		messages: make([]*Message, 0),
		join: make(chan *Client),
	}

	chatRoom.Listen()
	return chatRoom
}

func (chatRoom *ChatRoom) Listen() {
	go func() {
		for {
			select {
			case message := <- chatRoom.incoming:
				chatRoom.SendMessage(message)
			case client := <-chatRoom.join:
				chatRoom.Join(client)
			}
			
		}
	}()
}

func (chatRoom *ChatRoom) Join(client *Client) {
	chatRoom.clients = append(chatRoom.clients, client)
	client.outgoing <- MSG_CONNECT
	go func() {
		for message := range client.incoming {
			chatRoom.incoming <- message
		}
	}()
}

func (chatRoom *ChatRoom) SendMessage(message *Message) {
	message.client.chatRoom.Broadcast(message)
}

func (chatRoom *ChatRoom) Broadcast(message *Message) {
	message.text = time.Now().String()
	chatRoom.messages = append(chatRoom.messages, message)
	for _, client := range chatRoom.clients {
		client.outgoing <- message.String()
	}
}

type Message struct {
	time time.Time
	client *Client
	text string
}

func (message *Message) String() string {
	return fmt.Sprintf("%s - %s: %s\n", message.time.Format(time.Kitchen), message.client.name, message.text)
}

func main() {

	chatRoom := NewChatRoom()

	listener, err := net.Listen(CONN_TYPE, CONN_PORT)
	if err != nil {
		log.Println("Error", err)
		os.Exit(1)
	}
	defer listener.Close()
	log.Println("listening on port", CONN_PORT)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error: ", err)
			continue
		}
		chatRoom.Join(NewClient(conn))
	}
}