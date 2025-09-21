package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/quic-go/quic-go"
)

func main() {
	fmt.Println("🔌 Connecting to QUIC server...")

	// Configure TLS to accept self-signed certificates (for testing only!)
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-learning-lab"},
	}

	// Connect to the server
	conn, err := quic.DialAddr(context.Background(), "localhost:4242", tlsConf, nil)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer conn.CloseWithError(0, "client done")

	fmt.Printf("✅ Connected to %s\n", conn.RemoteAddr())

	// Demonstrate multiple streams
	for i := 1; i <= 3; i++ {
		fmt.Printf("\n🔄 Creating stream %d...\n", i)
		
		// Open a new stream
		stream, err := conn.OpenStreamSync(context.Background())
		if err != nil {
			log.Fatal("Failed to open stream:", err)
		}

		// Send a message
		message := fmt.Sprintf("Hello from stream %d! Time: %v", i, time.Now().Format("15:04:05"))
		fmt.Printf("📤 Sending: %s\n", message)
		
		_, err = stream.Write([]byte(message))
		if err != nil {
			log.Fatal("Failed to send message:", err)
		}

		// Close the write side to signal we're done sending
		stream.Close()

		// Read the response
		buffer := make([]byte, 1024)
		n, err := stream.Read(buffer)
		if err != nil && err != io.EOF {
			log.Fatal("Failed to read response:", err)
		}

		if n > 0 {
			response := string(buffer[:n])
			fmt.Printf("📨 Received: %s\n", response)
		}
		
		// Wait a bit between streams to see the multiplexing
		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n🎉 All streams completed!")
}