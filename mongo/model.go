package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BaseModel `bson:",inline"`

type BaseModel struct {
	ID        primitive.ObjectID `json:"id"         bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	DeletedAt *time.Time         `json:"-"          bson:"deleted_at,omitempty"`
	IsDeleted bool               `json:"-"          bson:"is_deleted"`
}

func (b *BaseModel) BeforeCreate() {
	b.CreatedAt = time.Now()
	b.UpdatedAt = b.CreatedAt
}
