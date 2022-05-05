package repository

import (
	"context"
	"fmt"

	"bitbucket.org/atiwataqs/super-backend/authen"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	Id              primitive.ObjectID `bson:"_id" json:"id"`
	Name            string             `bson:"name" json:"name"`
	Email           string             `bson:"email" json:"email"`
	Role            string             `bson:"role" json:"role"`
	MetamaskAddress string             `bson:"metamask_address" json:"metamask_address"`
	Address         string             `bson:"address" json:"address"`
	Tel             string             `bson:"tel" json:"tel"`
	Password        string             `bson:"password" json:"password"`
	SuperAdmin      bool               `bson:"super_admin,omitempty" json:"super_admin"`
	Refreshtoken    string             `bson:"refresh_token" json:"refresh_token"`
}

type UserRepository interface {
	GetAll() (users []User)
	GetById(userId string) (user *User, err error)
	Create(newUser User) (user *User, err error)
	Update(newUser User) (user *User, err error)
	Delete(userId string) (user *User, err error)
}

type UserDb struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewUserDb(collection *mongo.Collection) UserRepository {
	return &UserDb{
		collection: collection,
		ctx:        context.Background(),
	}
}

func (u *UserDb) GetAll() (users []User) {
	cur, _ := u.collection.Find(u.ctx, bson.M{})
	for cur.Next(u.ctx) {
		user := User{}
		err := cur.Decode(&user)
		if err != nil {
			continue
		}
		users = append(users, user)
	}
	return users
}
func (u *UserDb) GetById(userId string) (user *User, err error) {
	userIdObjectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, err
	}
	filter := bson.D{bson.E{Key: "_id", Value: userIdObjectId}}
	res := u.collection.FindOne(u.ctx, filter)
	err = res.Decode(&user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func (u *UserDb) Create(newUser User) (user *User, err error) {
	user = &newUser
	user.Id = primitive.NewObjectID()
	user.Password, err = authen.HashPassword(newUser.Password)
	if err != nil {
		return nil, err
	}
	_, err = u.collection.InsertOne(u.ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserDb) Update(newUser User) (user *User, err error) {
	user, err = u.GetById(newUser.Id.Hex())
	if err != nil {
		return nil, err
	}
	if newUser.Name != "" {
		user.Name = newUser.Name
	}
	if newUser.Email != "" {
		user.Email = newUser.Email
	}
	if newUser.MetamaskAddress != "" {
		user.MetamaskAddress = newUser.MetamaskAddress
	}
	if newUser.Role != "" {
		user.Role = newUser.Role
	}
	if newUser.Tel != "" {
		user.Tel = newUser.Tel
	}
	if newUser.Address != "" {
		user.Address = newUser.Address
	}
	if newUser.Password != "" {
		newPass, err := authen.HashPassword(newUser.Password)
		if err != nil {
			return nil, err
		}
		user.Password = newPass
	}
	user.Refreshtoken = ""
	filter := genfilter("_id", newUser.Id)
	update := bson.D{bson.E{Key: "$set", Value: user}}
	res, err := u.collection.UpdateOne(u.ctx, filter, update)
	if err != nil {
		return nil, err
	}
	if res.ModifiedCount == 0 {
		return nil, fmt.Errorf("no item updated")
	}
	return user, nil
}
func (u *UserDb) Delete(userId string) (user *User, err error) {
	user, err = u.GetById(userId)
	if err != nil {
		return nil, err
	}
	filter := bson.D{bson.E{Key: "_id", Value: user.Id}}
	delete, err := u.collection.DeleteOne(u.ctx, filter)
	if err != nil {
		return nil, err
	}
	if delete.DeletedCount == 0 {
		return nil, fmt.Errorf("no item deleted")
	}
	return user, nil
}
