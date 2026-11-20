// SPDX-License-Identifier: GPL-2.0-only
// Copyright (C) 2026, Shu De Zheng <imchuncai@gmail.com>. All Rights Reserved.

package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	client "github.com/imchuncai/umem-cache-client-Go"
	"github.com/imchuncai/umem-cache-client-Go/proto"
	"github.com/peterh/liner"
)

var COMMANDS = []string{"help", "exit", "clear", "get", "set", "del"}

type ClientInterface interface {
	GetOrSet(key []byte, get proto.FallbackGetFunc) ([]byte, error)
	Del(key []byte) error
	Close()
}

func printHelp() {
	fmt.Printf("Supported commands: %s\n", strings.Join(COMMANDS, ", "))
}

func printErrorMsg(msg string) {
	fmt.Fprintf(os.Stderr, "\033[31m[ERROR]\033[0m %s\n", msg)
}

func printError(err error) {
	fmt.Fprintf(os.Stderr, "\033[31m[ERROR]\033[0m %s\n", err.Error())
}

func parseConfig() (client.Config, error) {
	timeout := flag.Int("timeout", 2, "timeout seconds, default 2 seconds")
	tlsKey := flag.String("tls-key", "", `tls key (default "")`)
	tlsCert := flag.String("tls-cert", "", `tls cert (default "")`)
	tlsCaCert := flag.String("tls-ca-cert", "", `tls ca cert (default "")`)
	flag.Parse()

	tlsFull := *tlsKey != "" && *tlsCert != "" && *tlsCaCert != ""
	tlsEmpty := *tlsKey == "" && *tlsCert == "" && *tlsCaCert == ""
	if tlsFull == tlsEmpty {
		return client.Config{}, errors.New("bad tls config")
	}

	var tlsConfig *tls.Config
	if tlsFull {
		cert, err := tls.LoadX509KeyPair(*tlsCert, *tlsKey)
		if err != nil {
			return client.Config{}, fmt.Errorf("load tls_key and tls_cert failed: %w", err)
		}

		caCert, err := os.ReadFile(*tlsCaCert)
		if err != nil {
			return client.Config{}, fmt.Errorf("load tls_ca_cert failed: %w", err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		}
	}
	return client.Config{time.Duration(*timeout) * time.Second, 0, tlsConfig}, nil
}

func usage() {
	fmt.Printf(`
Usage:
  umem-cache-toy [-tls_key <path> -tls_cert <path> -tls_ca_cert <path>]
                 [-timeout <seconds>] <address...>

Example:
  umem-cache-toy [::1]:10047
`)
}

func main() {
	flag.Usage = usage

	config, err := parseConfig()
	if err != nil {
		printError(err)
		return
	}

	addresses := flag.Args()
	if len(addresses) == 0 {
		usage()
		return
	}

	var cli ClientInterface
	cli, err = client.NewCluster(addresses, config)
	if err != nil {
		cli, err = client.New(addresses[0], config)
		if err != nil {
			printError(fmt.Errorf("new client failed: %w", err))
			return
		}
		fmt.Println("umem-cache-toy - Singleton\n")
	} else {
		fmt.Println("umem-cache-toy - Cluster\n")
	}
	printHelp()

	__line := liner.NewLiner()
	defer __line.Close()

	__line.SetCtrlCAborts(true)
	__line.SetCompleter(func(line string) (ret []string) {
		for _, command := range COMMANDS {
			if strings.HasPrefix(command, strings.ToLower(line)) {
				ret = append(ret, command)
			}
		}
		return
	})

	for {
		line, err := __line.Prompt("umem-cache> ")
		if err != nil {
			if err == liner.ErrPromptAborted {
				fmt.Println("Bye!")
				return
			}
			printError(fmt.Errorf("Error reading line: %w\n", err))
			continue
		}

		__line.AppendHistory(line)

		inputs := strings.Fields(line)
		if len(inputs) == 0 {
			continue
		}

		switch inputs[0] {
		case "help":
			printHelp()
		case "exit":
			fmt.Println("Bye!")
			return
		case "clear":
			fmt.Print("\033[H\033[2J\033[3J")
		case "get":
			if len(inputs) != 2 {
				printErrorMsg("bad get, want: get <key>")
				continue
			}

			val, err := cli.GetOrSet([]byte(inputs[1]), nil)
			if errors.Is(err, proto.ErrFallbackGet) {
				fmt.Println("Miss!")
			} else if err == nil {
				fmt.Println("Hit!\n" + string(val))
			} else {
				printError(fmt.Errorf("GetOrSet failed: %w", err))
			}
		case "del":
			if len(inputs) != 2 {
				printErrorMsg("bad del, want: del <key>")
				continue
			}

			err := cli.Del([]byte(inputs[1]))
			if err != nil {
				printError(fmt.Errorf("Del failed: %w", err))
			}
		case "set":
			if len(inputs) != 3 {
				printErrorMsg("bad set, want: set <key> <value>")
				continue
			}

			err := cli.Del([]byte(inputs[1]))
			if err != nil {
				printError(fmt.Errorf("Del failed: %w", err))
				continue
			}

			_, err = cli.GetOrSet([]byte(inputs[1]), func(key []byte) (val []byte, err error) {
				return []byte(inputs[2]), nil
			})
			if err != nil {
				printError(fmt.Errorf("GetOrSet failed: %w", err))
			}
		default:
			printError(fmt.Errorf("Unknown command: %s", line))
		}
	}
}
