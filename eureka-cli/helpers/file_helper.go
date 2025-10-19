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

func ReadJsonFromFile(action *action.Action, filePath string, data any) {
	jsonFile, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		slog.Error(action.Name, "error", err)
		panic(err)
	}
	defer func() {
		_ = jsonFile.Close()
	}()

	decoder := json.NewDecoder(jsonFile)
	for {
		if err := decoder.Decode(&data); err == io.EOF {
			break
		} else if err != nil {
			slog.Error(action.Name, "error", err)
			panic(err)
		}
	}
}

func WriteJsonToFile(action *action.Action, filePath string, packageJson any) {
	jsonFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		slog.Error(action.Name, "error", err)
		panic(err)
	}
	defer func() {
		_ = jsonFile.Close()
	}()

	writer := bufio.NewWriter(jsonFile)
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err = encoder.Encode(packageJson); err != nil {
		slog.Error(action.Name, "error", err)
		panic(err)
	}

	_ = writer.Flush()
}

func CheckIsRegularFile(action *action.Action, fileName string) {
	fileStat, err := os.Stat(fileName)
	if err != nil {
		slog.Error(action.Name, "error", err)
		panic(err)
	}

	if !fileStat.Mode().IsRegular() {
		LogErrorPanic(action, fmt.Errorf("%s is not a regular file", fileName))
	}
}

func CopySingleFile(action *action.Action, srcPath string, dstPath string) {
	CheckIsRegularFile(action, srcPath)

	src, err := os.Open(srcPath)
	if err != nil {
		slog.Error(action.Name, "error", err)
		panic(err)
	}
	defer func() {
		_ = src.Close()
	}()

	dst, err := os.Create(dstPath)
	if err != nil {
		slog.Error(action.Name, "error", err)
		panic(err)
	}
	defer func() {
		_ = dst.Close()
	}()

	_, err = io.Copy(dst, src)
	if err != nil {
		slog.Error(action.Name, "error", err)
		panic(err)
	}

	slog.Info(action.Name, "text", fmt.Sprintf("Copied a single file from %s to %s", filepath.FromSlash(srcPath), filepath.FromSlash(dstPath)))
}

func GetCurrentWorkDirPath(action *action.Action) string {
	cwd, err := os.Getwd()
	if err != nil {
		slog.Error(action.Name, "error", err)
		panic(err)
	}

	return cwd
}

func GetHomeDirPath(action *action.Action) string {
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error(action.Name, "error", err)
		panic(err)
	}

	homeDir := filepath.Join(home, constant.ConfigDir)

	if err = os.MkdirAll(homeDir, 0644); err != nil {
		slog.Error(action.Name, "error", err)
		panic(err)
	}

	return homeDir
}

func GetHomeMiscDir(action *action.Action) string {
	return filepath.Join(GetHomeDirPath(action), constant.DockerComposeWorkDir)
}
