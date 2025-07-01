package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

func (s *Store) TunnelUpdateDeviceUID(ctx context.Context, tenantID, oldUID, newUID string) error {
	_, err := s.db.Collection("tunnels").UpdateMany(ctx, bson.M{"namespace": tenantID, "device": oldUID}, bson.M{"$set": bson.M{"device": newUID}})
	if err != nil {
		return FromMongoError(err)
	}

	return nil
}
