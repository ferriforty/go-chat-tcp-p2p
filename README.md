# go-chat-tcp-p2p: Peer-to-Peer Group Chat

go-chat-tcp-p2p is a simple peer-to-peer group chat system, implemented using TCP sockets, that allows users to connect to chat groups by knowing the port of a peer already in the group. 

## Usage Guide

To use this group chat system, follow these steps: \
(If you have downloaded the .exe file, than just follow these steps by running the .exe instead of calling "go run main.go")

### 1. Clone the Repository

First, clone the repository to your local machine by running:
```bash
git clone https://github.com/ferriforty/go-chat-tcp-p2p
```
### 2. Navigate to the Project Folder

After cloning the repository, navigate to the project folder:
```bash
cd go-chat-tcp-p2p
```
### 3. Start Your Peer

If you want to create your own new group-chat just run the main.go file and specify the port you want to use. It will default to localhost:
```bash
go run main.go <your-port>
```
This will create your own peer on the specified port.
And you will need to wait to someone to join you to start chatting, keep waiting i assure you someone will want to talk to you 😘 (Probably just yourself 💀)
### 4. Join a Group Chat

To join a group chat, you need the port number of a peer that is already part of the chat group (Most of the times yourself in another terminal 😭). You only need one peer’s port to connect to the group.

Once you have the peer’s port, run your peer with both your own port and the peer’s port as parameters:
```bash
go run main.go <your-port> <peer-port>
```
### 5. Enjoy the Chat

Once you connect, everyone else will be notified that you've joined the group chat. You will also receive the complete chat history up to that point. From there, you can start chatting with other members of the group!
Have fun 😆


---

## Features
1. **Decentralized Communication**: 
   - New clients can connect to any existing client in the network.
   - Each client broadcasts its presence, ensuring the chat room is dynamically updated.
2. **Chat History Sharing**:
   - Existing chat history is shared with new clients to provide context.
3. **Color-Coded Output**:
   - Differentiates between messages, system notifications, and user input using ANSI color codes for clarity.
4. **Concurrency**:
   - Utilizes Goroutines for non-blocking message reading, writing, and broadcasting.
5. **Extensibility**:
   - Modular design with clear roles for `Client`, `ChatRoom`, and `Message`.

---

## Implementation Details

### 1. **Client Handling**
- **Purpose**: Each client represents a participant in the chat room.
- **Key Components**:
  - `incoming`: Channel to handle received messages.
  - `outgoing`: Channel to queue messages for sending.
  - `Read()`, `Write()`, `ClientRead()`, and `ClientWrite()` functions to handle message processing.
- **Concurrency**:
  - Goroutines are employed for listening to incoming connections, reading from stdin, and sending messages to the server.

### 2. **Chat Room Management**
- **Purpose**: Hub to track all active clients and broadcast messages.
- **Key Components**:
  - `clients`: A map to manage active clients by their connection addresses.
  - `Broadcast()`: Sends a message to all connected clients.
  - `Join()`: Adds a new client to the chat room.

### 3. **Message Format**
- **Purpose**: Standardized format for chat messages, ensuring clarity and consistency.
- **Implementation**:
  - `Message` struct encapsulates the timestamp, client, and text.
  - ANSI color codes differentiate message components visually.

### 4. **Peer-to-Peer Connectivity**
- **Purpose**: Allows dynamic client discovery and connection.
- **Implementation**:
  - `/join` command broadcasts the client's port and name.
  - `/dial` command connects to new peers.
  - `/chatS` and `/chatR` commands handle chat history sharing.

### 5. **Chat History**
- **Purpose**: Synchronize past messages for new clients.
- **Implementation**:
  - Chat history is serialized and shared using JSON via `/chatS` and `/chatR`.

---

## Why This Design?
1. **Concurrency**: 
   - Go’s Goroutines make handling multiple clients efficient.
2. **Simplicity**:
   - Using the `net` package for TCP abstracts low-level socket operations.
3. **Scalability**:
   - Peer-to-peer architecture supports dynamic client addition.
4. **Readability**:
   - Modular and reusable components enable future enhancements.

