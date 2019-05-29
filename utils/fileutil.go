package utils

import "os"

func PathExist(path string) error {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return err
	}
	return nil
}
