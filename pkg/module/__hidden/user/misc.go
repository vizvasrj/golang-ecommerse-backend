package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"src/common"
	"src/pkg/conf"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateMerchantUser(app *conf.Config, ctx context.Context, email, name string, merchantID primitive.ObjectID) (*mongo.UpdateResult, error) {
	firstName := name
	lastName := ""
	var existingUser User
	err := app.UserCollection.FindOne(ctx, bson.M{"email": email}).Decode(&existingUser)
	if err == nil {
		query := bson.M{"_id": existingUser.ID}
		update := bson.M{
			"$set": bson.M{
				"merchant": merchantID,
				"role":     common.RoleMerchant,
			},
		}

		var merchantDoc Merchant
		err = app.MerchantCollection.FindOne(ctx, bson.M{"email": email}).Decode(&merchantDoc)
		if err != nil {
			return nil, err
		}

		// err = sendEmail(email, "merchant-welcome", nil, name)
		// if err != nil {
		// 	return nil, err
		// }

		return app.UserCollection.UpdateOne(ctx, query, update)
	}

	buffer := make([]byte, 48)
	_, err = rand.Read(buffer)
	if err != nil {
		return nil, err
	}
	resetToken := hex.EncodeToString(buffer)
	resetPasswordToken := resetToken

	user := User{
		Email:              email,
		FirstName:          firstName,
		LastName:           lastName,
		ResetPasswordToken: resetPasswordToken,
		Merchant:           merchantID,
		Role:               common.RoleMerchant,
	}
	// todo send email
	// err = sendEmail(email, "merchant-signup", host, map[string]string{
	// 	"resetToken": resetToken,
	// 	"email":      email,
	// })
	// if err != nil {
	// 	return nil, err
	// }

	_, err = app.UserCollection.InsertOne(ctx, user)
	return nil, err
}
