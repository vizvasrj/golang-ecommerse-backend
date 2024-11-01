package misc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"src/pkg/conf"
	"src/pkg/module/user"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func createMerchantUser(app *conf.Config, ctx context.Context, email, name string, merchantID primitive.ObjectID, host string) (*mongo.UpdateResult, error) {
	firstName := name
	lastName := ""
	var existingUser user.User
	err := app.UserCollection.FindOne(ctx, bson.M{"email": email}).Decode(&existingUser)
	if err == nil {
		query := bson.M{"_id": existingUser.ID}
		update := bson.M{
			"$set": bson.M{
				"merchant": merchantID,
				"role":     ROLE_MERCHANT,
			},
		}

		merchantCollection := app.DB.Collection("merchants")
		var merchantDoc Merchant
		err = merchantCollection.FindOne(ctx, bson.M{"email": email}).Decode(&merchantDoc)
		if err != nil {
			return nil, err
		}

		_, err = createMerchantBrand(ctx, merchantID, merchantDoc.Email, merchantDoc.Email)
		if err != nil {
			return nil, err
		}

		err = sendEmail(email, "merchant-welcome", nil, name)
		if err != nil {
			return nil, err
		}

		return userCollection.UpdateOne(ctx, query, update)
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
		Role:               ROLE_MERCHANT,
	}
	// todo send email
	// err = sendEmail(email, "merchant-signup", host, map[string]string{
	// 	"resetToken": resetToken,
	// 	"email":      email,
	// })
	// if err != nil {
	// 	return nil, err
	// }

	_, err = userCollection.InsertOne(ctx, user)
	return nil, err
}
