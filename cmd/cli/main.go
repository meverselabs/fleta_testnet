package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var (
	ErrNotExist = errors.New("not exist")
)

type Info struct {
	ServerName string
	ServerIP   string
	LocalIP    string
	Region     string
	Type       string
	Name       string
	Address    string
	Gen        string
	GenHash    string
	Key        string
	KeyHash    string
	Conn       *ServerConn
}

func GetBalance(c *websocket.Conn, addr string) (string, error) {
	res, err := DoRequest2(c, "vault.balance", []interface{}{addr})
	if err != nil {
		return "", err
	} else {
		bs, err := json.MarshalIndent(res, "", "\t")
		if err != nil {
			return "", err
		} else {
			return string(bs), nil
		}
	}
}

func main() {
	pems, err := ioutil.ReadFile("testnet.pem")
	if err != nil {
		panic(err)
	}
	priv, err := ssh.ParsePrivateKey(pems)
	if err != nil {
		panic(err)
	}
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{ssh.PublicKeys(priv)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	body, err := ioutil.ReadFile("data.txt")
	if err != nil {
		panic(err)
	}
	Alls := []*Info{}
	Observers := []*Info{}
	Formulators := []*Info{}
	TxGens := []*Info{}
	ls := strings.Split(string(body), "\n")
	for _, v := range ls {
		ns := strings.Split(v, "\t")
		if len(ns) < 4 {
			break
		}
		info := &Info{
			ServerName: ns[0],
			ServerIP:   ns[1],
			LocalIP:    ns[2],
			Region:     ns[3],
			Type:       ns[4],
			Name:       ns[5],
			Address:    ns[6],
			Gen:        ns[7],
			GenHash:    ns[8],
			Key:        ns[9],
			KeyHash:    ns[10],
		}
		Alls = append(Alls, info)

		switch info.Type {
		case "observer":
			Observers = append(Observers, info)
		case "formulator":
			Formulators = append(Formulators, info)
		case "txgen":
			TxGens = append(TxGens, info)
		}
	}

	rootCmd := &cobra.Command{Use: "cli"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "start testnet",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			// 서버에 연결
			var wg sync.WaitGroup
			for _, v := range Alls {
				wg.Add(1)
				go func(info *Info) {
					defer wg.Done()

					s := NewServer(info.ServerIP, config)
					c, err := s.Open()
					if err != nil {
						panic(err)
					}
					info.Conn = c
				}(v)
			}
			wg.Wait()

			//서비스 시작
			fmt.Print("Start All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StartService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "stop",
		Short: "stop testnet",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			// 서버에 연결
			var wg sync.WaitGroup
			for _, v := range Alls {
				wg.Add(1)
				go func(info *Info) {
					defer wg.Done()

					s := NewServer(info.ServerIP, config)
					c, err := s.Open()
					if err != nil {
						panic(err)
					}
					info.Conn = c
				}(v)
			}
			wg.Wait()

			//서비스 정지
			fmt.Print("Stop All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StopService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "check status of testnet",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			// 서버에 연결
			var wg sync.WaitGroup
			for _, v := range Alls {
				wg.Add(1)
				go func(info *Info) {
					defer wg.Done()

					s := NewServer(info.ServerIP, config)
					c, err := s.Open()
					if err != nil {
						panic(err)
					}
					info.Conn = c
				}(v)
			}
			wg.Wait()

			//옵저버 설치 체크
			fmt.Print("Check Observer Installation ... ")
			if err := ExecuteServer(config, Observers, func(info *Info) error {
				return InstallCheck(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//포뮬레이터 설치 체크
			fmt.Print("Check Formulator Installation ... ")
			if err := ExecuteServer(config, Formulators, func(info *Info) error {
				return InstallCheck(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//Tx 생성기 설치 체크
			fmt.Print("Check TxGen Installation ... ")
			if err := ExecuteServer(config, TxGens, func(info *Info) error {
				return InstallCheck(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//서비스 정지
			fmt.Print("Stop All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StopService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//데이터 클리어
			fmt.Print("Clear Data ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ClearData(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//생성 모드 On
			fmt.Print("Turn On Create Mode ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ChangeMode(info, "create", "")
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//서비스 시작
			fmt.Print("Start All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StartService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			time.Sleep(5 * time.Second)
			//가동 높이 진행 확인하고 100블록 진행
			fmt.Println("Wait until 10 blocks")
			var TryCount uint32
			for {
				Height, err := GetHeight(Formulators[0].ServerIP)
				if err != nil {
					fmt.Println("[Fail] - ", err)
					return
				}
				if Height >= 10 {
					fmt.Println("Reach to 10 blocks [Success]")
					break
				}
				time.Sleep(1 * time.Second)
				TryCount++
				if TryCount%5 == 0 {
					fmt.Println("Fetching Height :", Height)
				}
				if TryCount > 10 {
					fmt.Println("Reach to 10 blocks [Fail] - Cannot reach to objective height within 10 seconds")
					return
				}
			}
			//요약 정보
			fmt.Println("[Print Summary]")
			ret, err := GetSummary(Formulators[0].ServerIP, "10")
			if err != nil {
				fmt.Println("[Fail] - ", err)
			}
			fmt.Println(ret)
			//서비스 정지
			fmt.Print("Stop All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StopService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
		},
	})
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "run test",
	}
	testCmd.AddCommand(&cobra.Command{
		Use:   "1 [block]",
		Short: "create mode test",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			iv, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				panic(err)
			}
			BlockCount := uint32(iv)

			// 서버에 연결
			var wg sync.WaitGroup
			for _, v := range Alls {
				wg.Add(1)
				go func(info *Info) {
					defer wg.Done()

					s := NewServer(info.ServerIP, config)
					c, err := s.Open()
					if err != nil {
						panic(err)
					}
					info.Conn = c
				}(v)
			}
			wg.Wait()

			//서비스 정지
			fmt.Print("Stop All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StopService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//데이터 클리어
			fmt.Print("Clear Data ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ClearData(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//생성 모드 On
			fmt.Print("Turn On Create Mode ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ChangeMode(info, "create", "")
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//서비스 시작
			fmt.Print("Start All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StartService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			time.Sleep(5 * time.Second)
			//가동 높이 진행 확인하고 지정된 블록 진행
			fmt.Println("Wait until " + args[0] + " blocks")
			var TryCount uint32
			for {
				Height, err := GetHeight(Formulators[0].ServerIP)
				if err != nil {
					fmt.Println("[Fail] - ", err)
					return
				}
				if Height >= BlockCount {
					fmt.Println("Reach to " + args[0] + " blocks [Success]")
					break
				}
				time.Sleep(1 * time.Second)
				TryCount++
				if TryCount%5 == 0 {
					fmt.Println("Fetching Height :", Height)
				}
				if TryCount > BlockCount {
					fmt.Println("Reach to " + args[0] + " blocks [Fail] - Cannot reach to objective height within " + args[0] + " seconds")
					return
				}
			}
			//요약 정보
			fmt.Println("[Print Summary]")
			ret, err := GetSummary0Tx(Formulators[0].ServerIP, args[0])
			if err != nil {
				fmt.Println("[Fail] - ", err)
				return
			}
			fmt.Println(ret)
		},
	})
	testCmd.AddCommand(&cobra.Command{
		Use:   "2 [user count] [request per user]",
		Short: "read test",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := strconv.ParseUint(args[0], 10, 32); err != nil {
				panic(err)
			}
			if _, err := strconv.ParseUint(args[1], 10, 64); err != nil {
				panic(err)
			}

			// 서버에 연결
			var wg sync.WaitGroup
			for _, v := range Alls {
				wg.Add(1)
				go func(info *Info) {
					defer wg.Done()

					s := NewServer(info.ServerIP, config)
					c, err := s.Open()
					if err != nil {
						panic(err)
					}
					info.Conn = c
				}(v)
			}
			wg.Wait()

			//서비스 정지
			fmt.Print("Stop All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StopService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//데이터 클리어
			fmt.Print("Clear Data ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ClearData(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//생성 모드 On
			fmt.Print("Turn On Create Mode ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ChangeMode(info, "create", "")
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//서비스 시작
			fmt.Print("Start All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StartService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			time.Sleep(5 * time.Second)
			//가동 높이 진행 확인하고 지정된 블록 진행
			BlockCount := uint32(10)
			BlockCountStr := strconv.FormatUint(uint64(BlockCount), 10)
			fmt.Println("Wait until " + BlockCountStr + " blocks")
			var TryCount uint32
			for {
				Height, err := GetHeight(Formulators[0].ServerIP)
				if err != nil {
					fmt.Println("[Fail] - ", err)
					return
				}
				if Height >= BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Success]")
					break
				}
				time.Sleep(1 * time.Second)
				TryCount++
				if TryCount%5 == 0 {
					fmt.Println("Fetching Height :", Height)
				}
				if TryCount > BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Fail] - Cannot reach to objective height within " + BlockCountStr + " seconds")
					return
				}
			}
			//서비스 정지
			fmt.Print("Stop All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StopService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//일반 모드 On
			fmt.Print("Turn On Normal Mode ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ChangeMode(info, "normal", "")
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//서비스 시작
			fmt.Print("Start All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StartService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			time.Sleep(5 * time.Second)
			//읽기 테스트
			fmt.Println("[Start Read Test]")
			ret, err := GetReadTest(Formulators[0].ServerIP, args[0], args[1])
			if err != nil {
				fmt.Println("[Fail] - ", err)
				return
			}
			fmt.Println(ret)
		},
	})
	testCmd.AddCommand(&cobra.Command{
		Use:   "2a [user count] [request per user]",
		Short: "read test",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := strconv.ParseUint(args[0], 10, 32); err != nil {
				panic(err)
			}
			if _, err := strconv.ParseUint(args[1], 10, 64); err != nil {
				panic(err)
			}

			// 서버에 연결
			var wg sync.WaitGroup
			for _, v := range Alls {
				wg.Add(1)
				go func(info *Info) {
					defer wg.Done()

					s := NewServer(info.ServerIP, config)
					c, err := s.Open()
					if err != nil {
						panic(err)
					}
					info.Conn = c
				}(v)
			}
			wg.Wait()

			//서비스 정지
			fmt.Print("Stop All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StopService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//데이터 클리어
			fmt.Print("Clear Data ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ClearData(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//생성 모드 On
			fmt.Print("Turn On Create Mode ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ChangeMode(info, "create", "")
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//서비스 시작
			fmt.Print("Start All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StartService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			time.Sleep(5 * time.Second)
			//가동 높이 진행 확인하고 지정된 블록 진행
			BlockCount := uint32(10)
			BlockCountStr := strconv.FormatUint(uint64(BlockCount), 10)
			fmt.Println("Wait until " + BlockCountStr + " blocks")
			var TryCount uint32
			for {
				Height, err := GetHeight(Formulators[0].ServerIP)
				if err != nil {
					fmt.Println("[Fail] - ", err)
					return
				}
				if Height >= BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Success]")
					break
				}
				time.Sleep(1 * time.Second)
				TryCount++
				if TryCount%5 == 0 {
					fmt.Println("Fetching Height :", Height)
				}
				if TryCount > BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Fail] - Cannot reach to objective height within " + BlockCountStr + " seconds")
					return
				}
			}
			//서비스 정지
			fmt.Print("Stop All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StopService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//일반 모드 On
			fmt.Print("Turn On Normal Mode ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ChangeMode(info, "normal", "")
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//서비스 시작
			fmt.Print("Start All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StartService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			time.Sleep(5 * time.Second)
			//읽기 테스트
			fmt.Println("[Start Read Test (From Local)]")
			userCount, _ := strconv.Atoi(args[0])
			requestPerUser, _ := strconv.Atoi(args[1])

			var SuccessCount uint64
			var ErrorCount uint64
			start := time.Now()
			var wg2 sync.WaitGroup
			for i := 0; i < userCount; i++ {
				wg2.Add(1)
				go func() {
					defer wg2.Done()

					c, _, _ := websocket.DefaultDialer.Dial("ws://45.77.147.144:48000/api/endpoints/websocket", nil)
					defer c.Close()

					for q := 0; q < requestPerUser; q++ {
						if _, err := GetBalance(c, "5CyLcFhpyN"); err != nil {
							log.Println(err)
							atomic.AddUint64(&ErrorCount, 1)
						} else {
							atomic.AddUint64(&SuccessCount, 1)
						}
					}
				}()
			}
			wg2.Wait()

			TimeElapsed := time.Now().Sub(start)
			bs, err := json.Marshal(struct {
				SuccessCount uint64
				ErrorCount   uint64
				TimeElapsed  float64
				TPS          float64
			}{
				SuccessCount: SuccessCount,
				ErrorCount:   ErrorCount,
				TimeElapsed:  float64(TimeElapsed) / float64(time.Second),
				TPS:          float64(SuccessCount+ErrorCount) * float64(time.Second) / float64(TimeElapsed),
			})
			if err != nil {
				fmt.Println("[Fail] - ", err)
				return
			}
			fmt.Println(string(bs))
		},
	})
	testCmd.AddCommand(&cobra.Command{
		Use:   "3 [block] [tx per block]",
		Short: "insert mode test",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			iv, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				panic(err)
			}
			BlockCount := uint32(iv)
			if _, err := strconv.ParseUint(args[1], 10, 32); err != nil {
				panic(err)
			}

			// 서버에 연결
			var wg sync.WaitGroup
			for _, v := range Alls {
				wg.Add(1)
				go func(info *Info) {
					defer wg.Done()

					s := NewServer(info.ServerIP, config)
					c, err := s.Open()
					if err != nil {
						panic(err)
					}
					info.Conn = c
				}(v)
			}
			wg.Wait()

			//서비스 정지
			fmt.Print("Stop All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StopService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//데이터 클리어
			fmt.Print("Clear Data ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ClearData(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//생성 모드 On
			fmt.Print("Turn On Insert Mode ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ChangeMode(info, "insert", args[1])
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//서비스 시작
			fmt.Print("Start All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StartService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			time.Sleep(5 * time.Second)
			//가동 높이 진행 확인하고 지정된 블록 진행
			BlockCountStr := args[0]
			fmt.Println("Wait until " + BlockCountStr + " blocks")
			var TryCount uint32
			for {
				Height, err := GetHeight(Formulators[0].ServerIP)
				if err != nil {
					fmt.Println("[Fail] - ", err)
					return
				}
				if Height >= BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Success]")
					break
				}
				time.Sleep(1 * time.Second)
				TryCount++
				if TryCount%5 == 0 {
					fmt.Println("Fetching Height :", Height)
				}
				if TryCount > BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Fail] - Cannot reach to objective height within " + BlockCountStr + " seconds")
					return
				}
			}
			//요약 정보
			fmt.Println("[Print Summary]")
			ret, err := GetSummary(Formulators[0].ServerIP, args[0])
			if err != nil {
				fmt.Println("[Fail] - ", err)
				return
			}
			fmt.Println(ret)
		},
	})
	testCmd.AddCommand(&cobra.Command{
		Use:   "4 [custom text]",
		Short: "custom text test",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// 서버에 연결
			var wg sync.WaitGroup
			for _, v := range Alls {
				wg.Add(1)
				go func(info *Info) {
					defer wg.Done()

					s := NewServer(info.ServerIP, config)
					c, err := s.Open()
					if err != nil {
						panic(err)
					}
					info.Conn = c
				}(v)
			}
			wg.Wait()

			//서비스 정지
			fmt.Print("Stop All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StopService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//데이터 클리어
			fmt.Print("Clear Data ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ClearData(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//일반 모드 On
			fmt.Print("Turn On Normal Mode ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ChangeMode(info, "normal", "")
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//서비스 시작
			fmt.Print("Start All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StartService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			time.Sleep(5 * time.Second)
			//가동 높이 진행 확인하고 지정된 블록 진행
			BlockCount := uint32(10)
			BlockCountStr := strconv.FormatUint(uint64(BlockCount), 10)
			fmt.Println("Wait until " + BlockCountStr + " blocks")
			var TryCount uint32
			for {
				Height, err := GetHeight(Formulators[0].ServerIP)
				if err != nil {
					fmt.Println("[Fail] - ", err)
					return
				}
				if Height >= BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Success]")
					break
				}
				time.Sleep(1 * time.Second)
				TryCount++
				if TryCount%5 == 0 {
					fmt.Println("Fetching Height :", Height)
				}
				if TryCount > BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Fail] - Cannot reach to objective height within " + BlockCountStr + " seconds")
					return
				}
			}
			//임의의 문자 Tx 발송
			fmt.Print("Send Custom Data ... ")
			id, err := SendCustomText(Formulators[0].ServerIP, args[0])
			if err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			time.Sleep(5 * time.Second)
			fmt.Print("Get Custom Data ... ")
			ret, err := GetCustomText(Formulators[0].ServerIP, id[1:len(id)-1])
			if err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println(ret)
			}
			//무결성 체크
			fmt.Println("[Print Check Integrity]")
			ret2, err := GetCheckIntegrity(Formulators[0].ServerIP, "20")
			if err != nil {
				fmt.Println("[Fail] - ", err)
				return
			}
			fmt.Println(ret2)
		},
	})
	testCmd.AddCommand(&cobra.Command{
		Use:   "5",
		Short: "disaster recovery test",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			// 서버에 연결
			var wg sync.WaitGroup
			for _, v := range Alls {
				wg.Add(1)
				go func(info *Info) {
					defer wg.Done()

					s := NewServer(info.ServerIP, config)
					c, err := s.Open()
					if err != nil {
						panic(err)
					}
					info.Conn = c
				}(v)
			}
			wg.Wait()

			//서비스 정지
			fmt.Print("Stop All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StopService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//데이터 클리어
			fmt.Print("Clear Data ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ClearData(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//일반 모드 On
			fmt.Print("Turn On Normal Mode ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ChangeMode(info, "normal", "")
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//서비스 시작
			fmt.Print("Start All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StartService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			time.Sleep(5 * time.Second)
			//가동 높이 진행 확인하고 지정된 블록 진행
			BlockCount := uint32(10)
			BlockCountStr := strconv.FormatUint(uint64(BlockCount), 10)
			fmt.Println("Wait until " + BlockCountStr + " blocks")
			var TryCount uint32
			for {
				Height, err := GetHeight(Formulators[0].ServerIP)
				if err != nil {
					fmt.Println("[Fail] - ", err)
					return
				}
				if Height >= BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Success]")
					break
				}
				time.Sleep(1 * time.Second)
				TryCount++
				if TryCount%5 == 0 {
					fmt.Println("Fetching Height :", Height)
				}
				if TryCount > BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Fail] - Cannot reach to objective height within " + BlockCountStr + " seconds")
					return
				}
			}
			fmt.Print("Stop All Formulators Except Formulator[0] ... ")
			if err := ExecuteServer(config, Formulators[1:], func(info *Info) error {
				return StopService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			fmt.Print("Stop Observer[0] ... ")
			if err := ExecuteServer(config, Observers[:1], func(info *Info) error {
				return StopService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//가동 높이 진행 확인하고 지정된 블록 진행
			BlockCount = uint32(50)
			BlockCountStr = "50"
			fmt.Println("Wait until " + BlockCountStr + " blocks")
			TryCount = 0
			for {
				Height, err := GetHeight(Formulators[0].ServerIP)
				if err != nil {
					fmt.Println("[Fail] - ", err)
					return
				}
				if Height >= BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Success]")
					break
				}
				time.Sleep(1 * time.Second)
				TryCount++
				if TryCount%5 == 0 {
					fmt.Println("Fetching Height :", Height)
				}
				if TryCount > BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Fail] - Cannot reach to objective height within " + BlockCountStr + " seconds")
					return
				}
			}
		},
	})
	testCmd.AddCommand(&cobra.Command{
		Use:   "5a",
		Short: "disaster recovery test",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			// 서버에 연결
			var wg sync.WaitGroup
			for _, v := range Alls {
				wg.Add(1)
				go func(info *Info) {
					defer wg.Done()

					s := NewServer(info.ServerIP, config)
					c, err := s.Open()
					if err != nil {
						panic(err)
					}
					info.Conn = c
				}(v)
			}
			wg.Wait()

			//서비스 정지
			fmt.Print("Stop All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StopService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//데이터 클리어
			fmt.Print("Clear Data ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ClearData(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//일반 모드 On
			fmt.Print("Turn On Normal Mode ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ChangeMode(info, "normal", "")
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//서비스 시작
			fmt.Print("Start All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StartService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			time.Sleep(5 * time.Second)
			//가동 높이 진행 확인하고 지정된 블록 진행
			BlockCount := uint32(10)
			BlockCountStr := strconv.FormatUint(uint64(BlockCount), 10)
			fmt.Println("Wait until " + BlockCountStr + " blocks")
			var TryCount uint32
			for {
				Height, err := GetHeight(Formulators[0].ServerIP)
				if err != nil {
					fmt.Println("[Fail] - ", err)
					return
				}
				if Height >= BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Success]")
					break
				}
				time.Sleep(1 * time.Second)
				TryCount++
				if TryCount%5 == 0 {
					fmt.Println("Fetching Height :", Height)
				}
				if TryCount > BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Fail] - Cannot reach to objective height within " + BlockCountStr + " seconds")
					return
				}
			}
			fmt.Print("Stop All Formulators Except Formulator[0] ... ")
			if err := ExecuteServer(config, Formulators[1:], func(info *Info) error {
				return StopService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//가동 높이 진행 확인하고 지정된 블록 진행
			BlockCount = uint32(50)
			BlockCountStr = "50"
			fmt.Println("Wait until " + BlockCountStr + " blocks")
			TryCount = 0
			for {
				Height, err := GetHeight(Formulators[0].ServerIP)
				if err != nil {
					fmt.Println("[Fail] - ", err)
					return
				}
				if Height >= BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Success]")
					break
				}
				time.Sleep(1 * time.Second)
				TryCount++
				if TryCount%5 == 0 {
					fmt.Println("Fetching Height :", Height)
				}
				if TryCount > BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Fail] - Cannot reach to objective height within " + BlockCountStr + " seconds")
					return
				}
			}
		},
	})
	testCmd.AddCommand(&cobra.Command{
		Use:   "5b",
		Short: "disaster recovery test",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			// 서버에 연결
			var wg sync.WaitGroup
			for _, v := range Alls {
				wg.Add(1)
				go func(info *Info) {
					defer wg.Done()

					s := NewServer(info.ServerIP, config)
					c, err := s.Open()
					if err != nil {
						panic(err)
					}
					info.Conn = c
				}(v)
			}
			wg.Wait()

			//서비스 정지
			fmt.Print("Stop All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StopService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//데이터 클리어
			fmt.Print("Clear Data ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ClearData(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//일반 모드 On
			fmt.Print("Turn On Normal Mode ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ChangeMode(info, "normal", "")
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//서비스 시작
			fmt.Print("Start All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StartService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			time.Sleep(5 * time.Second)
			//가동 높이 진행 확인하고 지정된 블록 진행
			BlockCount := uint32(10)
			BlockCountStr := strconv.FormatUint(uint64(BlockCount), 10)
			fmt.Println("Wait until " + BlockCountStr + " blocks")
			var TryCount uint32
			for {
				Height, err := GetHeight(Formulators[0].ServerIP)
				if err != nil {
					fmt.Println("[Fail] - ", err)
					return
				}
				if Height >= BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Success]")
					break
				}
				time.Sleep(1 * time.Second)
				TryCount++
				if TryCount%5 == 0 {
					fmt.Println("Fetching Height :", Height)
				}
				if TryCount > BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Fail] - Cannot reach to objective height within " + BlockCountStr + " seconds")
					return
				}
			}
			fmt.Print("Stop Observer[0] ... ")
			if err := ExecuteServer(config, Observers[:1], func(info *Info) error {
				return StopService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//가동 높이 진행 확인하고 지정된 블록 진행
			BlockCount = uint32(50)
			BlockCountStr = "50"
			fmt.Println("Wait until " + BlockCountStr + " blocks")
			TryCount = 0
			for {
				Height, err := GetHeight(Formulators[0].ServerIP)
				if err != nil {
					fmt.Println("[Fail] - ", err)
					return
				}
				if Height >= BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Success]")
					break
				}
				time.Sleep(1 * time.Second)
				TryCount++
				if TryCount%5 == 0 {
					fmt.Println("Fetching Height :", Height)
				}
				if TryCount > BlockCount {
					fmt.Println("Reach to " + BlockCountStr + " blocks [Fail] - Cannot reach to objective height within " + BlockCountStr + " seconds")
					return
				}
			}
		},
	})
	testCmd.AddCommand(&cobra.Command{
		Use:   "6 [height]",
		Short: "check integrity test",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			iv, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				panic(err)
			}
			BlockCount := uint32(iv)

			// 서버에 연결
			var wg sync.WaitGroup
			for _, v := range Alls {
				wg.Add(1)
				go func(info *Info) {
					defer wg.Done()

					s := NewServer(info.ServerIP, config)
					c, err := s.Open()
					if err != nil {
						panic(err)
					}
					info.Conn = c
				}(v)
			}
			wg.Wait()

			//서비스 정지
			fmt.Print("Stop All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StopService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//데이터 클리어
			fmt.Print("Clear Data ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ClearData(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//생성 모드 On
			fmt.Print("Turn On Create Mode ... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return ChangeMode(info, "create", "")
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			//서비스 시작
			fmt.Print("Start All Services... ")
			if err := ExecuteServer(config, Alls, func(info *Info) error {
				return StartService(info)
			}); err != nil {
				fmt.Println("[Fail] - ", err)
				return
			} else {
				fmt.Println("[Success]")
			}
			time.Sleep(5 * time.Second)
			//가동 높이 진행 확인하고 지정된 블록 진행
			fmt.Println("Wait until " + args[0] + " blocks")
			var TryCount uint32
			for {
				Height, err := GetHeight(Formulators[0].ServerIP)
				if err != nil {
					fmt.Println("[Fail] - ", err)
					return
				}
				if Height >= BlockCount {
					fmt.Println("Reach to " + args[0] + " blocks [Success]")
					break
				}
				time.Sleep(1 * time.Second)
				TryCount++
				if TryCount%5 == 0 {
					fmt.Println("Fetching Height :", Height)
				}
				if TryCount > BlockCount {
					fmt.Println("Reach to " + args[0] + " blocks [Fail] - Cannot reach to objective height within " + args[0] + " seconds")
					return
				}
			}
			//무결성 체크
			fmt.Println("[Print Check Integrity]")
			ret, err := GetCheckIntegrity(Formulators[0].ServerIP, args[0])
			if err != nil {
				fmt.Println("[Fail] - ", err)
				return
			}
			fmt.Println(ret)
		},
	})
	rootCmd.AddCommand(testCmd)
	rootCmd.Execute()
}

func ExecuteServer(config *ssh.ClientConfig, infos []*Info, fn func(info *Info) error) error {
	var inErr error
	var wg sync.WaitGroup
	for _, v := range infos {
		wg.Add(1)
		go func(info *Info) {
			defer wg.Done()
			if err := fn(info); err != nil {
				inErr = err
			}
		}(v)
	}
	wg.Wait()
	return inErr
}

func InstallCheck(info *Info) error {
	bs, err := info.Conn.Output("ls /root/fleta_testnet/" + info.Type)
	if err != nil {
		return err
	}
	ret := strings.TrimSpace(string(bs))
	if ret == "/root/fleta_testnet/"+info.Type {
		return nil
	} else {
		return ErrNotExist
	}
}

func ClearData(info *Info) error {
	cmds := []string{
		"rm -Rf /root/fleta_testnet/ndata",
		"rm -Rf /root/fleta_testnet/odata",
		"rm -Rf /root/fleta_testnet/fdata",
	}
	for _, cmd := range cmds {
		if _, err := info.Conn.Output(cmd); err != nil {
			return err
		}
	}
	return nil
}

func StopService(info *Info) error {
	if _, err := info.Conn.Output("systemctl stop " + info.Type); err != nil {
		return err
	}
	return nil
}

func StartService(info *Info) error {
	cmds := []string{
		"systemctl daemon-reload",
		"systemctl start " + info.Type,
	}
	for _, cmd := range cmds {
		bs, err := info.Conn.Output(cmd)
		ret := strings.TrimSpace(string(bs))
		if err != nil {
			return err
		}
		if len(ret) > 0 {
			return errors.New(ret)
		}
	}
	return nil
}

func ChangeMode(info *Info, mode string, InsertTxCount string) error {
	switch mode {
	case "create":
		cmds := []string{
			"rm -Rf /root/fleta_testnet/config.toml",
		}
		switch info.Type {
		case "observer":
			cmd := `cat > /root/fleta_testnet/config.toml << 'EOF'
KeyHex = "` + info.Gen + `"
ObseverPort = 45000
FormulatorPort = 47000
APIPort = 48000
StoreRoot = "./odata"
BackendVersion = 1
RLogHost = ""
RLogPath = ""
UseRLog = false

[ObserverKeyMap]
4JDtZL53jhs7akrTjeaJicnA1ub99vUKkXeySUy6uVZ = "10.1.96.3:45000"
4f52SK2FEc6XzNuQfdQbLmV6o9Dg6UwD5Ajf8NM8XxR = "10.1.96.4:45000"
37mZ3Gt3yW1TU3tt9zPF9hstUHedoXLsfXi8RTPp8Ze = "10.1.96.5:45000"
4c3FinyoBt1BNwv17tHc5gQKVSvu785rM7zq1R58hhL = "10.1.96.6:45000"
3BeyVF3kiCgYZRdPwC5D2C5xddrzmhB8kSaPzjSi59S = "10.1.96.7:45000"
EOF`
			cmds = append(cmds, cmd)
		case "formulator":
			cmd := `cat > /root/fleta_testnet/config.toml << 'EOF'
Port = 41000
APIPort = 48000
GenKeyHex = "` + info.Gen + `"
Formulator = "` + info.Address + `"
StoreRoot = "./fdata"
BackendVersion = 1
RLogHost = ""
RLogPath = ""
UseRLog = false

[ObserverKeyMap]
4JDtZL53jhs7akrTjeaJicnA1ub99vUKkXeySUy6uVZ = "observer0.fletatest.net:47000"
4f52SK2FEc6XzNuQfdQbLmV6o9Dg6UwD5Ajf8NM8XxR = "observer1.fletatest.net:47000"
37mZ3Gt3yW1TU3tt9zPF9hstUHedoXLsfXi8RTPp8Ze = "observer2.fletatest.net:47000"
4c3FinyoBt1BNwv17tHc5gQKVSvu785rM7zq1R58hhL = "observer3.fletatest.net:47000"
3BeyVF3kiCgYZRdPwC5D2C5xddrzmhB8kSaPzjSi59S = "observer4.fletatest.net:47000"

[SeedNodeMap]
3yTFnJJqx3wCiK2Edk9f9JwdvdkC4DP4T1y8xYztMkf = "seednode1.fletatest.net:41000"
3EjA1hKkfYZ4KL1c4f67CfaNwb9fCqUneiYkyQEhsGi = "seednode2.fletatest.net:41000"
314AUADxjj7nWjeNpR8XEoAh4DdX3ArNHaipPGMFQ4u = "seednode3.fletatest.net:41000"
3n8QNWd7M839ouauhdHvmgmk4NsLj4qGM6tpfoaLNxc = "seednode4.fletatest.net:41000"
EOF`
			cmds = append(cmds, cmd)
		case "txgen":
			cmd := `cat > /root/fleta_testnet/config.toml << 'EOF'
ObserverKeys = [
"4JDtZL53jhs7akrTjeaJicnA1ub99vUKkXeySUy6uVZ",
"4f52SK2FEc6XzNuQfdQbLmV6o9Dg6UwD5Ajf8NM8XxR",
"37mZ3Gt3yW1TU3tt9zPF9hstUHedoXLsfXi8RTPp8Ze",
"4c3FinyoBt1BNwv17tHc5gQKVSvu785rM7zq1R58hhL",
"3BeyVF3kiCgYZRdPwC5D2C5xddrzmhB8kSaPzjSi59S",
]
Port = 41000
APIPort = 48000
WebPort = 8080
NodeKeyHex = "` + info.Gen + `"
StoreRoot = "./ndata"
BackendVersion = 1
RLogHost = ""
RLogPath = ""
UseRLog = false
CreateMode = true

[SeedNodeMap]
3yTFnJJqx3wCiK2Edk9f9JwdvdkC4DP4T1y8xYztMkf = "seednode1.fletatest.net:41000"
3EjA1hKkfYZ4KL1c4f67CfaNwb9fCqUneiYkyQEhsGi = "seednode2.fletatest.net:41000"
314AUADxjj7nWjeNpR8XEoAh4DdX3ArNHaipPGMFQ4u = "seednode3.fletatest.net:41000"
3n8QNWd7M839ouauhdHvmgmk4NsLj4qGM6tpfoaLNxc = "seednode4.fletatest.net:41000"
EOF`
			cmds = append(cmds, cmd)
		}
		for _, cmd := range cmds {
			if _, err := info.Conn.Output(cmd); err != nil {
				return err
			}
		}
		return nil
	case "insert":
		cmds := []string{
			"rm -Rf /root/fleta_testnet/config.toml",
		}
		switch info.Type {
		case "observer":
			cmd := `cat > /root/fleta_testnet/config.toml << 'EOF'
KeyHex = "` + info.Gen + `"
ObseverPort = 45000
FormulatorPort = 47000
APIPort = 48000
StoreRoot = "./odata"
BackendVersion = 1
RLogHost = ""
RLogPath = ""
UseRLog = false

[ObserverKeyMap]
4JDtZL53jhs7akrTjeaJicnA1ub99vUKkXeySUy6uVZ = "10.1.96.3:45000"
4f52SK2FEc6XzNuQfdQbLmV6o9Dg6UwD5Ajf8NM8XxR = "10.1.96.4:45000"
37mZ3Gt3yW1TU3tt9zPF9hstUHedoXLsfXi8RTPp8Ze = "10.1.96.5:45000"
4c3FinyoBt1BNwv17tHc5gQKVSvu785rM7zq1R58hhL = "10.1.96.6:45000"
3BeyVF3kiCgYZRdPwC5D2C5xddrzmhB8kSaPzjSi59S = "10.1.96.7:45000"
EOF`
			cmds = append(cmds, cmd)
		case "formulator":
			cmd := `cat > /root/fleta_testnet/config.toml << 'EOF'
Port = 41000
APIPort = 48000
GenKeyHex = "` + info.Gen + `"
Formulator = "` + info.Address + `"
StoreRoot = "./fdata"
BackendVersion = 1
RLogHost = ""
RLogPath = ""
UseRLog = false
InsertMode = true
InsertTxCount = ` + InsertTxCount + `

[ObserverKeyMap]
4JDtZL53jhs7akrTjeaJicnA1ub99vUKkXeySUy6uVZ = "observer0.fletatest.net:47000"
4f52SK2FEc6XzNuQfdQbLmV6o9Dg6UwD5Ajf8NM8XxR = "observer1.fletatest.net:47000"
37mZ3Gt3yW1TU3tt9zPF9hstUHedoXLsfXi8RTPp8Ze = "observer2.fletatest.net:47000"
4c3FinyoBt1BNwv17tHc5gQKVSvu785rM7zq1R58hhL = "observer3.fletatest.net:47000"
3BeyVF3kiCgYZRdPwC5D2C5xddrzmhB8kSaPzjSi59S = "observer4.fletatest.net:47000"

[SeedNodeMap]
3yTFnJJqx3wCiK2Edk9f9JwdvdkC4DP4T1y8xYztMkf = "seednode1.fletatest.net:41000"
3EjA1hKkfYZ4KL1c4f67CfaNwb9fCqUneiYkyQEhsGi = "seednode2.fletatest.net:41000"
314AUADxjj7nWjeNpR8XEoAh4DdX3ArNHaipPGMFQ4u = "seednode3.fletatest.net:41000"
3n8QNWd7M839ouauhdHvmgmk4NsLj4qGM6tpfoaLNxc = "seednode4.fletatest.net:41000"
EOF`
			cmds = append(cmds, cmd)
		case "txgen":
			cmd := `cat > /root/fleta_testnet/config.toml << 'EOF'
ObserverKeys = [
"4JDtZL53jhs7akrTjeaJicnA1ub99vUKkXeySUy6uVZ",
"4f52SK2FEc6XzNuQfdQbLmV6o9Dg6UwD5Ajf8NM8XxR",
"37mZ3Gt3yW1TU3tt9zPF9hstUHedoXLsfXi8RTPp8Ze",
"4c3FinyoBt1BNwv17tHc5gQKVSvu785rM7zq1R58hhL",
"3BeyVF3kiCgYZRdPwC5D2C5xddrzmhB8kSaPzjSi59S",
]
Port = 41000
APIPort = 48000
WebPort = 8080
NodeKeyHex = "` + info.Gen + `"
StoreRoot = "./ndata"
BackendVersion = 1
RLogHost = ""
RLogPath = ""
UseRLog = false

[SeedNodeMap]
3yTFnJJqx3wCiK2Edk9f9JwdvdkC4DP4T1y8xYztMkf = "seednode1.fletatest.net:41000"
3EjA1hKkfYZ4KL1c4f67CfaNwb9fCqUneiYkyQEhsGi = "seednode2.fletatest.net:41000"
314AUADxjj7nWjeNpR8XEoAh4DdX3ArNHaipPGMFQ4u = "seednode3.fletatest.net:41000"
3n8QNWd7M839ouauhdHvmgmk4NsLj4qGM6tpfoaLNxc = "seednode4.fletatest.net:41000"
EOF`
			cmds = append(cmds, cmd)
		}
		for _, cmd := range cmds {
			if _, err := info.Conn.Output(cmd); err != nil {
				return err
			}
		}
		return nil
	default:
		cmds := []string{
			"rm -Rf /root/fleta_testnet/config.toml",
		}
		switch info.Type {
		case "observer":
			cmd := `cat > /root/fleta_testnet/config.toml << 'EOF'
KeyHex = "` + info.Gen + `"
ObseverPort = 45000
FormulatorPort = 47000
APIPort = 48000
StoreRoot = "./odata"
BackendVersion = 1
RLogHost = ""
RLogPath = ""
UseRLog = false

[ObserverKeyMap]
4JDtZL53jhs7akrTjeaJicnA1ub99vUKkXeySUy6uVZ = "10.1.96.3:45000"
4f52SK2FEc6XzNuQfdQbLmV6o9Dg6UwD5Ajf8NM8XxR = "10.1.96.4:45000"
37mZ3Gt3yW1TU3tt9zPF9hstUHedoXLsfXi8RTPp8Ze = "10.1.96.5:45000"
4c3FinyoBt1BNwv17tHc5gQKVSvu785rM7zq1R58hhL = "10.1.96.6:45000"
3BeyVF3kiCgYZRdPwC5D2C5xddrzmhB8kSaPzjSi59S = "10.1.96.7:45000"
EOF`
			cmds = append(cmds, cmd)
		case "formulator":
			cmd := `cat > /root/fleta_testnet/config.toml << 'EOF'
Port = 41000
APIPort = 48000
GenKeyHex = "` + info.Gen + `"
Formulator = "` + info.Address + `"
StoreRoot = "./fdata"
BackendVersion = 1
RLogHost = ""
RLogPath = ""
UseRLog = false

[ObserverKeyMap]
4JDtZL53jhs7akrTjeaJicnA1ub99vUKkXeySUy6uVZ = "observer0.fletatest.net:47000"
4f52SK2FEc6XzNuQfdQbLmV6o9Dg6UwD5Ajf8NM8XxR = "observer1.fletatest.net:47000"
37mZ3Gt3yW1TU3tt9zPF9hstUHedoXLsfXi8RTPp8Ze = "observer2.fletatest.net:47000"
4c3FinyoBt1BNwv17tHc5gQKVSvu785rM7zq1R58hhL = "observer3.fletatest.net:47000"
3BeyVF3kiCgYZRdPwC5D2C5xddrzmhB8kSaPzjSi59S = "observer4.fletatest.net:47000"

[SeedNodeMap]
3yTFnJJqx3wCiK2Edk9f9JwdvdkC4DP4T1y8xYztMkf = "seednode1.fletatest.net:41000"
3EjA1hKkfYZ4KL1c4f67CfaNwb9fCqUneiYkyQEhsGi = "seednode2.fletatest.net:41000"
314AUADxjj7nWjeNpR8XEoAh4DdX3ArNHaipPGMFQ4u = "seednode3.fletatest.net:41000"
3n8QNWd7M839ouauhdHvmgmk4NsLj4qGM6tpfoaLNxc = "seednode4.fletatest.net:41000"
EOF`
			cmds = append(cmds, cmd)
		case "txgen":
			cmd := `cat > /root/fleta_testnet/config.toml << 'EOF'
ObserverKeys = [
"4JDtZL53jhs7akrTjeaJicnA1ub99vUKkXeySUy6uVZ",
"4f52SK2FEc6XzNuQfdQbLmV6o9Dg6UwD5Ajf8NM8XxR",
"37mZ3Gt3yW1TU3tt9zPF9hstUHedoXLsfXi8RTPp8Ze",
"4c3FinyoBt1BNwv17tHc5gQKVSvu785rM7zq1R58hhL",
"3BeyVF3kiCgYZRdPwC5D2C5xddrzmhB8kSaPzjSi59S",
]
Port = 41000
APIPort = 48000
WebPort = 8080
NodeKeyHex = "` + info.Gen + `"
StoreRoot = "./ndata"
BackendVersion = 1
RLogHost = ""
RLogPath = ""
UseRLog = false

[SeedNodeMap]
3yTFnJJqx3wCiK2Edk9f9JwdvdkC4DP4T1y8xYztMkf = "seednode1.fletatest.net:41000"
3EjA1hKkfYZ4KL1c4f67CfaNwb9fCqUneiYkyQEhsGi = "seednode2.fletatest.net:41000"
314AUADxjj7nWjeNpR8XEoAh4DdX3ArNHaipPGMFQ4u = "seednode3.fletatest.net:41000"
3n8QNWd7M839ouauhdHvmgmk4NsLj4qGM6tpfoaLNxc = "seednode4.fletatest.net:41000"
EOF`
			cmds = append(cmds, cmd)
		}
		for _, cmd := range cmds {
			if _, err := info.Conn.Output(cmd); err != nil {
				return err
			}
		}
		return nil
	}
}

func GetHeight(IP string) (uint32, error) {
	res, err := DoRequest("http://"+IP+":48000", "chain.height", []interface{}{})
	if err != nil {
		return 0, err
	} else {
		bs, err := json.MarshalIndent(res, "", "\t")
		if err != nil {
			return 0, err
		} else {
			iv, err := strconv.ParseUint(string(bs), 10, 32)
			if err != nil {
				return 0, err
			}
			return uint32(iv), nil
		}
	}
}

func GetSummary(IP string, Height string) (string, error) {
	res, err := DoRequest("http://"+IP+":48000", "chain.summary", []interface{}{Height})
	if err != nil {
		return "", err
	} else {
		bs, err := json.MarshalIndent(res, "", "\t")
		if err != nil {
			return "", err
		} else {
			return string(bs), nil
		}
	}
}

func GetSummary0Tx(IP string, Height string) (string, error) {
	res, err := DoRequest("http://"+IP+":48000", "chain.summary0tx", []interface{}{Height})
	if err != nil {
		return "", err
	} else {
		bs, err := json.MarshalIndent(res, "", "\t")
		if err != nil {
			return "", err
		} else {
			return string(bs), nil
		}
	}
}

func SendCustomText(IP string, Text string) (string, error) {
	res, err := DoRequest("http://"+IP+":48000", "chain.sendText", []interface{}{Text})
	if err != nil {
		return "", err
	} else {
		bs, err := json.MarshalIndent(res, "", "\t")
		if err != nil {
			return "", err
		} else {
			return string(bs), nil
		}
	}
}

func GetCustomText(IP string, ID string) (string, error) {
	res, err := DoRequest("http://"+IP+":48000", "chain.getText", []interface{}{ID})
	if err != nil {
		return "", err
	} else {
		bs, err := json.MarshalIndent(res, "", "\t")
		if err != nil {
			return "", err
		} else {
			return string(bs), nil
		}
	}
}

func GetCheckIntegrity(IP string, Height string) (string, error) {
	res, err := DoRequest("http://"+IP+":48000", "chain.checkIntegrity", []interface{}{Height})
	if err != nil {
		return "", err
	} else {
		bs, err := json.MarshalIndent(res, "", "\t")
		if err != nil {
			return "", err
		} else {
			return string(bs), nil
		}
	}
}

func GetReadTest(IP string, UserCount string, RequestPerSecond string) (string, error) {
	res, err := DoRequest("http://"+IP+":48000", "chain.readTest", []interface{}{UserCount, RequestPerSecond})
	if err != nil {
		return "", err
	} else {
		bs, err := json.MarshalIndent(res, "", "\t")
		if err != nil {
			return "", err
		} else {
			return string(bs), nil
		}
	}
}
