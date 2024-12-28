package manty_dns

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

// DNSHeader represents the DNS packet header
type DNSHeader struct {
	ID      uint16
	Flags   uint16
	QDCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}

// DNSQuestion represents a DNS question
type DNSQuestion struct {
	Name  []byte
	Type  uint16
	Class uint16
}

// DNSAnswer represents a DNS answer
type DNSAnswer struct {
	Name  uint16
	Type  uint16
	Class uint16
	TTL   uint32
	Len   uint16
	IP    [4]byte
}

func Start(port int, ip string) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(ip),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("Failed to set up UDP listener: %v", err)
	}
	defer conn.Close()

	fmt.Println("DNS server is running on port 53")

	for {
		buffer := make([]byte, 512)
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Failed to read UDP packet: %v", err)
			continue
		}

		// Handle each request in a separate goroutine
		go handleRequest(conn, buffer[:n], addr)
	}
}

func handleRequest(conn *net.UDPConn, buffer []byte, addr *net.UDPAddr) {
	var header DNSHeader
	reader := bytes.NewReader(buffer)
	err := binary.Read(reader, binary.BigEndian, &header)
	if err != nil {
		log.Printf("Failed to parse DNS header: %v", err)
		return
	}

	// Prepare the response
	response, err := createResponse(header, buffer)
	if err != nil {
		log.Printf("Failed to create UDP response: %v", err)
	}

	// Send the response
	_, err = conn.WriteToUDP(response, addr)
	if err != nil {
		log.Printf("Failed to send UDP response: %v", err)
	}
}

func parseQuestion(data []byte) (DNSQuestion, int) {
	var question DNSQuestion
	var offset int

	// Read the domain name
	for {
		length := int(data[offset])
		if length == 0 {
			offset++
			break
		}
		offset += length + 1
	}

	question.Name = data[:offset]
	question.Type = binary.BigEndian.Uint16(data[offset : offset+2])
	question.Class = binary.BigEndian.Uint16(data[offset+2 : offset+4])
	offset += 4

	return question, offset
}

func createResponse(header DNSHeader, request []byte) ([]byte, error) {
	var response bytes.Buffer

	// Write DNS header
	header.Flags = 0x8180           // Standard query response, no error
	header.ANCount = header.QDCount // One answer per question
	binary.Write(&response, binary.BigEndian, header)

	// Offset to start of questions
	offset := 12

	// Process each question
	for i := 0; i < int(header.QDCount); i++ {
		question, questionEnd := parseQuestion(request[offset:])
		offset += questionEnd

		// Write original question
		response.Write(request[offset-questionEnd : offset])

		// Write answer
		answer := DNSAnswer{
			Name:  0xC00C, // Pointer to the domain name in the question
			Type:  question.Type,
			Class: question.Class,
			TTL:   600,
			Len:   4,
			IP:    [4]byte{192, 0, 2, 1}, // Example IP address
		}
		err := binary.Write(&response, binary.BigEndian, answer)
		if err != nil {
			return nil, err
		}
	}

	return response.Bytes(), nil
}
