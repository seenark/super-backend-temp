package repository

// import (
// 	"context"
// 	"fmt"

// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// 	"go.mongodb.org/mongo-driver/mongo"
// )

// type RedeemEvent struct {
// 	Id             primitive.ObjectID  `bson:"_id" json:"id"`
// 	RedeemId       int                 `bson:"redeemId" json:"redeemId"`
// 	Customer       string              `bson:"customer" json:"customer"`
// 	FutureTokenId  int                 `bson:"futureTokenId" json:"futureTokenId"`
// 	Amount         int                 `bson:"amount" json:"amount"`
// }

// type RedeemEventDb struct {
// 	cl  *mongo.Collection
// 	ctx context.Context
// }

// type IRedeemEventRepository interface {
// 	GetRedeemByFutureContractId(id int) []RedeemEvent
// 	GetRedeemByCustomerAddress(address string) []RedeemEvent
// }

// // GetRedeemByFutureContractId implements IFutureContractEventRepository
// func (fe RedeemEventDb) GetRedeemByFutureContractId(id int) []RedeemEvent {
// 	pipeline := []bson.D{
// 		{
// 			{Key: "$match", Value: bson.M{"futureTokenId": id}},
// 		},
// 		{
// 			{Key: "$lookup", Value: bson.D{
// 				{Key: "from", Value: "futurecontracts"},
// 				{Key: "localField", Value: "futureToken"},
// 				{Key: "foreignField", Value: "_id"},
// 				{Key: "as", Value: "futureContract"},
// 			}},
// 		},
// 		{
// 			{Key: "$lookup", Value: bson.D{
// 				{Key: "from", Value: "pricefeeds"},
// 				{Key: "localField", Value: "futureContract.keyForPrice"},
// 				{Key: "foreignField", Value: "_id"},
// 				{Key: "as", Value: "pricefeed"},
// 			}},
// 		},
// 		{
// 			{Key: "$project", Value: bson.D{
// 				{Key: "redeemId", Value: "$redeemId"},
// 				{Key: "customer", Value: "$customer"},
// 				{Key: "futureTokenId", Value: "$futureTokenId"},
// 				{Key: "amount", Value: "$amount"},
// 				{Key: "futureContract", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$futureContract", 0}}}},
// 				{Key: "pricefeed", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$pricefeed", 0}}}},
// 			},
// 			},
// 		},
// 	}
// 	redeems := []RedeemEvent{}
// 	cur, err := fe.cl.Aggregate(fe.ctx, mongo.Pipeline(pipeline))
// 	if err != nil {
// 		fmt.Println(err)
// 		return redeems
// 	}
// 	for cur.Next(fe.ctx) {
// 		redeem := RedeemEvent{}
// 		err := cur.Decode(&redeem)
// 		if err != nil {
// 			fmt.Println(err)
// 		}
// 		redeems = append(redeems, redeem)
// 	}

// 	return redeems
// }

// // GetRedeemByCustomerAddress implements IRedeemEventRepository
// func (fe RedeemEventDb) GetRedeemByCustomerAddress(address string) []RedeemEvent {
// 	filter := genfilter("customer", bson.D{{Key: "$regex", Value: primitive.Regex{Pattern: address, Options: "i"}}})
// 	redeems := []RedeemEvent{}

// 	cur, err := fe.cl.Find(fe.ctx, filter)
// 	if err != nil {
// 		fmt.Println(err)
// 		return redeems
// 	}
// 	for cur.Next(fe.ctx) {
// 		redeem := RedeemEvent{}
// 		err := cur.Decode(&redeem)
// 		if err != nil {
// 			fmt.Println(err)
// 		}
// 		redeems = append(redeems, redeem)
// 	}

// 	return redeems
// }

// func NewRedeemEvent(db *mongo.Database) IRedeemEventRepository {
// 	cl := db.Collection("redeems")
// 	return RedeemEventDb{
// 		cl:  cl,
// 		ctx: context.Background(),
// 	}
// }
