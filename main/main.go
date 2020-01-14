package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
	"github.com/garyburd/redigo/redis"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	const healthCheckPeriod = time.Minute
	conn, err := redis.Dial("tcp", "127.0.0.1:6379")

	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	defer conn.Close()
	psc := redis.PubSubConn{conn}
	if err := psc.Subscribe(redis.Args{}.AddFlat("cmd1")...); err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("成功连接")
	done := make(chan error, 1)
	go func() {
		for {
			switch n := psc.Receive().(type) {
			case error:
				fmt.Println("出错")
				done <- n
				return
			case redis.Message:
				fmt.Printf("m1:%v,%v", n.Channel, n.Data)
				if string(n.Data) == "lock" {
					cmd := exec.Command("LockWorkStation.bat")
					err := cmd.Run()
					if err != nil {
						fmt.Println("Execute Command failed:" + err.Error())
						return
					}
				}
				if string(n.Data) == "shutdown" {
					cmd := exec.Command("shutdown.bat")
					err := cmd.Run()
					if err != nil {
						fmt.Println("Execute Command failed:" + err.Error())
						return
					}
				}
			case redis.Subscription:
				fmt.Println("监听")
				switch n.Count {
				case 1:
					// Notify application when all channels are subscribed.
					fmt.Printf("m2:%v", n.Channel)
					fmt.Println("监听成功")
				case 0:
					fmt.Println("监听消息错误")
					// Return from the goroutine when all channels are unsubscribed.
					done <- nil
					return
				}
			}
		}
	}()
	ticker := time.NewTicker(healthCheckPeriod)
	defer ticker.Stop()
loop:
	for err == nil {
		select {
		case <-ticker.C:
			// Send ping to test health of connection and server. If
			// corresponding pong is not received, then receive on the
			// connection will timeout and the receive goroutine will exit.
			if err = psc.Ping(""); err != nil {
				break loop
			}
		case <-ctx.Done():
			break loop
		case err := <-done:
			// Return error from the receive goroutine.
			fmt.Print(err)
			return
		}
	}

	// Signal the receiving goroutine to exit by unsubscribing from all channels.
	psc.Unsubscribe()
	// Wait for goroutine to complete.
	return
}
