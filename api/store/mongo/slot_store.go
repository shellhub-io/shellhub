package mongo

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Store) SlotSet(ctx context.Context, tenant string, uid models.UID, status string) error {
	_, err := s.db.Collection("slots").UpdateOne(ctx, bson.M{"tenant_id": tenant, "uid": uid}, bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}}, options.Update().SetUpsert(true))
	if err != nil {
		return FromMongoError(err)
	}

	return nil
}

// SlotDelete
func (s *Store) SlotDelete(ctx context.Context, tenant string, uid models.UID) error {
	_, err := s.db.Collection("slots").DeleteOne(ctx, bson.M{"tenant_id": tenant, "uid": uid})
	if err != nil {
		return FromMongoError(err)
	}

	return nil
}

func (s *Store) SlotsList(ctx context.Context, tenant string) ([]models.Slot, error) {
	slots, err := s.db.Collection("slots").Find(ctx, bson.M{"tenant_id": tenant})
	if err != nil {
		return nil, FromMongoError(err)
	}

	var slotsList []models.Slot
	for slots.Next(ctx) {
		var slot models.Slot
		if err := slots.Decode(&slot); err != nil {
			return nil, FromMongoError(err)
		}

		slotsList = append(slotsList, slot)
	}

	return slotsList, nil
}
