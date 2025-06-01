package db

// "go.mongodb.org/mongo-driver/mongo"
// "go.mongodb.org/mongo-driver/mongo/options"

// // InitializeMongoDB initializes the MongoDB client
// func InitializeMongoDB(uri string) (*mongo.Client, error) {
// 	ctx := context.Background()
// 	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

// 	clientOptions := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
// 	client, err := mongo.Connect(ctx, clientOptions)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Check the connection
// 	err = client.Ping(ctx, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	log.Println("Connected to MongoDB!")
// 	return client, nil
// }

// // GetCollection returns a MongoDB collection
// func GetCollection(client *mongo.Client, database, collection string) *mongo.Collection {
// 	return client.Database(database).Collection(collection)
// }
