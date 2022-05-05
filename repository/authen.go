package repository

import (
	"context"
	"fmt"

	"bitbucket.org/atiwataqs/super-backend/authen"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SignInData struct {
	Token   *authen.TokenData `json:"token"`
	Refresh *authen.TokenData `json:"refresh"`
	User    *User             `json:"user"`
}

type AuthenticationRepository interface {
	Signin(email string, password string) (*SignInData, error)
	AdminSignin(email string, password string) (*SignInData, error)
	EventLoggerSigin(email string, password string) (*SignInData, error)
	NewAccessTokenAndRefreshToken(userId string, refreshToken string) (*SignInData, error)
	Signout(userId string) error
}

type AuthenDb struct {
	col *mongo.Collection
	ctx context.Context
}

func NewAuthDB(col *mongo.Collection) AuthenticationRepository {
	return &AuthenDb{
		col: col,
		ctx: context.Background(),
	}
}

func (a *AuthenDb) Signin(email string, password string) (*SignInData, error) {
	filer := genfilter("email", email)
	user := User{}
	res := a.col.FindOne(a.ctx, filer)
	err := res.Decode(&user)
	if err != nil {
		return nil, err
	}
	passwordOk := authen.VerifyPassword(user.Password, password)
	if !passwordOk {
		return nil, fmt.Errorf("password incorected")
	}
	token, err := authen.GenerateJWT(user.Id.Hex(), user.Email, user.MetamaskAddress, user.Role, 0)
	if err != nil {
		return nil, err
	}
	refresh, err := authen.GenerateRefreshJWT(user.Id.Hex(), user.Email, token.Token)
	if err != nil {
		return nil, err
	}

	// update refresh token in DB
	updateRefreshTokenInDB(user.Id, refresh.Token, a)
	signinData := SignInData{
		Token:   token,
		Refresh: refresh,
		User:    &user,
	}
	return &signinData, nil
}
func (a *AuthenDb) AdminSignin(email string, password string) (*SignInData, error) {
	// { $and: [{email: "super@super.energy2"}, {role: "admin"}] }
	filerFor := []bson.E{
		{Key: "email", Value: email},
		{Key: "role", Value: "admin"},
	}

	filter := bson.D{bson.E{Key: "$and", Value: filerFor}}
	user := User{}
	res := a.col.FindOne(a.ctx, filter)
	err := res.Decode(&user)
	if err != nil {
		return nil, err
	}
	token, err := authen.GenerateJWT(user.Id.Hex(), user.Email, user.MetamaskAddress, user.Role, 0)
	if err != nil {
		return nil, err
	}
	return &SignInData{
		Token: token,
		User:  &user,
	}, nil
}

func (a *AuthenDb) EventLoggerSigin(email string, password string) (*SignInData, error) {
	filerFor := []bson.E{
		{Key: "email", Value: email},
		{Key: "role", Value: "eventLogger"},
	}

	filter := bson.D{bson.E{Key: "$and", Value: filerFor}}
	user := User{}
	res := a.col.FindOne(a.ctx, filter)
	err := res.Decode(&user)
	if err != nil {
		return nil, err
	}
	token, err := authen.GenerateJWT(user.Id.Hex(), user.Email, user.MetamaskAddress, user.Role, 0)
	if err != nil {
		return nil, err
	}
	return &SignInData{
		Token: token,
		User:  &user,
	}, nil
}

func (a *AuthenDb) NewAccessTokenAndRefreshToken(userId string, refreshToken string) (*SignInData, error) {
	id, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, err
	}
	filter := genfilter("_id", id)
	user := User{}
	res := a.col.FindOne(a.ctx, filter)
	err = res.Decode(&user)
	if err != nil {
		return nil, err
	}
	if user.Refreshtoken != refreshToken {
		return nil, fmt.Errorf("refresh token is no longer valid")
	}
	token, err := authen.GenerateJWT(user.Id.Hex(), user.Email, user.MetamaskAddress, user.Role, 0)
	if err != nil {
		return nil, err
	}
	refresh, err := authen.GenerateRefreshJWT(user.Id.Hex(), user.Email, token.Token)
	if err != nil {
		return nil, err
	}

	// update refresh token in DB
	updateRefreshTokenInDB(user.Id, refresh.Token, a)
	signinData := SignInData{
		Token:   token,
		Refresh: refresh,
		User:    &user,
	}
	return &signinData, nil
}

func (a *AuthenDb) Signout(userId string) error {
	id, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return err
	}
	res := a.col.FindOne(a.ctx, genfilter("_id", id))
	user := User{}
	err = res.Decode(&user)
	if err != nil {
		return err
	}
	user.Refreshtoken = ""
	update := bson.D{bson.E{Key: "$set", Value: user}}
	updateRes, err := a.col.UpdateByID(a.ctx, id, update)
	if err != nil {
		return err
	}
	if updateRes.ModifiedCount <= 0 {
		return fmt.Errorf("signout -> update user error")
	}
	return nil
}

// helper
func updateRefreshTokenInDB(userId primitive.ObjectID, refreshToken string, authDb *AuthenDb) error {
	filter := genfilter("_id", userId)
	update := bson.D{bson.E{Key: "$set", Value: bson.M{"refresh_token": refreshToken}}}
	res := authDb.col.FindOneAndUpdate(authDb.ctx, filter, update)
	if res.Err() != nil {
		fmt.Printf("res: %v\n", res.Err().Error())
	}
	return res.Err()
}

func genfilter(key string, value interface{}) primitive.D {
	return bson.D{bson.E{Key: key, Value: value}}
}
