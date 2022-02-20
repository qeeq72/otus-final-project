package file

import (
	"bufio"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

func ReadLineFromFile(path string) ([]byte, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	line, _, err := reader.ReadLine()
	if err != nil {
		return nil, err
	}
	return line, nil
}

func ReadYamlFile(path string, out interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, out)
	if err != nil {
		return err
	}
	return nil
}
