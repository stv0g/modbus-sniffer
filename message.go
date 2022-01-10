package main

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

const (
	DirectionRead  Direction = iota
	DirectionWrite Direction = iota
)

type Direction int

type Message struct {
	Time      time.Time
	Pid       int
	Fd        int
	Direction Direction
	Buffer    []byte
}

func ReadMessage(c *csv.Reader) (Message, error) {
	line, err := c.Read()
	if err != nil {
		return Message{}, fmt.Errorf("failed to read line from CSV: %w", err)
	}

	ts, _ := strconv.Atoi(line[0])
	pid, _ := strconv.Atoi(line[1])
	fd, _ := strconv.Atoi(line[2])
	dir := ParseDirection(line[3])
	buf, _ := hex.DecodeString(line[5])

	return Message{
		Time:      time.UnixMilli(int64(ts)),
		Pid:       pid,
		Fd:        fd,
		Direction: dir,
		Buffer:    buf,
	}, nil
}

func (d Direction) String() string {
	switch d {
	case DirectionRead:
		return "read"
	case DirectionWrite:
		return "write"
	}

	return ""
}

func ParseDirection(s string) Direction {
	switch s {
	case "read":
		return DirectionRead
	case "write":
		return DirectionWrite
	}

	return -1
}

func (m *Message) Bytes() []byte {
	b := bytes.Buffer{}

	binary.Write(&b, binary.BigEndian, int32(m.Pid))
	binary.Write(&b, binary.BigEndian, int8(m.Direction))
	b.Write(m.Buffer)

	return b.Bytes()
}

func (m *Message) Write(c *csv.Writer) {
	r := []string{
		fmt.Sprintf("%d", m.Time.UnixMilli()),
		fmt.Sprintf("%d", m.Pid),
		fmt.Sprintf("%d", m.Fd),
		m.Direction.String(),
		fmt.Sprintf("%d", len(m.Buffer)),
		hex.EncodeToString(m.Buffer),
	}

	if err := c.Write(r); err != nil {
		panic(err)
	}

	c.Flush()
}

func (m *Message) String() string {
	return fmt.Sprintf("time=%s, pid=%d, dir=%s, fd=%d, len=%d, buf=%s", m.Time.Format(time.RFC3339), m.Pid, m.Direction.String(), m.Fd, len(m.Buffer), hex.EncodeToString(m.Buffer))
}
