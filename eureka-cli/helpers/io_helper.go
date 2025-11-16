package helpers

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/folio-org/eureka-cli/constant"
	appErrors "github.com/folio-org/eureka-cli/errors"
)

func ReadJSONFromFile(filePath string, data any) error {
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

func WriteJSONToFile(filePath string, packageJSON any) error {
	jsonFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer CloseFile(jsonFile)

	writer := bufio.NewWriter(jsonFile)
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(packageJSON)
	if err != nil {
		return err
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

func CopySingleFile(srcPath, dstPath string) error {
	err := IsRegularFile(srcPath)
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

	return nil
}

func CloseFile(file *os.File) {
	_ = file.Close()
}

func CopyMultipleFiles(homeDir string, srcFs *embed.FS) error {
	return fs.WalkDir(*srcFs, ".", func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		dstPath := filepath.Join(homeDir, path)
		if dir.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
		} else {
			content, err := fs.ReadFile(*srcFs, path)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, content, 0644); err != nil {
				return err
			}
			if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
				fmt.Println("Created file:", dstPath)
			}
		}

		return nil
	})
}

func IsRegularFile(fileName string) error {
	s, err := os.Stat(fileName)
	if err != nil {
		return err
	}
	if !s.Mode().IsRegular() {
		return appErrors.NotRegularFile(fileName)
	}

	return nil
}

func GetCurrentWorkDirPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return cwd, nil
}

func GetHomeMiscDir() (string, error) {
	homeDir, err := GetHomeDirPath()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, constant.DockerComposeWorkDir), nil
}

func GetHomeDirPath() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	homeDir := filepath.Join(userHome, constant.ConfigDir)
	if err = os.MkdirAll(homeDir, 0644); err != nil {
		return "", err
	}

	return homeDir, nil
}

func CloseReader(reader io.ReadCloser) {
	_ = reader.Close()
}
