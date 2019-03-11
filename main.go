package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

type DomainProvider interface {
	AuthValid() (bool, error)
	SetRecord(domain string, ip string) error
}

func parseArgs() (domain string, token string) {
	flag.StringVarP(&domain, "domain", "d", "", "domain to post to")
	flag.StringVarP(&token, "token", "t", "", "API Token to use")
	var verbose = flag.BoolP("verbose", "v", false, "turn on debug logging")

	flag.Parse()
	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	log.Debugf("Parsed domain: %s\n", domain)

	var err error
	token, err = getToken(token)
	log.Debugf("Token: %s\n", token)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: No Token provided")
		os.Exit(1)
	}

	return
}

func main() {
	var lastIP string
	domain, token := parseArgs()

	// todo: implement any other provider
	// by fulfilling DomainProvider interface
	l := Linode{Token: token}
	tick := time.Tick(time.Hour * 4)
	for ; true; <-tick {
		ip, err := getIP()
		if err != nil {
			log.Errorf("Error determining IP: %v\n", err)
			continue
		}
		if ip != lastIP {
			update(l, domain, ip)
		}
		lastIP = ip
	}
}

func update(d DomainProvider, domain, ip string) {
	//validate token
	valid, err := d.AuthValid()
	if err != nil {
		log.Errorf("Error authenticating: %v\n", err)
		return
	}
	if !valid {
		log.Info("Authentication failed.")
		return
	}

	d.SetRecord(domain, ip)
}

func getToken(token string) (string, error) {
	env := strings.TrimSpace(os.Getenv("TOKEN"))
	if env != "" {
		return env, nil
	}
	tk := strings.TrimSpace(token)
	if tk != "" {
		return tk, nil
	}
	return "", errors.New("No Token Provided")
}

func getIP() (string, error) {
	r := net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			// @todo lookup this IP with normal DNS resolver?
			return d.DialContext(ctx, network, "208.67.222.222:53") // resolver1.opendns.com
		},
	}
	ctx := context.Background()
	addr, err := r.LookupHost(ctx, "myip.opendns.com")
	if err != nil {
		return "", err
	}

	if len(addr) == 0 {
		return "", errors.New("no address found")
	}

	log.Infof("Got IP Address: %s\n", addr[0])
	return addr[0], nil
}
