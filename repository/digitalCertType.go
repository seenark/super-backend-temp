package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DigitalCertType struct {
	TypeCode      string `bson:"type_code" json:"type_code"`
	TypeName      string `bson:"type_name" json:"type_name"`
	LogoImageName string `bson:"logo_image_name" json:"logo_image_name"`
	TypeOfUnit    string `bson:"type_of_unit" json:"type_of_unit"`
	Unit          string `bson:"unit" json:"unit"`
	VintageYear   string `bson:"vintage_year" json:"vintage_year"`
}

type IDigitalCertTypeRepository interface {
	GetAll() []DigitalCertType
	GetByTypeCode(string) (*DigitalCertType, error)
	Create(typeCode string, typeName string, logoImageName string, typeOfUnit string, unit string, vintageYear string) (*DigitalCertType, error)
	Update(string, DigitalCertType) (*DigitalCertType, error)
	Delete(string) (*DigitalCertType, error)
	FindInArrayByTypeCode(string, []DigitalCertType) (int, *DigitalCertType)
}

type DigitalCertTypeDb struct {
	col *mongo.Collection
	ctx context.Context
}

const (
	DIGITAL_CERT_TYPE_COLLECTION_NAME = "digital_cert_type"
)

// Create implements IDigitalCertTypeRepository
func (d DigitalCertTypeDb) Create(typeCode string, typeName string, logoImageName string, typeOfUnit string, unit string, vintageYear string) (*DigitalCertType, error) {
	certType := DigitalCertType{
		TypeCode:      typeCode,
		TypeName:      typeName,
		LogoImageName: logoImageName,
		TypeOfUnit:    typeOfUnit,
		Unit:          unit,
		VintageYear:   vintageYear,
	}
	_, err := d.col.InsertOne(d.ctx, certType)
	if err != nil {
		return nil, err
	}
	return &certType, nil
}

// Delte implements IDigitalCertTypeRepository
func (d DigitalCertTypeDb) Delete(typeCode string) (*DigitalCertType, error) {
	filter := genfilter("type_code", typeCode)

	res := d.col.FindOneAndDelete(d.ctx, filter)
	certType := DigitalCertType{}
	err := res.Decode(&certType)
	if err != nil {
		return nil, err
	}
	return &certType, nil
}

// GetAll implements IDigitalCertTypeRepository
func (d DigitalCertTypeDb) GetAll() []DigitalCertType {
	certTypes := []DigitalCertType{}
	cur, err := d.col.Find(d.ctx, bson.M{})
	if err != nil {
		return certTypes
	}
	for cur.Next(d.ctx) {
		certType := DigitalCertType{}
		err := cur.Decode(&certType)
		if err != nil {
			fmt.Println(err)
			continue
		}
		certTypes = append(certTypes, certType)
	}
	return certTypes
}

// GetByTypeCode implements IDigitalCertTypeRepository
func (d DigitalCertTypeDb) GetByTypeCode(typeCode string) (*DigitalCertType, error) {
	filter := genfilter("type_code", typeCode)

	res := d.col.FindOne(d.ctx, filter)
	certType := DigitalCertType{}
	err := res.Decode(&certType)
	if err != nil {
		return nil, err
	}
	return &certType, nil
}

// Update implements IDigitalCertTypeRepository
func (d DigitalCertTypeDb) Update(typeCode string, newCertType DigitalCertType) (*DigitalCertType, error) {
	filter := genfilter("type_code", typeCode)

	updateObj := bson.D{primitive.E{Key: "$set", Value: newCertType}}

	upRes, err := d.col.UpdateOne(d.ctx, filter, updateObj)
	if err != nil {
		return nil, err
	}
	if upRes.ModifiedCount == 0 {
		return nil, fmt.Errorf("no update any")
	}
	return &newCertType, nil
}

func (d DigitalCertTypeDb) FindInArrayByTypeCode(typeCode string, certTypes []DigitalCertType) (int, *DigitalCertType) {
	for i, v := range certTypes {
		if v.TypeCode == typeCode {
			return i, &v
		}
	}
	return 0, &DigitalCertType{}
}

func NewDigitalCertTypeDb(db *mongo.Database) IDigitalCertTypeRepository {
	col := db.Collection(DIGITAL_CERT_TYPE_COLLECTION_NAME)
	makeTypeCodeAsIndexes(col)
	return DigitalCertTypeDb{
		col: col,
		ctx: context.Background(),
	}
}

func makeTypeCodeAsIndexes(collection *mongo.Collection) {
	indexName, err := collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "type_code", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("index name:", indexName)
}
