package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"golangredis/datastructure"
	"net"
	"strings"
	"sync"
)

type Request struct {
    Command string
    Key     string
    Value   string
}

var authPassword = "redis_password" // Change this to your desired password

type ClientSession struct {
    Conn      net.Conn
    DataStore datastructure.DataStore
    Authenticated bool
    Mutex sync.Mutex
}

func createClientSession(conn net.Conn, dataStore datastructure.DataStore) *ClientSession {
    return &ClientSession{
        Conn:      conn,
        DataStore: dataStore,
        Authenticated: false,
    }
}

func handleSession(session *ClientSession) {
    conn := session.Conn

    defer conn.Close()
    clientAddr := conn.RemoteAddr()
    fmt.Printf("Accepted connection from %s\n", clientAddr)

    reader := bufio.NewReader(conn)

    for {
        requestJSON, err := reader.ReadString('\n')
        if err != nil {
            fmt.Printf("Connection from %s closed\n", clientAddr)
            return
        }
		//fmt.Println(requestJSON)
        var request Request
        if err := json.Unmarshal([]byte(requestJSON), &request); err != nil {
            fmt.Printf("Error unmarshaling request: %v\n", err)
            return
        }
		//fmt.Println(request)
		//fmt.Println(request.Value)
        if !session.Authenticated {
            if request.Command == "AUTH" {
                providedPassword := strings.TrimSpace(request.Value)
                if providedPassword == authPassword {
                    session.Authenticated = true
                    response := datastructure.Response{Message: "OK", Successful: true}
                    sendResponse(conn, response)
                } else {
                    response := datastructure.Response{Message: "Authentication failed", Successful: false}
                    sendResponse(conn, response)
                }
            } else {
                response := datastructure.Response{Message: "Please authenticate first", Successful: false}
                sendResponse(conn, response)
            }
        } else {
            // Client is authenticated, proceed with processing requests
			processRequest:= datastructure.Request{Command:request.Command,Key:request.Key,Value:request.Value}
            response := session.DataStore.ProcessRequest(processRequest)
            sendResponse(conn, response)
        }
    }
}

func sendResponse(conn net.Conn, response datastructure.Response) {
    responseJSON, err := json.Marshal(response)
    if err != nil {
        fmt.Printf("Error marshaling response: %v\n", err)
        return
    }
	_, err = conn.Write(append(responseJSON, '\n'))
    //_, err = conn.Write(responseJSON)
    if err != nil {
        fmt.Printf("Error writing response to %s: %v\n", conn.RemoteAddr(), err)
    }
}


func StartServer() {
    listener, err := net.Listen("tcp", "localhost:6379")
    if err != nil {
        fmt.Println("Error listening:", err)
        return
    }
    defer listener.Close()
    fmt.Println("Server is listening on localhost:6379")

    dataStore := datastructure.NewKeyValueStoreWithWorker(4) // Adjust the number of workers as needed

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting connection:", err)
            continue
        }

        // Create a new session for the client
        session := createClientSession(conn, dataStore)

        // Handle the session in a Goroutine
        go handleSession(session)
    }
}