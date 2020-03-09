package mongo

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Client contains the mongo Client to maintain connections
type Client struct {
	*mongo.Client // Client is a handle representing a pool of connections to a MongoDB deployment.
}

// NewClient creates a new client that connects to a cluster specified by the uri.
func NewClient(username, password, host string, port int) Client {
	uri := generateURI(username, password, host, port)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Println(err)
	}
	return Client{
		client,
	}
}

func generateURI(username, password, host string, port int) string {
	if (username != "") && (password != "") {
		password += "@"
		return fmt.Sprintf("mongodb://%s:%s%s:%d", username, password, host, port)
	}
	return fmt.Sprintf("mongodb://%s:%d", host, port)
}

// Connection initializes the new client
func (c Client) Connection() {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := c.Connect(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	ctxt, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()
	err = c.Ping(ctxt, readpref.Primary())
	if err != nil {
		log.Println(err)
		return
	}
}
