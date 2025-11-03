package helpers

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/folio-org/eureka-cli/constant"
	appErrors "github.com/folio-org/eureka-cli/errors"
)

func ReadJsonFromFile(actionName string, filePath string, data any) error {
	jsonFile, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer CloseFile(jsonFile)

	err = json.NewDecoder(jsonFile).Decode(&data)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}

func WriteJsonToFile(filePath string, packageJson any) error {
	jsonFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer CloseFile(jsonFile)

	writer := bufio.NewWriter(jsonFile)
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(packageJson)
	if err != nil {
		return err
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

func CopySingleFile(actionName, srcPath, dstPath string) error {
	err := CheckIsRegularFile(actionName, srcPath)
	if err != nil {
		return err
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer CloseFile(src)

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer CloseFile(dst)

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	slog.Info(actionName, "text", "Copied a single file", "srcPath", filepath.FromSlash(srcPath), "dstPath", filepath.FromSlash(dstPath))

	return nil
}

func CheckIsRegularFile(actionName, fileName string) error {
	fileStat, err := os.Stat(fileName)
	if err != nil {
		return err
	}

	if !fileStat.Mode().IsRegular() {
		return appErrors.NotRegularFile(fileName)
	}

	return nil
}

func GetCurrentWorkDirPath(actionName string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return cwd, nil
}

func GetHomeMiscDir(actionName string) (string, error) {
	homeDir, err := GetHomeDirPath()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, constant.DockerComposeWorkDir), nil
}

func GetHomeDirPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	homeDir := filepath.Join(home, constant.ConfigDir)
	if err = os.MkdirAll(homeDir, 0644); err != nil {
		return "", err
	}

	return homeDir, nil
}

func CloseFile(file *os.File) {
	_ = file.Close()
}

func CloseReader(reader io.ReadCloser) {
	_ = reader.Close()
}
