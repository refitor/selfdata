package main

import "sdcore"

func SDAuth() func([]byte) error {
	return sdcore.SDAuth()
}

func SDEncrypt(data []byte) []byte {
	return sdcore.SDEncrypt(data)
}

func SDDecrypt(data []byte) []byte {
	return sdcore.SDDecrypt(data)
}

func SDPush(selfID, fileName string, data []byte) (bool, error) {
	return sdcore.SDPush(selfID, fileName, data)
}

func SDPull(selfID string) (bool, []string, error) {
	return sdcore.SDPull(selfID)
}