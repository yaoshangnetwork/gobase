package mongo

import "go.mongodb.org/mongo-driver/bson/primitive"

type BaseModel struct {
	ID        primitive.ObjectID `json:"id"        bson:"_id,omitempty"`
	CreatedAt int64              `json:"created_at" bson:"created_at"`
	UpdatedAt int64              `json:"updated_at" bson:"updated_at"`
	DeletedAt *int64             `json:"-"         bson:"deleted_at,omitempty"`
	IsDeleted bool               `json:"-"         bson:"is_deleted"`
}
