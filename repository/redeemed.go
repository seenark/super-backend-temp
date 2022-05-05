package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	REQUEST_STATUS         = "requested"
	APPROVED_STATUS        = "approved"
	PENDING_STATUS         = "pending"
	DELIVERED_STATUS       = "delivered"
	REDEEM_COLLECTION_NAME = "redeemeds"
)

type Redeemed struct {
	TxHash        string `bson:"tx_hash" json:"tx_hash"` // indexed
	Name          string `bson:"name" json:"name"`
	Company       string `bson:"company" json:"company"`
	Email         string `bson:"email" json:"email"`
	Telephone     string `bson:"telephone" json:"telephone"`
	TaxID         string `bson:"tax_id" json:"tax_id"`
	Price         string `bson:"price" json:"price"` // BigNumber in string format
	RedeemId      int    `bson:"redeem_id" json:"redeem_id"`
	RedeemDate    int    `bson:"redeem_date" json:"redeem_date"` // Unix timestamp
	WalletAddress string `bson:"wallet_address" json:"wallet_address"`
	ApproveStatus string `bson:"approved_status" json:"approved_status"`
	Amount        int    `bson:"amount" json:"amount"`
	CertId        int    `bson:"cert_id" json:"cert_id"`
}

type IRedeemedRepository interface {
	create(Redeemed) (*Redeemed, error)
	update(Redeemed) (*Redeemed, error)
	Upsert(Redeemed) (*Redeemed, error)
	GetAll(approveStatus []string, name string, email string, redeemId []int, startDate int, endDate int, walletAddress string) []Redeemed
	GetByTxHash(txHash string) (*Redeemed, error)
	GetByRedeemId(redeemId int) (*Redeemed, error)
	UpdateStatus(redeemId int, status string) (*Redeemed, error)
}

type RedeemedDb struct {
	col *mongo.Collection
	ctx context.Context
}

// create implements IRedeemedRepository
func (r RedeemedDb) create(redeemed Redeemed) (*Redeemed, error) {
	_, err := r.col.InsertOne(r.ctx, redeemed)
	if err != nil {
		return nil, err
	}
	return &redeemed, nil
}

// update implements IRedeemedRepository
func (r RedeemedDb) update(redeemed Redeemed) (*Redeemed, error) {
	filter := genfilter("tx_hash", redeemed.TxHash)
	update := bson.M{"$set": redeemed}

	updateRes, err := r.col.UpdateOne(r.ctx, filter, update)
	if err != nil {
		return nil, err
	}
	if updateRes.ModifiedCount <= 0 {
		return nil, fmt.Errorf("no update any item")
	}
	return &redeemed, nil
}

// GetAll implements IRedeemedRepository
func (r RedeemedDb) GetAll(approveStatus []string, name string, email string, redeemId []int, startDate int, endDate int, walletAddress string) []Redeemed {
	filter := bson.M{}
	if len(approveStatus) > 0 {
		filter["approved_status"] = bson.M{"$in": approveStatus}
	}

	if name != "" {
		filter["name"] = name
	}

	if email != "" {
		filter["email"] = email
	}

	if len(redeemId) > 0 {
		filter["redeem_id"] = bson.M{"$in": redeemId}
	}

	if startDate > 0 && endDate > 0 {
		filter["$and"] = bson.A{
			bson.M{"redeem_date": bson.M{"$gte": startDate}},
			bson.M{"redeem_date": bson.M{"$lt": endDate}},
		}
	} else if startDate > 0 && endDate == 0 {
		filter["redeem_date"] = bson.M{"$gte": startDate}
	} else if endDate > 0 && startDate == 0 {
		filter["redeem_date"] = bson.M{"$lte": endDate}
	}

	if walletAddress != "" {
		filter["wallet_address"] = walletAddress
	}

	fmt.Printf("filter: %v\n", filter)
	allRedeemed := []Redeemed{}
	cur, err := r.col.Find(r.ctx, filter)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return allRedeemed
	}

	for cur.Next(r.ctx) {
		redeemed := Redeemed{}
		err := cur.Decode(&redeemed)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			continue
		}
		allRedeemed = append(allRedeemed, redeemed)
	}

	return allRedeemed
}

