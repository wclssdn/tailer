package lib

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"os"
	"time"
)

func SshConnect(host string) (*ssh.Client, error) {
	auth := viper.GetStringMapString("auth")
	user, ok := auth["user"]
	if !ok {
		return nil, errors.New("no auth.user in config file")
	}
	var authMethods []ssh.AuthMethod
	pass, ok := auth["pass"]
	privateKeyFile, ok2 := auth["private_key"]
	if !ok && !ok2 {
		return nil, errors.New("auth.pass or auth.private_key must have one in config file")
	}
	if ok {
		authMethods = append(authMethods, ssh.Password(pass))
	}
	if ok2 {
		privateKey, err := os.Open(privateKeyFile)
		if err != nil {
			return nil, errors.New("can't open auth.private_key file. error:" + err.Error())
		}
		privateKeyContent, err := ioutil.ReadAll(privateKey)
		if err != nil {
			return nil, errors.New("can't read auth.private_key file. error:" + err.Error())
		}
		signer, err := ssh.ParsePrivateKey(privateKeyContent)
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	conf := &ssh.ClientConfig{
		User: user,
		Auth: authMethods,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: time.Second * 3,
	}
	conf.SetDefaults()
	return ssh.Dial("tcp", fmt.Sprintf("%s:22", host), conf)
}

func SshSession(host string) (*ssh.Session, error) {
	client, err := SshConnect(host)
	if err != nil {
		return nil, err
	}
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	err = session.RequestPty("xterm", 25, 100, modes)
	if err != nil {
		return session, err
	}
	return session, nil
}
