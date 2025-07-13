/*
We need to hook up rpc to run something
an LSP is something that can be started (script or binary) and tells the client "hey send me something"
Once the LSP gets a message, it replies or sends a notification or does whatever
*/

package main

import (
	"bufio"
	"educationalsp/analysis"
	"educationalsp/lsp"
	"educationalsp/rpc"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	logger := getLogger("./log.txt")
	logger.Println("Hey I started :)")

	scanner := bufio.NewScanner(os.Stdin)

	scanner.Split(rpc.Split) // Scanner doesn't know how to read LSP messages -> we need to tell it when to split

	state := analysis.NewState()
	writer := os.Stdout

	for scanner.Scan() {
		fmt.Println("Inside scanner.Scan()")
		msg := scanner.Bytes()
		method, contents, err := rpc.DecodeMessage(msg)
		// Let's not panic and kill everything
		if err != nil {
			logger.Printf("Got an error: %s", err)
			continue
		}

		handleMessage(logger, writer, state, method, contents)
	}
}

// SplitFunc func(data []byte, atEOF bool) (advance int, token []byte, err error)
func handleMessage(logger *log.Logger, writer io.Writer, state analysis.State, method string, contents []byte) {
	logger.Printf("Received msg with method: %s", method)

	switch method {
	case "initialize":
		var request lsp.InitializeRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Printf("Hey, we couldn't parse this: %s", err)
		}

		logger.Printf("Connected to: %s %s", request.Params.ClientInfo.Name, request.Params.ClientInfo.Version)

		// hey... let's reply!
		msg := lsp.NewInitializeResponse(request.ID)
		writeResponse(writer, msg)

	case "textDocument/didOpen":
		var request lsp.DidOpenTextDocumentNotification
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Printf("textDocument/didOpen: %s", err)
			return
		}

		logger.Printf("Opened: %s", request.Params.TextDocument.URI)
		state.OpenDocument(request.Params.TextDocument.URI, request.Params.TextDocument.Text)
	case "textDocument/didChange":
		var request lsp.TextDocumentDidChangeNotification
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Printf("textDocument/didChange: %s", err)
			return
		}

		logger.Printf("Changed: %s", request.Params.TextDocument.URI)

		for _, change := range request.Params.ContentChanges {
			state.UpdateDocument(request.Params.TextDocument.URI, change.Text)
		}

	case "textDocument/hover":
		var request lsp.HoverRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Printf("textDocument/hover: %s", err)
			return
		}

		// Create a response
		response := lsp.HoverResponse{
			Response: lsp.Response{
				RPC: "2.0",
				ID:  &request.ID,
			},
			Result: lsp.HoverResult{
				Contents: "Hello, from LSP!",
			},
		}
		// Write it back
		writeResponse(writer, response)

	}
}

func writeResponse(writer io.Writer, msg any) {
	reply := rpc.EncodeMessage(msg)
	writer.Write([]byte(reply))
}

// Add logging by printing to a file (we're using stdout so we shouldn't use it; stderr is annoying)
func getLogger(filename string) *log.Logger {
	logfile, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		panic("hey, you didn't give me a good file")
	}

	return log.New(logfile, "[educationalsp]", log.Ldate|log.Ltime|log.Lshortfile)
}
