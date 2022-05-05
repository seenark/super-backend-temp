package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/seenark/super-backend-temp/cloudstorage"
	"github.com/seenark/super-backend-temp/config"
	"github.com/seenark/super-backend-temp/handler"
	"github.com/seenark/super-backend-temp/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// initTimeZone()
	cfg := config.GetConfig()
	fmt.Printf("running on %s\n", cfg.Environment)
	app := fiber.New()
	newApp := app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.App.AllowOrigin,
		AllowHeaders: "*",
	}))
	newApp = newApp.Use(limiter.New(limiter.Config{
		Next: func(c *fiber.Ctx) bool {
			if c.IP() == "127.0.0.1" {
				fmt.Println("ip == 127.0.0.1")
				return true
			}
			return false
		},
		Max:        60,
		Expiration: 60 * time.Second,
	}))

	uploader := cloudstorage.NewGoogleStorageUploader()

	mongoClient := connectMongo(cfg.Mongo.Username, cfg.Mongo.Password, cfg.Mongo.URL)
	db := mongoClient.Database("super_energy")
	userCollection := db.Collection("users")
	makeUserEmailAsIndexes(userCollection)
	userDb := repository.NewUserDb(userCollection)
	authDb := repository.NewAuthDB(userCollection)

	// superEventDb := mongoClient.Database("super_event")
	// futureCollection := db.Collection("future_contract")
	// makeFutureIdAsIndexes(futureCollection)
	// futureDb := repository.NewFutureContractDB(superEventDb)

	digitalCertDb := mongoClient.Database("digital_certificate")
	// metadata
	apiRoute := newApp.Group("/api")
	// user
	userRouter := apiRoute.Group("/user")
	handler.NewUserHandler(userRouter, userDb)
	// authen
	authenRouter := apiRoute.Group("/authen")
	handler.NewAuthHandler(authenRouter, authDb)

	// digital cert type
	digitalCertTypeRouter := apiRoute.Group("cert-type")
	digitalCertTypeDb := repository.NewDigitalCertTypeDb(digitalCertDb)
	handler.NewDigitalCertTypeHandler(digitalCertTypeRouter, digitalCertTypeDb, uploader)

	// metadata for digital certificate
	digitalCertMetadataRouter := apiRoute.Group("/metadata")
	digitalCertMetadataDb := repository.NewMetadataRepository(digitalCertDb)
	handler.NewMetadataHandler(digitalCertMetadataRouter, digitalCertMetadataDb, digitalCertTypeDb, uploader)

	// redeemed
	redeemedRouter := apiRoute.Group("redeemed")
	redeemedDb := repository.NewRedeemedDb(digitalCertDb)
	handler.NewRedeemedHandler(redeemedRouter, redeemedDb)

	newApp.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hey you got me!!")
	})
	newApp.Static("/images", handler.StaticPath)

	// app.Listen(fmt.Sprintf(":%d", cfg.App.Port))
	app.Listen("")
}

// func initTimeZone() {
// 	ict, err := time.LoadLocation("Asia/Bangkok")
// 	if err != nil {
// 		panic(err)
// 	}

// 	time.Local = ict
// }

func connectMongo(username string, password string, url string) *mongo.Client {
	cfg := config.GetConfig()
	mongoUri := fmt.Sprintf("mongodb+srv://%s:%s@hdgcluster.xmgsx.mongodb.net", cfg.Mongo.Username, cfg.Mongo.Password)
	// mongoUri := "mongodb://HadesGod3:HadesGod@hdgcluster.xmgsx.mongodb.net"

	if cfg.Environment == "production" {
		mongoUri = fmt.Sprintf("mongodb://%s:%s@%s", username, password, url)
	}

	clientOptions := options.Client().ApplyURI(mongoUri)
	// ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s", username, password, url))
	// ApplyURI(fmt.Sprintf("mongodb+srv://HadesGod3:HadesGod@hdgcluster.xmgsx.mongodb.net"))
	// mongodb+srv://<username>:<password>@hdgcluster.xmgsx.mongodb.net/myFirstDatabase?retryWrites=true&w=majority
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Printf("connect to mongo server err: %v\n", err)
		log.Fatal(err)
	}
	return client
}

func makeUserEmailAsIndexes(col *mongo.Collection) {
	indexName, err := col.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("index name:", indexName)

	// indexName, err = col.Indexes().CreateOne(
	// 	context.Background(),
	// 	mongo.IndexModel{
	// 		Keys:    bson.D{{Key: "super_admin", Value: 1}},
	// 		Options: options.Index().SetUnique(true),
	// 	},
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("index name:", indexName)
}
