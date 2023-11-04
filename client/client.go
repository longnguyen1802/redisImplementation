package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"golangredis/datastructure"
	"net"
	"os"
	"strings"
)
type Request struct {
    Command string
    Key     string
    Value   string
}

func main() {
    serverAddr := "localhost:6379"

    conn, err := net.Dial("tcp", serverAddr)
    if err != nil {
        fmt.Println("Error connecting to the server:", err)
        return
    }
    defer conn.Close()

    fmt.Println("Connected to the server.")

    // Create a buffered reader to read responses from the server
    reader := bufio.NewReader(conn)

    // Create a buffered reader to read input from the user
    userInput := bufio.NewReader(os.Stdin)

    for {
        // Read input from the user
        fmt.Print("Enter a command (e.g., SET key value, GET key, DEL key, or QUIT to close the session): ")
        input, err := userInput.ReadString('\n')
        if err != nil {
            fmt.Println("Error reading input:", err)
            return
        }

        input = strings.TrimSpace(input)

		command, key, value, err := parseCommand(input)
		if err != nil {
			fmt.Println(err)
			continue
		}

        // Construct the request
        request := Request{Command: command, Key: key, Value: value}

        // Convert the command to uppercase
        inputUpper := strings.ToUpper(command)

        if inputUpper == "QUIT" {
            // The user wants to gracefully close the session
            closeSession(conn)
            fmt.Println("Session closed.")
            return
        }
        requestData, err := json.Marshal(request)
        if err != nil {
            fmt.Println("Error marshaling request:", err)
            return
        }

        _, err = conn.Write(append(requestData, '\n'))
		//fmt.Println(requestData)
        if err != nil {
            fmt.Println("Error writing to the server:", err)
            return
        }

        // Read and deserialize the server's response
        responseData, err := reader.ReadString('\n')
        if err != nil {
            fmt.Println("Error reading response from the server:", err)
            return
        }

        var response datastructure.Response
        if err := json.Unmarshal([]byte(responseData), &response); err != nil {
            fmt.Println("Error unmarshaling response:", err)
            return
        }

        fmt.Println("Server response:", response.Message)
    }
}

func closeSession(conn net.Conn) {
    // Send a special command to the server to request session closure
    request := Request{Command: "QUIT"}
    requestData, err := json.Marshal(request)
    if err != nil {
        fmt.Println("Error marshaling request:", err)
        return
    }

    _, err = conn.Write(requestData)
    if err != nil {
        fmt.Println("Error writing to the server:", err)
        return
    }
}

func parseCommand(input string) (string, string, string, error) {
    // Split the input by spaces
    parts := strings.Fields(input)

    // Validate the number of parts
    if len(parts) > 3 || len(parts) < 2 {
        return "", "", "", fmt.Errorf("invalid command")
    }

    command := parts[0]
    key := ""
    value := ""

    if len(parts) == 3 {
        key = parts[1]
        value = parts[2]
    } else if len(parts) == 2 {
        if command == "AUTH" {
            // For AUTH command, the single argument is the value
            value = parts[1]
        } else if command == "GET" {
            // For GET command, the argument is the key
            key = parts[1]
        } else {
            return "", "", "", fmt.Errorf("invalid command")
        }
    }

    return command, key, value, nil
}