package handler

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

const StaticPath = "../static/images"

func saveImage(filename string, file *multipart.FileHeader, c *fiber.Ctx) (string, error) {

	ext := filepath.Ext(file.Filename)
	fullName := fmt.Sprintf("%s%s", filename, ext)
	return fullName, c.SaveFile(file, fmt.Sprintf("%s/%s%s", StaticPath, filename, ext))
}

func deleteImage(fileName string) (string, error) {

	err := os.Remove(fmt.Sprintf("%s/%s", StaticPath, fileName))
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return "", err
	}
	return fileName, nil
}
