package mongo

import "errors"

// CollectionName is a generic type for view26 collections
type CollectionName string

const (
	// AccountsDB contains user account info
	AccountsDB CollectionName = "accounts"
	// ViewShareDB contains shared views info
	ViewShareDB CollectionName = "view-share"
	// IssuesDB contains issues info
	IssuesDB CollectionName = "issues"
	// FieldsDB contains fields info
	FieldsDB CollectionName = "fields"
	// WidgetsDB contains widgets info
	WidgetsDB CollectionName = "widgets"
	// UsersDB contains users info
	UsersDB CollectionName = "users"
	// ViewsDB contains views info
	ViewsDB CollectionName = "views"
	// FieldValueDB contains fieldvalue info
	FieldValueDB CollectionName = "fieldvalue"
)

// GetCollectionName returns the proper collection name
func GetCollectionName(accID ID, collection CollectionName) (c string, err error) {
	if collection == "" {
		err = errors.New("collection name cannot be empty string")
		return "", err
	}
	acc := accID.Hex()
	switch collection {
	case AccountsDB, ViewShareDB:
		return string(collection), nil
	case IssuesDB, FieldsDB, WidgetsDB, ViewsDB, FieldValueDB, UsersDB:
		return acc + "-" + string(collection), nil
	default:
	}
	return "", errors.New("invalid collection name")
}
