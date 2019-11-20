package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Server struct {
	Addr   string
	Config *ssh.ClientConfig
}

func NewServer(addr string, Config *ssh.ClientConfig) *Server {
	s := &Server{
		Addr:   addr + ":22",
		Config: Config,
	}
	return s
}

func (s *Server) Execute(cmds []string) error {
	log.Println("\nConnecting to", s.Addr)

	client, err := ssh.Dial("tcp", s.Addr, s.Config)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	w, err := session.StdinPipe()
	if err != nil {
		return err
	}
	r, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	req := struct {
		Command string
	}{
		Command: "bash",
	}
	if ok, err := session.SendRequest("exec", true, ssh.Marshal(&req)); err != nil {
		return err
	} else if !ok {
		panic("not ok")
	}
	data := make([]byte, 1024*1024)
	keyword := "NEXT1234567890NEXT"
	for _, cmd := range cmds {
		log.Println(s.Addr, "Execute", cmd)
		if _, err := w.Write([]byte(cmd + "\n")); err != nil {
			return err
		}
		if _, err := w.Write([]byte("echo " + keyword + "\n")); err != nil {
			return err
		}
		idx := 0
		for {
			n, err := r.Read(data[idx:])
			if err != nil {
				return err
			}
			idx += n
			if strings.HasSuffix(string(data[:idx]), keyword+"\n") {
				break
			}
		}
		log.Println(s.Addr, "Response", string(data[:idx-len(keyword)-1]))
	}
	return nil
}

func (s *Server) Output(cmds []string) ([]byte, error) {
	//log.Println("\nConnecting to", s.Addr)

	client, err := ssh.Dial("tcp", s.Addr, s.Config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	w, err := session.StdinPipe()
	if err != nil {
		return nil, err
	}
	r, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}
	req := struct {
		Command string
	}{
		Command: "bash",
	}
	if ok, err := session.SendRequest("exec", true, ssh.Marshal(&req)); err != nil {
		return nil, err
	} else if !ok {
		panic("not ok")
	}
	data := make([]byte, 1024*1024)
	keyword := "NEXT1234567890NEXT"

	var buffer bytes.Buffer
	for _, cmd := range cmds {
		//log.Println("Execute", cmd)
		if _, err := w.Write([]byte(cmd + "\n")); err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte("echo " + keyword + "\n")); err != nil {
			return nil, err
		}
		idx := 0
		for {
			n, err := r.Read(data[idx:])
			if err != nil {
				return nil, err
			}
			idx += n
			if strings.HasSuffix(string(data[:idx]), keyword+"\n") {
				break
			}
		}
		//log.Println("Response", string(data[:idx-len(keyword)-1]))
		buffer.Write(data[:idx-len(keyword)-1])
	}
	return buffer.Bytes(), nil
}

func (s *Server) Open() (*ServerConn, error) {
	//log.Println("\nConnecting to", s.Addr)

	c := &ServerConn{
		Addr:   s.Addr,
		Config: s.Config,
	}
	if err := c.Open(); err != nil {
		return nil, err
	}
	return c, nil
}

type ServerConn struct {
	Addr    string
	Config  *ssh.ClientConfig
	client  *ssh.Client
	session *ssh.Session
	w       io.Writer
	r       io.Reader
}

func (c *ServerConn) Open() error {
	client, err := ssh.Dial("tcp", c.Addr, c.Config)
	if err != nil {
		c.Close()
		return err
	}
	c.client = client

	session, err := client.NewSession()
	if err != nil {
		c.Close()
		return err
	}
	c.session = session

	w, err := session.StdinPipe()
	if err != nil {
		c.Close()
		return err
	}
	c.w = w

	r, err := session.StdoutPipe()
	if err != nil {
		c.Close()
		return err
	}
	c.r = r
	go func() {
		e, err := session.StderrPipe()
		if err != nil {
			return
		}
		for {
			data := make([]byte, 1024*1024)
			n, err := e.Read(data)
			if err != nil {
				return
			}
			fmt.Print(string(data[:n]))
		}
	}()

	req := struct {
		Command string
	}{
		Command: "bash",
	}
	if ok, err := session.SendRequest("exec", true, ssh.Marshal(&req)); err != nil {
		c.Close()
		return err
	} else if !ok {
		c.Close()
		panic("not ok")
	}
	return nil
}

func (c *ServerConn) Close() {
	if c.client != nil {
		c.client.Close()
		c.client = nil
	}
	if c.session != nil {
		c.session.Close()
		c.session = nil
	}
}

func (c *ServerConn) Do(cmd string) ([]byte, error) {
	data := make([]byte, 1024*1024)
	keyword := "NEXT1234567890NEXT"

	var buffer bytes.Buffer
	log.Println(c.Addr, "Execute", cmd)
	if _, err := c.w.Write([]byte(cmd + "\n")); err != nil {
		return nil, err
	}
	if _, err := c.w.Write([]byte("echo " + keyword + "\n")); err != nil {
		return nil, err
	}
	idx := 0
	for {
		n, err := c.r.Read(data[idx:])
		if err != nil {
			return nil, err
		}
		idx += n
		if strings.HasSuffix(string(data[:idx]), keyword+"\n") {
			break
		}
	}
	log.Println(c.Addr, "Response", string(data[:idx-len(keyword)-1]))
	buffer.Write(data[:idx-len(keyword)-1])
	return buffer.Bytes(), nil
}

func (c *ServerConn) Output(cmd string) ([]byte, error) {
	data := make([]byte, 1024*1024)
	keyword := "NEXT1234567890NEXT"

	var buffer bytes.Buffer
	//log.Println("Execute", cmd)
	if _, err := c.w.Write([]byte(cmd + "\n")); err != nil {
		return nil, err
	}
	if _, err := c.w.Write([]byte("echo " + keyword + "\n")); err != nil {
		return nil, err
	}
	idx := 0
	for {
		n, err := c.r.Read(data[idx:])
		if err != nil {
			return nil, err
		}
		idx += n
		if strings.HasSuffix(string(data[:idx]), keyword+"\n") {
			break
		}
	}
	//log.Println("Response", string(data[:idx-len(keyword)-1]))
	buffer.Write(data[:idx-len(keyword)-1])
	return buffer.Bytes(), nil
}
