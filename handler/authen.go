package handler

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/seenark/super-backend-temp/authen"
	"github.com/seenark/super-backend-temp/repository"
)

func NewAuthHandler(router fiber.Router, authenDb repository.AuthenticationRepository) {

	router.Post("/signin", func(c *fiber.Ctx) error {

		email := c.FormValue("email")
		password := c.FormValue("password")

		if email == "" || password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid email or password"})
		}
		token, err := authenDb.Signin(email, password)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid email or password"})
		}
		userRes := UserResponse{
			Id:              token.User.Id.Hex(),
			Name:            token.User.Name,
			Email:           token.User.Email,
			Role:            token.User.Role,
			MetamaskAddress: token.User.MetamaskAddress,
			Address:         token.User.Address,
			Tel:             token.User.Tel,
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"user": userRes, "token": token.Token, "refresh": token.Refresh})
	})

	/*
		func description require body: {
			refresh: <refresh-token>
		}
	*/
	router.Post("/refresh-access-token", func(c *fiber.Ctx) error {
		bearer := c.Get("Authorization")
		if bearer == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "no jwt token found in header"})
		}
		jwt := strings.Split(bearer, " ")[1]
		if jwt == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "no jwt token found in header"})
		}

		refreshMap := map[string]string{}
		c.BodyParser(&refreshMap)
		rt := refreshMap["refresh"]
		claims, err := authen.ValidateRefreshJWT(rt)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "refresh token is invalid"})
		}

		originalJWT := claims["jwt"]
		if jwt != originalJWT {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "refresh token is invalid"})
		}
		newAccessToken, err := authenDb.NewAccessTokenAndRefreshToken(fmt.Sprintf("%s", claims["id"]), rt)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": fmt.Sprintf("refresh token is invalid -> %s", err.Error())})
		}
		userRes := UserResponse{
			Id:              newAccessToken.User.Id.Hex(),
			Name:            newAccessToken.User.Name,
			Email:           newAccessToken.User.Email,
			Role:            newAccessToken.User.Role,
			MetamaskAddress: newAccessToken.User.MetamaskAddress,
			Address:         newAccessToken.User.Address,
			Tel:             newAccessToken.User.Tel,
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"user": userRes, "token": newAccessToken.Token, "refresh": newAccessToken.Refresh})
	})

	router.Post("/signout", RequiredValidJWT, func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").([]string)
		if !ok {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		err := authenDb.Signout(user[0])
		return err
	})

	router.Get("/verify-token", RequiredValidJWT, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"valid": true})
	})

}

func RequiredValidJWT(c *fiber.Ctx) error {
	bearer := c.Get("Authorization")
	if bearer == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "no jwt token found in header"})
	}
	jwt := strings.Split(bearer, " ")[1]
	if jwt == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "no jwt token found in header"})
	}
	claims, err := authen.ValidateJWT(jwt)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "jwt invalid"})
	}
	id := fmt.Sprintf("%v", claims["id"])
	fmt.Printf("id: %v\n", id)
	email := fmt.Sprintf("%v", claims["id"])
	metamask := fmt.Sprintf("%v", claims["metamask_address"])
	role := fmt.Sprintf("%v", claims["role"])
	c.Locals("user", []string{id, email, metamask, role})
	return c.Next()
}

/* must call after middleware NeedValidJWT */
func RequiredAdminRole(c *fiber.Ctx) error {
	userData, ok := c.Locals("user").([]string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "you are not admin"})
	}
	role := userData[3]
	if role != "admin" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "you are not admin"})
	}
	return c.Next()
}

/* must call after middleware NeedValidJWT */
func RequiredEventLoggerRole(c *fiber.Ctx) error {
	userData, ok := c.Locals("user").([]string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "you are not eventLogger role"})
	}
	role := userData[3]
	if role != "eventLogger" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "you are not eventLogger role"})
	}
	return c.Next()
}
