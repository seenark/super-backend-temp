package handler

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/atiwataqs/super-backend/cloudstorage"
	"bitbucket.org/atiwataqs/super-backend/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type MetadataAndType struct {
	repository.Metadata
	TypeName      string `json:"type_name"`
	LogoImageName string `json:"logo_image_name"`
	TypeOfUnit    string `json:"type_of_unit"`
	Unit          string `json:"unit"`
	VintageYear   string `json:"vintage_year"`
}

func NewMetadataHandler(router fiber.Router, metadataRepo repository.IMetadataRepository, certTypeRepo repository.IDigitalCertTypeRepository, uploader *cloudstorage.ClientUploader) {

	// create
	router.Post("/", func(c *fiber.Ctx) error {

		typeCode := c.FormValue("type_code")
		if typeCode == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "type_code invalid"})
		}

		certType, err := certTypeRepo.GetByTypeCode(typeCode)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "not found provided type_code please create it first"})
		}

		imageName := uuid.New()
		file, err := c.FormFile("image")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "image file invalid"})
		}

		// fullName, err := saveImage(imageName.String(), file, c)
		// if err != nil {
		// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "could not save image file please try again"})
		// }

		newFile, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "image file invalid"})
		}

		certIdStr := c.FormValue("cert_id")
		if certIdStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cert_id invalid"})
		}
		certId, err := strconv.Atoi(certIdStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cert_id invalid"})
		}
		projectName := c.FormValue("project_name")
		// if projectName == "" {
		// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "project_name invalid"})
		// }
		projectType := c.FormValue("project_type")
		// if projectType == "" {
		// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "project_type invalid"})
		// }
		description := c.FormValue("description")

		// listedDate := c.FormValue("listed_date")
		// if listedDate == "" {
		// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "listed_date invalid"})
		// }
		listedDate := time.Now().Format(time.RFC3339)
		ext := filepath.Ext(file.Filename)
		fullName := fmt.Sprintf("%s%s", imageName, ext)
		metadata, err := metadataRepo.Create(typeCode, certId, projectName, projectType, fullName, description, listedDate)
		if err != nil {
			// deleteImage(fullName)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "could not create new metadata please try another certId"})
		}

		err = uploader.UploadFile(newFile, fullName)
		if err != nil {

			fmt.Println("upload to cloud error")
			fmt.Printf("err: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "upload to cloud error"})
		}

		combine := MetadataAndType{
			Metadata:      *metadata,
			TypeName:      certType.TypeName,
			LogoImageName: certType.LogoImageName,
			TypeOfUnit:    certType.TypeOfUnit,
			Unit:          certType.Unit,
			VintageYear:   certType.VintageYear,
		}

		return c.JSON(combine)
	})

	// get all
	router.Get("/", func(c *fiber.Ctx) error {
		ids := []int{}
		idsStr := c.Query("ids")
		idSplit := strings.Split(idsStr, ",")
		for _, v := range idSplit {
			id, err := strconv.Atoi(strings.Trim(v, " "))
			if err != nil {
				log.Println(err)
			}
			ids = append(ids, id)
		}
		metadatas := metadataRepo.GetAll(ids)
		allCertType := certTypeRepo.GetAll()
		combines := []MetadataAndType{}
		for _, v := range metadatas {
			_, certType := certTypeRepo.FindInArrayByTypeCode(v.TypeCode, allCertType)
			metadataAndType := MetadataAndType{
				Metadata:      v,
				TypeName:      certType.TypeName,
				LogoImageName: certType.LogoImageName,
				TypeOfUnit:    certType.TypeOfUnit,
				Unit:          certType.Unit,
				VintageYear:   certType.VintageYear,
			}
			combines = append(combines, metadataAndType)
		}
		return c.JSON(combines)
	})

	// get by digital_cert_id
	router.Get("/:certId", func(c *fiber.Ctx) error {
		certIdStr := c.Params("certId")
		if certIdStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "digital certificate id is empty"})
		}
		certId, err := strconv.Atoi(certIdStr)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "digital certificate id is invalid"})
		}
		cert, err := metadataRepo.GetByDigitalCertId(certId)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusNotFound).JSON(err)
		}

		certType, err := certTypeRepo.GetByTypeCode(cert.TypeCode)
		if err != nil {
			log.Println(err)
			return c.JSON(MetadataAndType{
				Metadata:      *cert,
				TypeName:      "",
				LogoImageName: "",
				TypeOfUnit:    "",
				Unit:          "",
				VintageYear:   "",
			})
		}

		return c.JSON(MetadataAndType{
			Metadata:      *cert,
			TypeName:      certType.TypeName,
			LogoImageName: certType.LogoImageName,
			TypeOfUnit:    certType.TypeOfUnit,
			Unit:          certType.Unit,
			VintageYear:   certType.VintageYear,
		})
	})

	// get by type_code
	router.Get("/by-type-code/:typeCode", func(c *fiber.Ctx) error {
		typeCode := c.Params("typeCode")
		fmt.Printf("typeCode: %v\n", typeCode)
		if typeCode == "" {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		medataDataAndTypes := metadataRepo.GetByTypeCode(typeCode)
		allCertType := certTypeRepo.GetAll()
		all := []MetadataAndType{}
		for _, v := range medataDataAndTypes {
			_, certType := certTypeRepo.FindInArrayByTypeCode(v.TypeCode, allCertType)
			metadaAndType := MetadataAndType{
				Metadata:      v,
				TypeName:      certType.TypeName,
				LogoImageName: certType.LogoImageName,
				TypeOfUnit:    certType.TypeOfUnit,
				Unit:          certType.Unit,
				VintageYear:   certType.VintageYear,
			}
			all = append(all, metadaAndType)
		}
		return c.JSON(all)
	})

	// update by digital_cert_id
	router.Patch("/:certId", func(c *fiber.Ctx) error {
		certIdStr := c.Params("certId")
		if certIdStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cert id invalid"})
		}
		if certIdStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cert_id invalid"})
		}
		certId, err := strconv.Atoi(certIdStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cert_id invalid"})
		}

		notFoundMetadata := false
		cert, err := metadataRepo.GetByDigitalCertId(certId)
		if err != nil {
			log.Println(err)
			log.Println("While update metadata there is no metadata so create new one")
			notFoundMetadata = true
			cert = &repository.Metadata{
				TypeCode:      "",
				DigitalCertID: certId,
				ProjectName:   "",
				ProjectType:   "",
				ImageName:     "",
				Description:   "",
				ListedDate:    "",
			}
			// return c.Status(fiber.StatusNotFound).JSON(err)
		}

		typeCode := c.FormValue("type_code")
		if typeCode == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "type_code invalid"})
		}
		cert.TypeCode = typeCode

		certType, err := certTypeRepo.GetByTypeCode(typeCode)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "not found provided type_code please create it first"})
		}

		imageName := uuid.New()
		file, err := c.FormFile("image")
		oldImageName := ""
		if err == nil {
			newFile, err := file.Open()
			if err == nil {
				oldImageName = cert.ImageName
				ext := filepath.Ext(file.Filename)
				fullName := fmt.Sprintf("%s%s", imageName, ext)
				err := uploader.UploadFile(newFile, fullName)
				// fullName, err := saveImage(imageName.String(), file, c)
				if err != nil {
					log.Println(err)
				}
				cert.ImageName = fullName
			}
		}

		projectName := c.FormValue("project_name")
		cert.ProjectName = projectName
		// if projectName != "" {
		// 	cert.ProjectName = projectName
		// }
		projectType := c.FormValue("project_type")
		cert.ProjectType = projectType
		// if projectType != "" {
		// 	cert.ProjectType = projectType
		// }
		description := c.FormValue("project_type")
		cert.Description = description
		// if description != "" {
		// 	cert.Description = description
		// }
		// listedDate := c.FormValue("listed_date")
		// if listedDate != "" {
		// 	cert.ListedDate = listedDate
		// }
		if !notFoundMetadata {

			newCert, err := metadataRepo.UpdateByDigitalCertId(certId, *cert)
			if err != nil {
				log.Println(err)
				// _, err := deleteImage(cert.ImageName)
				// if err != nil {
				// 	log.Println(err)
				// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "error while deleting new image"})
				// }
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "error while update metadata"})
			}
			// _, err = deleteImage(oldImageName)
			err = uploader.DeleteFile(oldImageName)
			if err != nil {
				log.Println("error while delete old image")
			}
			return c.JSON(MetadataAndType{
				Metadata:      *newCert,
				TypeName:      certType.TypeName,
				LogoImageName: certType.LogoImageName,
				TypeOfUnit:    certType.TypeOfUnit,
				Unit:          certType.Unit,
				VintageYear:   certType.VintageYear,
			})
		} else {
			newCert, err := metadataRepo.Create(cert.TypeCode, cert.DigitalCertID, cert.ProjectName, cert.ProjectType, cert.ImageName, cert.Description, time.Now().Format(time.RFC3339))
			if err != nil {
				log.Println(err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err})
			}
			return c.JSON(MetadataAndType{
				Metadata:      *newCert,
				TypeName:      certType.TypeName,
				LogoImageName: certType.LogoImageName,
				TypeOfUnit:    certType.TypeOfUnit,
				Unit:          certType.Unit,
				VintageYear:   certType.VintageYear,
			})
		}
	})
}
