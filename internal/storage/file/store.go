package file

import (
	"bufio"
	"encoding/json"
	"os"
)

type Record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Producer struct {
	file *os.File
}

func NewProducer(filename string) (*Producer, error) {
	// Open the file for appending at the end.
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{file: file}, nil
}

func (p *Producer) Close() error {
	return p.file.Close()
}

func (p *Producer) WriteRecord(record *Record) error {
	data, err := json.Marshal(&record)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	_, err = p.file.Write(data)
	return err
}

type Consumer struct {
	file    *os.File
	scanner *bufio.Scanner
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) ReadRecord() (*Record, error) {
	// Scan one line.
	if !c.scanner.Scan() {
		return nil, c.scanner.Err()
	}
	// Read data from the scanner.
	data := c.scanner.Bytes()

	record := Record{}
	err := json.Unmarshal(data, &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}
