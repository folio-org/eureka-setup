package helpers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
)

func ReadJsonFromFile(action *action.Action, filePath string, data any) error {
	jsonFile, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = jsonFile.Close()
	}()

	decoder := json.NewDecoder(jsonFile)
	for {
		if err := decoder.Decode(&data); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
	}

	return nil
}

func WriteJsonToFile(action *action.Action, filePath string, packageJson any) error {
	jsonFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = jsonFile.Close()
	}()

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

func CheckIsRegularFile(action *action.Action, fileName string) error {
	fileStat, err := os.Stat(fileName)
	if err != nil {
		return err
	}

	if !fileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", fileName)
	}

	return nil
}

func CopySingleFile(action *action.Action, srcPath string, dstPath string) error {
	CheckIsRegularFile(action, srcPath)

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = src.Close()
	}()

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = dst.Close()
	}()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	slog.Info(action.Name, "text", fmt.Sprintf("Copied a single file from %s to %s", filepath.FromSlash(srcPath), filepath.FromSlash(dstPath)))

	return nil
}

func GetCurrentWorkDirPath(action *action.Action) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return cwd, nil
}

func GetHomeMiscDir(action *action.Action) (string, error) {
	homeDir, err := GetHomeDirPath(action)
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, constant.DockerComposeWorkDir), nil
}

func GetHomeDirPath(action *action.Action) (string, error) {
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
