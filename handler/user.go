package handler

import (
	"fmt"
	"log"
	"strings"

	"bitbucket.org/atiwataqs/super-backend/authen"
	"bitbucket.org/atiwataqs/super-backend/repository"
	"github.com/gofiber/fiber/v2"
)

type UserResponse struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Role            string `json:"role"`
	MetamaskAddress string `json:"metamask_address"`
	Address         string `json:"address"`
	Tel             string `json:"tel"`
}

func mapUserToUserResponse(user repository.User) *UserResponse {
	return &UserResponse{
		Id:              user.Id.Hex(),
		Name:            user.Name,
		Email:           user.Email,
		Role:            user.Role,
		MetamaskAddress: user.MetamaskAddress,
		Address:         user.Address,
		Tel:             user.Tel,
	}
}

func NewUserHandler(router fiber.Router, db repository.UserRepository) {
	// get all
	router.Get("/", func(c *fiber.Ctx) error {
		usersRes := []UserResponse{}
		users := db.GetAll()
		for _, v := range users {
			userRes := mapUserToUserResponse(v)
			usersRes = append(usersRes, *userRes)
		}
		return c.Status(fiber.StatusOK).JSON(usersRes)
	})

	// get by id
	router.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		user, err := db.GetById(id)
		if err != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}
		userRes := mapUserToUserResponse(*user)
		return c.Status(fiber.StatusOK).JSON(userRes)
	})

	// create admin
	router.Post("/create-admin", RequiredValidJWT, RequiredAdminRole, func(c *fiber.Ctx) error {
		user := repository.User{}
		err := c.BodyParser(&user)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid some field", "error": err})
		}
		if user.Password == "" || user.Email == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "email or password is empty"})
		}
		user.Role = "admin"
		newUser, err := db.Create(user)
		if err != nil {
			// check if there is some error code in error text
			if strings.Contains(err.Error(), "E11000") {
				return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"message": "email already taken"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "error while creating new user"})
		}
		userRes := mapUserToUserResponse(*newUser)
		return c.Status(fiber.StatusOK).JSON(userRes)
	})

	// create
	router.Post("/", func(c *fiber.Ctx) error {
		user := repository.User{}
		err := c.BodyParser(&user)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid some field", "error": err})
		}
		if user.Password == "" || user.Email == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "email or password is empty"})
		}
		user.Role = "user"
		newUser, err := db.Create(user)
		if err != nil {
			// check if there is some error code in error text
			if strings.Contains(err.Error(), "E11000") {
				return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"message": "email already taken"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "error while creating new user"})
		}
		userRes := mapUserToUserResponse(*newUser)
		return c.Status(fiber.StatusOK).JSON(userRes)
	})

	// reset-password
	router.Post("/admin/reset-password", RequiredValidJWT, func(c *fiber.Ctx) error {
		type ResetPassword struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}

		userArray, ok := c.Locals("user").([]string)
		if !ok {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		userId := userArray[0]
		resetPassword := ResetPassword{}
		err := c.BodyParser(&resetPassword)
		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusBadRequest)
		}

		if resetPassword.NewPassword == "" || resetPassword.OldPassword == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "password could not be empty"})
		}
		user, err := db.GetById(userId)
		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusNotFound)
		}
		passwordOk := authen.VerifyPassword(user.Password, resetPassword.OldPassword)
		if !passwordOk {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "password is incorrected"})
		}
		user.Password = resetPassword.NewPassword
		_, err = db.Update(*user)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
		return c.SendStatus(fiber.StatusOK)
	})

	// hashPassword
	router.Post("/hash-password", func(c *fiber.Ctx) error {
		password := struct {
			Password string `json:"password"`
		}{}
		err := c.BodyParser(&password)
		if err != nil {
			fmt.Println(err)
			return c.SendStatus(fiber.StatusBadRequest)
		}
		fmt.Printf("password.password: %v\n", password.Password)
		hash, err := authen.HashPassword(password.Password)
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		return c.JSON(hash)
	})

}
