package handler

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/seenark/super-backend-temp/cloudstorage"
	"github.com/seenark/super-backend-temp/repository"
)

func NewDigitalCertTypeHandler(router fiber.Router, certTypeRepo repository.IDigitalCertTypeRepository, uploader *cloudstorage.ClientUploader) {

	router.Post("/", func(c *fiber.Ctx) error {

		typeCode := c.FormValue("type_code")
		if typeCode == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "type_code is invalid"})
		}
		typeName := c.FormValue("type_name")
		if typeName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "type_name is invalid"})
		}
		logoFile, err := c.FormFile("logo_file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "logo_file is invalid"})
		}
		newFile, err := logoFile.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "image file invalid"})
		}
		logoName := uuid.New()
		// fullName, err := saveImage(logoName.String(), logoFile, c)
		// if err != nil {
		// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "save logo image error"})
		// }
		typeOfUnit := c.FormValue("type_of_unit")
		if typeOfUnit == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "type_of_unit is invalid"})
		}
		unit := c.FormValue("unit")
		if typeCode == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "unit is invalid"})
		}
		vintageYear := c.FormValue("vintage_year")
		ext := filepath.Ext(logoFile.Filename)
		fullName := fmt.Sprintf("%s%s", logoName, ext)
		certType, err := certTypeRepo.Create(typeCode, typeName, fullName, typeOfUnit, unit, vintageYear)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			if strings.Contains(err.Error(), "E11000") {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "type_code is duplicate"})
			}
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "error while create new digital cert type"})
		}

		err = uploader.UploadFile(newFile, fullName)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": fmt.Sprintf("upload image error please upload later %v", err)})
		}
		return c.JSON(certType)
	})

	router.Get("/", func(c *fiber.Ctx) error {
		allCertTypes := certTypeRepo.GetAll()
		return c.JSON(allCertTypes)
	})

	router.Get("/:typeCode", func(c *fiber.Ctx) error {
		typeCode := c.Params("typeCode")
		certType, err := certTypeRepo.GetByTypeCode(typeCode)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(err)
		}
		return c.JSON(certType)
	})

	router.Patch("/:typeCode", func(c *fiber.Ctx) error {
		typeCode := c.Params("typeCode")

		certType, err := certTypeRepo.GetByTypeCode(typeCode)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"message": "type_code is not exist"})
		}

		typeName := c.FormValue("type_name")
		if typeName != "" {
			certType.TypeName = typeName
		}
		logoFile, err := c.FormFile("logo_file")

		if err == nil {
			logoName := uuid.New()
			ext := filepath.Ext(logoFile.Filename)
			fullName := fmt.Sprintf("%s%s", logoName, ext)
			// fullName, err := saveImage(logoName.String(), logoFile, c)
			newFile, err := logoFile.Open()
			if err == nil {
				err = uploader.UploadFile(newFile, fullName)
				if err == nil {
					// deleteImage(certType.LogoImageName)
					err := uploader.DeleteFile(certType.LogoImageName)
					if err != nil {
						log.Println(err)
					}
					certType.LogoImageName = fullName
				}

			}
		}
		typeOfUnit := c.FormValue("type_of_unit")
		if typeOfUnit != "" {
			certType.TypeOfUnit = typeOfUnit
		}
		unit := c.FormValue("unit")
		if typeCode != "" {
			certType.Unit = unit
		}
		vintageYear := c.FormValue("vintage_year")
		if vintageYear != "" {
			certType.VintageYear = vintageYear
		}

		newCertType, err := certTypeRepo.Update(typeCode, *certType)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "error while update"})
		}
		return c.JSON(newCertType)
	})

	router.Delete("/:typeCode", func(c *fiber.Ctx) error {
		typeCode := c.Params("typeCode")

		certType, err := certTypeRepo.Delete(typeCode)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusBadRequest).JSON(err)
		}
		deleteImage(certType.LogoImageName)
		return c.JSON(certType)
	})
}