// GetByTxHash implements IRedeemedRepository
func (r RedeemedDb) GetByTxHash(txHash string) (*Redeemed, error) {
	filter := genfilter("tx_hash", txHash)
	res := r.col.FindOne(r.ctx, filter)
	redeemed := Redeemed{}
	err := res.Decode(&redeemed)
	if err != nil {
		return nil, err
	}
	return &redeemed, nil
}

func (r RedeemedDb) GetByRedeemId(redeemId int) (*Redeemed, error) {
	filter := genfilter("redeem_id", redeemId)
	res := r.col.FindOne(r.ctx, filter)
	redeemed := Redeemed{}
	err := res.Decode(&redeemed)
	if err != nil {
		return nil, err
	}
	return &redeemed, nil
}

// Upsert implements IRedeemedRepository
func (r RedeemedDb) Upsert(redeemed Redeemed) (*Redeemed, error) {
	findRedeemed, err := r.GetByTxHash(redeemed.TxHash)
	if err != nil {
		// not found
		// should create new redeemed
		newRedeemed, err := r.create(redeemed)
		if err != nil {
			return nil, err
		}
		return newRedeemed, nil
	} else {
		// found old one
		// update data
		if redeemed.Amount > 0 {
			findRedeemed.Amount = redeemed.Amount
		}
		if redeemed.ApproveStatus != findRedeemed.ApproveStatus {
			findRedeemed.ApproveStatus = redeemed.ApproveStatus
		}
		if redeemed.Company != "" {
			findRedeemed.Company = redeemed.Company
		}
		if redeemed.Email != "" {
			findRedeemed.Email = redeemed.Email
		}
		if redeemed.Name != "" {
			findRedeemed.Name = redeemed.Name
		}
		if redeemed.Price != "" {
			findRedeemed.Price = redeemed.Price
		}
		if redeemed.RedeemDate != 0 {
			findRedeemed.RedeemDate = redeemed.RedeemDate
		}
		if redeemed.RedeemId > 0 {
			findRedeemed.RedeemId = redeemed.RedeemId
		}
		if redeemed.TaxID != "" {
			findRedeemed.TaxID = redeemed.TaxID
		}
		if redeemed.Telephone != "" {
			findRedeemed.Telephone = redeemed.Telephone
		}
		if redeemed.WalletAddress != "" {
			findRedeemed.WalletAddress = redeemed.WalletAddress
		}
		if redeemed.CertId != 0 {
			findRedeemed.CertId = redeemed.CertId
		}

		newRedeemed, err := r.update(*findRedeemed)
		if err != nil {
			return nil, err
		}
		return newRedeemed, nil
	}
}

// UpdateStatus implements IRedeemedRepository
func (r RedeemedDb) UpdateStatus(redeemId int, status string) (*Redeemed, error) {
	filter := genfilter("redeem_id", redeemId)
	findRedeem, err := r.GetByRedeemId(redeemId)
	if err != nil {
		return nil, err
	}
	findRedeem.ApproveStatus = status
	update := bson.M{"$set": findRedeem}
	updateRes, err := r.col.UpdateOne(r.ctx, filter, update)
	if err != nil {
		return nil, err
	}
	if updateRes.ModifiedCount <= 0 {
		return nil, fmt.Errorf("no update any items")
	}
	return findRedeem, nil
}

func NewRedeemedDb(db *mongo.Database) IRedeemedRepository {
	col := db.Collection(REDEEM_COLLECTION_NAME)
	makeTransactionHashAsIndexes(col)
	return RedeemedDb{
		col: col,
		ctx: context.Background(),
	}
}

func makeTransactionHashAsIndexes(collection *mongo.Collection) {
	indexName, err := collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "tx_hash", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("index name:", indexName)
}
