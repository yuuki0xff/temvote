package main

import (
	"bufio"
	"os"
	"strings"
)

type SecretManager struct {
	secretMap map[string]string
}

func NewSecretManager(secretFile string) (*SecretManager, error) {
	sm := &SecretManager{}
	sm.secretMap = make(map[string]string)

	f, err := os.Open(secretFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		words := strings.SplitN(line, " ", 2)

		name := strings.TrimSpace(words[0])
		secret := strings.TrimSpace(words[1])
		sm.secretMap[name] = secret
	}
	return sm, nil
}

func (sm *SecretManager) hasAuth(hostid, secret string) bool {
	return secret != "" && sm.secretMap[hostid] == secret
}
