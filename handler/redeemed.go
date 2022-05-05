package handler

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/seenark/super-backend-temp/repository"
)

func NewRedeemedHandler(router fiber.Router, redeemedRepo repository.IRedeemedRepository) {

	router.Post("/", func(c *fiber.Ctx) error {
		redeemed := repository.Redeemed{}
		err := c.BodyParser(&redeemed)
		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusBadRequest)
		}
		fmt.Printf("redeemed: %v\n", redeemed)
		redeemed.Amount = 0
		redeemed.Price = ""
		redeemed.RedeemDate = 0
		redeemed.RedeemId = 0
		redeemed.WalletAddress = ""
		newRedeemed, err := redeemedRepo.Upsert(redeemed)
		if err != nil {
			log.Println(err)
			c.Status(fiber.StatusBadRequest).JSON(err.Error())
		}

		return c.JSON(newRedeemed)
	})

	router.Post("/redeem-event", RequiredValidJWT, RequiredEventLoggerRole, func(c *fiber.Ctx) error {
		redeemed := repository.Redeemed{}
		err := c.BodyParser(&redeemed)
		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusBadRequest)
		}
		redeemed.Company = ""
		redeemed.Email = ""
		redeemed.Name = ""
		redeemed.TaxID = ""
		redeemed.Telephone = ""
		newRedeemed, err := redeemedRepo.Upsert(redeemed)
		if err != nil {
			log.Println(err)
			c.Status(fiber.StatusBadRequest).JSON(err.Error())
		}

		return c.JSON(newRedeemed)

	})

	router.Patch("/update-status", RequiredValidJWT, RequiredAdminRole, func(c *fiber.Ctx) error {
		type CertIdAndStatus struct {
			RedeemedId    int    `json:"redeemed_id"`
			ApproveStatus string `json:"approve_status"`
		}
		certIdAndStatus := CertIdAndStatus{}
		err := c.BodyParser(&certIdAndStatus)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "json is invalid"})
		}

		redeemed, err := redeemedRepo.UpdateStatus(certIdAndStatus.RedeemedId, certIdAndStatus.ApproveStatus)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "could not change status or status is not changed"})
		}
		return c.JSON(redeemed)
	})

	router.Get("/", func(c *fiber.Ctx) error {
		approveStatusesStr := c.Query("approveStatus")
		statuses := []string{}
		if approveStatusesStr != "" {
			splitApproveStatus := strings.Split(approveStatusesStr, ",")
			for _, v := range splitApproveStatus {
				statuses = append(statuses, strings.Trim(v, " "))
			}
		}

		var startDate int = 0
		startDateUnixStr := c.Query("redeemStartDate")
		if startDateUnixStr != "" {
			startDateInt, err := strconv.Atoi(startDateUnixStr)
			if err != nil {
				fmt.Printf("err: %v\n", err)
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "redeemStartDate is invalid"})
			}
			startDate = startDateInt
		}
		var endDate int = 0
		endDateUnixStr := c.Query("redeemEndDate")
		if endDateUnixStr != "" {
			endDateInt, err := strconv.Atoi(endDateUnixStr)
			if err != nil {
				fmt.Printf("err: %v\n", err)
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "redeemEndDate is invalid"})
			}
			endDate = endDateInt
		}

		redeemId := []int{}
		redeemIdStr := c.Query("redeemIds")
		if redeemIdStr != "" {
			idSplit := strings.Split(redeemIdStr, ",")
			fmt.Println(idSplit)
			for _, v := range idSplit {
				id, err := strconv.Atoi(strings.Trim(v, " "))
				if err != nil {
					log.Println(err)
				}
				redeemId = append(redeemId, id)
			}
		}

		name := c.Query("name")
		email := c.Query("email")
		walletAddress := c.Query("walletAddress")

		all := redeemedRepo.GetAll(statuses, name, email, redeemId, startDate, endDate, walletAddress)
		return c.JSON(all)
	})
}
