package mongo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var timeout = 5 * time.Minute

type ID = primitive.ObjectID

// NewID generates a new ID.
func NewID() (id ID) {
	return ID(primitive.NewObjectID())
}

// SetID creates an ID object from string.
func SetID(s string) (id ID, err error) {
	id, err = primitive.ObjectIDFromHex(s)
	if err != nil {
		log.Println(err)
		return
	}
	return
}

// GetID returns the query to filter out specific ID
func GetID(id ID) (query interface{}) {
	return bson.D{
		{
			Key:   "_id",
			Value: id,
		},
	}
}

// Collection contains the mongo Collection to perform operations on a given collection
type Collection struct {
	*mongo.Collection      //Collection is a handle to a MongoDB collection.
	flag              bool // To log errors
}

// NewCollection returns a new mongo collection
func (c Client) NewCollection(db, collection string, flag bool) Collection {
	return Collection{
		c.Database(db).Collection(collection), flag,
	}
}

// Find finds documents based on a query condition.
// The data parameter must be a pointer to a slice, otherwise it will PANIC
func (c Collection) Find(query, projection, data interface{}) (err error) {

	opts := options.Find().SetProjection(projection)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cur, err := c.Collection.Find(ctx, query, opts)
	if err != nil {
		log.Println(err)
		return
	}
	if err = cur.All(ctx, data); err != nil {
		log.Println(err)
		logFindError(c, query, projection, data)
		return
	}
	return
}

// FindOne finds one document based on a query condition.
func (c Collection) FindOne(query, projection, data interface{}) (err error) {

	opts := options.FindOne().SetProjection(projection)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err = c.Collection.FindOne(ctx, query, opts).Decode(data)
	if err != nil {
		log.Println(err)
		logFindError(c, query, projection, data)
		return
	}
	return
}

// FindWithDistinct finds the distinct values for a specified field across a single collection.
func (c Collection) FindWithDistinct(distinctField string, query interface{}) (values []interface{}, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	values, err = c.Distinct(ctx, distinctField, query)
	if err != nil {
		log.Println(err)
		logError(c, query, values)
		return
	}
	return
}

// FindWithSortAndLimit finds sorted documents based on given conditions.
// The data parameter must be a pointer to a slice, otherwise it will PANIC.
// If limit is not required, Set limit = 0,
func (c Collection) FindWithSortAndLimit(query, projection, data interface{}, sort string, limit int) (err error) {

	newSort, err := c.EnsureIndexKey(sort)
	if err != nil {
		fmt.Println(err)
		return
	}
	opts := setOptions(newSort, projection, limit)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cur, err := c.Collection.Find(ctx, query, opts)
	if err != nil {
		log.Println(err)
		return
	}

	if err = cur.All(ctx, data); err != nil {
		log.Println(err)
		logFindWithSortError(c, query, projection, data, sort, limit)
		return
	}
	return
}

// InsertOne creates a new document.
func (c Collection) InsertOne(data interface{}) (err error) {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	res, err := c.Collection.InsertOne(ctx, data)
	if err != nil {
		log.Println(err)
		logResult(c, nil, data, res)
		return
	}
	return
}

// InsertMany creates new documents.
func (c Collection) InsertMany(data []interface{}) (err error) {

	if len(data) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	res, err := c.Collection.InsertMany(ctx, data)
	if err != nil {
		log.Println(err)
		logError(c, nil, data)
		return
	}
	createResultFileLog(res, "InsertResult.log")
	return
}

// DeleteOne deletes a single document from the collection.
func (c Collection) DeleteOne(query interface{}) (err error) {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	res, err := c.Collection.DeleteOne(ctx, query)
	if err != nil {
		log.Println(err)
		logResult(c, query, nil, res)
		return
	}
	return
}

// DeleteMany deletes multiple documents from the collection.
// query cannot be an empty document (bson.D{}). Use DeleteAll() to delete all the documents in the collection
func (c Collection) DeleteMany(query interface{}) (err error) {

	if reflect.ValueOf(query).Len() == 0 {
		err = errors.New("query cannot be empty! All the documents in the collection will be deleted")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	res, err := c.Collection.DeleteMany(ctx, query)
	if err != nil {
		log.Println(err)
		logError(c, query, nil)
		return
	}
	createResultFileLog(res, "DeleteResult.log")
	return
}

// DeleteAll deletes all the documents from the collection
func (c Collection) DeleteAll() (err error) {
	query := bson.D{}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	res, err := c.Collection.DeleteMany(ctx, query)
	if err != nil {
		log.Println(err)
		logError(c, query, nil)
		return
	}
	createResultFileLog(res, "DeleteAllResult.log")
	return
}

// UpdateOne updates a queried document.
func (c Collection) UpdateOne(query, data interface{}) (err error) {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	res, err := c.Collection.UpdateOne(ctx, query, data)
	if err != nil {
		log.Println(err)
		logResult(c, query, data, res)
		return
	}
	return
}

// UpsertOne allows the creation of a new document if no document matches the query
func (c Collection) UpsertOne(query, data interface{}) (err error) {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	opts := options.FindOneAndUpdate().SetUpsert(true)
	err = c.FindOneAndUpdate(ctx, query, data, opts).Decode(data)
	if err != nil {
		log.Println(err)
		logFindError(c, query, nil, data)
		return
	}
	return
}

// EnsureIndexKey ensures an index with the given key exists, creating it if necessary.
func (c Collection) EnsureIndexKey(sort string) (newSort string, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if sort != "-$natural" {
		indexView := c.Indexes()
		mod := mongo.IndexModel{
			Keys:    getIndexModelKeys(sort), // convert the sort string into bson.D
			Options: nil,
		}
		newSort, err = indexView.CreateOne(ctx, mod)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	return newSort, err // The index name will be generated as Eg: "name_1" for "name" and "name_-1" for "-name" by the driver
}

func getIndexModelKeys(sort string) bson.D {

	if strings.HasPrefix(sort, "-") {
		sort = strings.TrimPrefix(sort, "-")
		return bson.D{
			{
				Key:   sort,
				Value: -1,
			},
		}
	}
	return bson.D{
		{
			Key:   sort,
			Value: 1,
		},
	}
}

func setOptions(newSort string, projection interface{}, limit int) *options.FindOptions {

	opts := options.Find()
	opts.SetSort(convertintoBSON(newSort)) // convert the sort string into bson.D
	opts.SetProjection(projection)
	if limit != 0 {
		opts.SetLimit(int64(limit))
	}
	return opts
}

func convertintoBSON(sort string) bson.D {

	v := strings.Split(sort, "_") // The index name will be generated as Eg: "name_1" for "name" and "name_-1" for "-name" by the driver
	b := v[1]
	order, err := strconv.Atoi(b)
	if err != nil {
		log.Println(err)
	}
	return bson.D{
		{
			Key:   v[0],
			Value: order,
		},
	}
}
