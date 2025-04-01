package database

import (
	"context"
	"log"
	
	"time"

	"github.com/ChanchalS7/go-graphql-mongodb/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var connectionString = "mongodb+srv://chanchal:12345@cluster0.dpggmv8.mongodb.net/go-graphql-mongodb"

type DB struct {
	client *mongo.Client
}

func Connect() *DB {

	ctx, cancel := context.WithTimeout(context.Background(),10*time.Second)
	defer cancel()

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts:= options.Client().ApplyURI(connectionString).SetServerAPIOptions(serverAPI)


	client, err := mongo.Connect(ctx,opts)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB %v",err)
	}
	

	//check connection by pinging mongodb
	err = client.Ping(ctx,readpref.Primary())
	if err != nil {
		log.Fatalf("Could not ping MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB successfully")

	return &DB{
		client: client,
	}
}

func (db *DB) Close(){
	if err := db.client.Disconnect(context.Background()); err!=nil{
		log.Fatalf("Error while disconnecting from MongoDB: %v",err)
	}
	log.Println("Disconnected from MongoDB")
}


func (db *DB) GetJob(id string) *model.JobListing {
	jobCollection:=db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_id,_:= primitive.ObjectIDFromHex(id)
	filter:=bson.M{"_id": _id}
		var jobListing model.JobListing
	err:= jobCollection.FindOne(ctx, filter).Decode(&jobListing)
	if err!=nil{
		log.Fatal(err)
	}

	return &jobListing
}

func (db *DB) GetJobs() []*model.JobListing{
	jobCollection:= db.client.Database("graphql-job-board").Collection("jobs")	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var jobListings []*model.JobListing
		cursor, err:= jobCollection.Find(ctx, bson.D{})
		if err!= nil {
			log.Fatal(err)
		}
		if err = cursor.All(context.TODO(), &jobListings); err!=nil{
			panic(err)
		}
	return jobListings
}
func(db *DB) CreateJobListing(jobInfo model.CreateJobListingInput) *model.JobListing {
	jobCollection:=db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(),30*time.Second)
	defer cancel()

	inserted, err := jobCollection.InsertOne(ctx,bson.M{
		"title":jobInfo.Title,
		"description":jobInfo.Description,
		"url":jobInfo.URL,
		"company":jobInfo.Company,

	})
	insertedID := inserted.InsertedID.(primitive.ObjectID).Hex()
	returnJobListing := model.JobListing{
	ID:insertedID,
	Title : jobInfo.Title,
	Company:  jobInfo.Company,
	Description: jobInfo.Description,
	URL : jobInfo.URL,
	}
	if err !=nil{
		log.Fatal(err)
	}


	return &returnJobListing
}

func (db *DB) UpdateJobListing(jobId string, jobInfo model.UpdateJobListingInput) *model.JobListing{ 
	jobCollection := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(),30*time.Second)
	defer cancel()

	updateJobInfo := bson.M{}
	if jobInfo.Title != nil {
		updateJobInfo["title"]=jobInfo.Title
	}
	if jobInfo.Description!=nil{
		updateJobInfo["description"]=jobInfo.Description
	} 
	// if jobInfo.Company!=nil{
	// 	updateJobInfo["company"]=jobInfo.Company
	// }
	if jobInfo.URL!=nil{
		updateJobInfo["URL"]=jobInfo.URL
	}
	_id, _ := primitive.ObjectIDFromHex(jobId)
	filter := bson.M{"_id": _id}
	update := bson.M{"$set":updateJobInfo}
 	results := jobCollection.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(1))
	var jobListing model.JobListing
	if err := results.Decode(&jobListing); err!=nil{
		log.Fatal(err)
	}
	return &jobListing
}

func(db *DB) DeleteJobListing(jobId string) *model.DeleteJobResponse{
	jobCollection := db.client.Database("grapql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_id,_ := primitive.ObjectIDFromHex(jobId)

	filter := bson.M{"_id":_id}
	_, err := jobCollection.DeleteOne(ctx, filter)
	if err!=nil{
		log.Fatal(err)
	}

return &model.DeleteJobResponse{DeletedJobID: jobId}
}