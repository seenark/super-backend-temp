package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Metadata struct {
	TypeCode      string `bson:"type_code" json:"type_code"`
	DigitalCertID int    `bson:"digital_cert_id" json:"digital_cert_id"`
	ProjectName   string `bson:"project_name" json:"project_name"`
	ProjectType   string `bson:"project_type" json:"project_type"`
	ImageName     string `bson:"image_name" json:"image_name"`
	Description   string `bson:"description" json:"description"`
	ListedDate    string `bson:"listed_date" json:"listed_date"`
}

type IMetadataRepository interface {
	GetAll(ids []int) []Metadata
	GetByDigitalCertId(certId int) (*Metadata, error)
	GetByTypeCode(typeCode string) []Metadata
	Create(typeCode string, certId int, projectName string, projectType string, imageName string, description string, listedDate string) (*Metadata, error)
	UpdateByDigitalCertId(certId int, metadata Metadata) (*Metadata, error)
	DeleteByDigitalCertId(certId int) (*Metadata, error)
}

type MetadataDb struct {
	col *mongo.Collection
	ctx context.Context
}

const (
	METADATA_COLLECTION_NAME = "metadata"
)

// Create implements IMetadataRepository
func (m MetadataDb) Create(typeCode string, certId int, projectName string, projectType string, imageName string, description string, listedDate string) (*Metadata, error) {

	metadata := Metadata{
		TypeCode:      typeCode,
		DigitalCertID: certId,
		ProjectName:   projectName,
		ProjectType:   projectType,
		ImageName:     imageName,
		Description:   description,
		ListedDate:    listedDate,
	}
	_, err := m.col.InsertOne(m.ctx, metadata)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

// DeleteByDigitalCertId implements IMetadataRepository
func (m MetadataDb) DeleteByDigitalCertId(certId int) (*Metadata, error) {
	filter := genfilter("digital_cert_id", certId)
	res := m.col.FindOne(m.ctx, filter)
	metadata := Metadata{}
	err := res.Decode(&metadata)
	if err != nil {
		return nil, err
	}
	delRes, err := m.col.DeleteOne(m.ctx, filter)
	if err != nil {
		return nil, err
	}
	if delRes.DeletedCount <= 0 {
		return nil, fmt.Errorf("delete count == 0")
	}
	return &metadata, nil
}

// GetAll implements IMetadataRepository
func (m MetadataDb) GetAll(ids []int) []Metadata {
	metadatas := []Metadata{}
	filter := bson.M{}
	if len(ids) > 0 {
		filter = bson.M{
			"digital_cert_id": bson.M{
				"$in": ids,
			},
		}
	}

	cur, err := m.col.Find(m.ctx, filter)
	if err != nil {
		return metadatas
	}
	for cur.Next(m.ctx) {
		metadata := Metadata{}
		err = cur.Decode(&metadata)
		if err != nil {
			continue
		}
		metadatas = append(metadatas, metadata)
	}
	return metadatas
}

// GetByDigitalCertId implements IMetadataRepository
func (m MetadataDb) GetByDigitalCertId(certId int) (*Metadata, error) {
	filter := genfilter("digital_cert_id", certId)
	res := m.col.FindOne(m.ctx, filter)
	medadata := Metadata{}
	err := res.Decode(&medadata)
	if err != nil {
		return nil, err
	}
	return &medadata, nil
}

// GetByTypeId implements IMetadataRepository
func (m MetadataDb) GetByTypeCode(typeCode string) []Metadata {
	filter := genfilter("type_code", typeCode)
	metadatas := []Metadata{}
	cur, err := m.col.Find(m.ctx, filter)
	if err != nil {
		return metadatas
	}
	for cur.Next(m.ctx) {
		metadata := Metadata{}
		err := cur.Decode(&metadata)
		if err != nil {
			continue
		}
		metadatas = append(metadatas, metadata)
	}
	return metadatas
}

// UpdateByDigitalCertId implements IMetadataRepository
func (m MetadataDb) UpdateByDigitalCertId(certId int, metadata Metadata) (*Metadata, error) {
	filter := genfilter("digital_cert_id", certId)
	update := bson.D{primitive.E{Key: "$set", Value: metadata}}

	updateRes, err := m.col.UpdateOne(m.ctx, filter, update)
	if err != nil {
		return nil, err
	}
	if updateRes.ModifiedCount <= 0 {
		return nil, fmt.Errorf("no update any metadata")
	}
	return &metadata, nil
}

func NewMetadataRepository(db *mongo.Database) IMetadataRepository {
	col := db.Collection(METADATA_COLLECTION_NAME)
	makeDigitalCertIDAsIndexes(col)
	return MetadataDb{
		col: col,
		ctx: context.Background(),
	}
}

func makeDigitalCertIDAsIndexes(collection *mongo.Collection) {
	indexName, err := collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "digital_cert_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("index name:", indexName)
}
