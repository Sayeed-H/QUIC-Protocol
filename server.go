package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

func main() {
	// Generate a self-signed certificate for testing
	tlsConf := generateTLSConfig()

	// Listen on localhost:4242
	listener, err := quic.ListenAddr("localhost:4242", tlsConf, nil)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}
	defer listener.Close()

	fmt.Println("🚀 QUIC Server listening on localhost:4242")
	fmt.Println("📡 Waiting for connections...")

	for {
		// Accept a QUIC connection
		conn, err := listener.Accept(context.Background())
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		fmt.Printf("🔗 New connection from %s\n", conn.RemoteAddr())

		// Handle connection in a goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn *quic.Conn) {
	defer conn.CloseWithError(0, "")

	for {
		// Accept a stream from the client
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			fmt.Printf("❌ Connection closed: %v\n", err)
			return
		}

		fmt.Printf("📋 New stream %d opened\n", stream.StreamID())

		// Handle stream in goroutine
		go handleStream(stream)
	}
}

func handleStream(stream *quic.Stream) {
	defer stream.Close()

	// Read data from client
	buffer := make([]byte, 1024)
	n, err := stream.Read(buffer)
	if err != nil && err != io.EOF {
		fmt.Printf("❌ Error reading from stream: %v\n", err)
		return
	}

	message := string(buffer[:n])
	fmt.Printf("📨 Received: %s\n", message)

	// Echo back with a prefix
	response := fmt.Sprintf("Echo: %s", message)
	_, err = stream.Write([]byte(response))
	if err != nil {
		fmt.Printf("❌ Error writing to stream: %v\n", err)
		return
	}

	fmt.Printf("📤 Sent: %s\n", response)
}

// Generate a self-signed certificate for testing
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"QUIC Learning Lab"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:     []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		log.Fatal(err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Fatal(err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-learning-lab"},
	}
}