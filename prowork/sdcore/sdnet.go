package sdcore

// By default, the official website provides mail sending, and external users can specify their own dynamic authentication method
func SDAuth() func(data []byte) error {
	return nil
}

func SDPull(selfID string) (bool, []string, error) {
	return false, nil, nil
}

func SDPush(selfID, fileName string, data []byte) (bool, error) {
	return false, nil
}