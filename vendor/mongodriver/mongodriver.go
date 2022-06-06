package mongodriver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
	mongolib "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Connect mongo ssl
const (
	connectMongoSSL string = "/mongo-ssl/%s"
)

// Mongo is struct
type Mongo struct {
	Hosts     []string           `json:"hosts"`
	HostSRV   string             `json:"host_srv"`
	Username  string             `json:"username"`
	Password  string             `json:"password"`
	IsAtlas   bool               `json:"is_atlas"`
	UsingSSL  bool               `json:"using_ssl"`
	CaFile    string             `json:"ca_file"`
	ClientCrt string             `json:"client_crt"`
	ClientKey string             `json:"client_key"`
	Mechanism string             `json:"mechanism"`
	SetAuth   bool               `json:"set_auth"`
	Readpref  *readpref.ReadPref `json:"readpref"`
	Session   mongo.Session      `json:"session"`
	Client    *mongo.Client      `json:"client"`
	Database  mongolib.Database  `json:"database"`
}

// Connect is func to connect server
func (mongo *Mongo) Connect(mongoMode string) mongolib.Session {
	/*
		mode mongo-driver

		readpref.Primary() Mode 1 = Strong with mgo Mode 2
		// PrimaryMode indicates that only a primary is
		// considered for reading. This is the default
		// mode.

		readpref.PrimaryPreferredMode() Mode 2 = PrimaryPreferred with mgo Mode 3
		// PrimaryPreferredMode indicates that if a primary
		// is available, use it; otherwise, eligible
		// secondaries will be considered.

		readpref.SecondaryMode() Mode 3 = Secondary with mgo Mode 4
		// SecondaryMode indicates that only secondaries
		// should be considered.

		readpref.SecondaryPreferredMode() Mode 4 = SecondaryPreferred with mgo Mode 5
		// SecondaryPreferredMode indicates that only secondaries
		// should be considered when one is available. If none
		// are available, then a primary will be considered.

		readpref.NearestMode() Mode 5 = Nearest with mgo Mode 6
		// NearestMode indicates that all primaries and secondaries
		// will be considered.
	*/

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	// Setting
	clientOptions := options.ClientOptions{}

	if !mongo.IsAtlas && !mongo.UsingSSL {
		clientOptions.SetHosts(mongo.Hosts)
	}

	if mongo.UsingSSL {
		// --sslCAFile
		rootCerts := x509.NewCertPool()
		if ca, err := ioutil.ReadFile(fmt.Sprintf(connectMongoSSL, mongo.CaFile)); err != nil {
			panic(fmt.Sprintf("ERROR Ca_file:%v", err))
		} else {
			fmt.Println("Read Ca_file")
			rootCerts.AppendCertsFromPEM(ca)
		}

		// --sslPEMKeyFile
		clientCerts := []tls.Certificate{}
		if cert, err := tls.LoadX509KeyPair(fmt.Sprintf(connectMongoSSL, mongo.ClientCrt), fmt.Sprintf(connectMongoSSL, mongo.ClientKey)); err != nil {
			panic(fmt.Sprintf("ERROR Client_Certs:%v", err))
		} else {
			fmt.Println("Read Client_Certs")
			clientCerts = append(clientCerts, cert)
		}
		tlsConfig := &tls.Config{}
		tlsConfig.RootCAs = rootCerts
		tlsConfig.Certificates = clientCerts

		clientOptions.SetTLSConfig(tlsConfig)
	}

	if mongo.Username != "" && mongo.Password != "" && mongo.Mechanism != "" && len(mongo.Hosts) > 0 {
		clientOptions.ApplyURI(
			fmt.Sprintf(
				"mongodb+srv://%s:%s@%s/test?ssl=true",
				mongo.Username,
				mongo.Password,
				mongo.Hosts[0],
			),
		)
	}

	modeMongo, err := readpref.ModeFromString(mongoMode)
	if err != nil {
		panic(fmt.Sprintf("Cannot select mode : %s", err))
	}

	readPrefMongo, err := readpref.New(modeMongo)
	if err != nil {
		panic(fmt.Sprintf("Cannot select mode : %s", err))
	}
	clientOptions.SetReadPreference(readPrefMongo)

	client, err := mongolib.Connect(ctx, &clientOptions)
	if err != nil {
		panic(fmt.Sprintf("Cannot connect mongo server: %s", err))
	}

	// Check Connection
	err = client.Ping(ctx, readPrefMongo)
	if err != nil {
		panic(fmt.Sprintf("Connection check fail %s", err))
	}

	sessionOptions := options.Session()
	sessionOptions.SetCausalConsistency(true)
	sessionOptions.SetDefaultReadConcern(readconcern.Majority())
	session, errSession := client.StartSession(sessionOptions)
	if errSession != nil {
		panic(fmt.Sprintf("Cannot create session mongo : %s", errSession))
	}
	mongo.Session = session
	mongo.Readpref = readPrefMongo
	mongo.Client = client

	return mongo.Session
}

// ConnectionCheck implements db reconnection
func (mongo *Mongo) ConnectionCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	if err := mongo.Client.Ping(ctx, mongo.Readpref); err != nil {
		fmt.Println(fmt.Sprintf("Lost connection to db!"))
		// mongo.Session.Refresh()
		// client := mongo.Session.Client()
		if err := mongo.Client.Ping(ctx, mongo.Readpref); err == nil {
			fmt.Println(fmt.Sprintf("Reconnect to db successful."))
		}
	}
}

// ChangeSchema is change database by schema
func (mongo Mongo) ChangeSchema(databaseName string) *mongolib.Database {
	mongo.ConnectionCheck()
	client := mongo.Session.Client()
	database := client.Database(databaseName)
	return database
}

// ChangeCollection is change collection by name
func (mongo Mongo) ChangeCollection(collectionName string) *mongolib.Collection {
	collection := mongo.Database.Collection(collectionName)
	return collection
}

// GetOne is find one data in collection
func (mongo Mongo) GetOne(collection *mongolib.Collection, query map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	ctx := context.Background()
	err := collection.FindOne(ctx, query).Decode(&result)
	if err != nil {
		return make(map[string]interface{}, 0), err
	}
	return result, nil
}

// GetOneWithHint is find one data in collection
func (mongo Mongo) GetOneWithHint(collection *mongolib.Collection, query map[string]interface{}, hint []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	ctx := context.Background()
	var err error
	if len(hint) > 0 {
		findOptions := options.FindOne()
		findOptions.SetHint(hint)
		err = collection.FindOne(ctx, query, findOptions).Decode(&result)
	} else {
		result, err = mongo.GetOne(collection, query)
	}
	return result, err
}

// GetAll is find all data in collection
func (mongo Mongo) GetAll(collection *mongolib.Collection, query map[string]interface{}) ([]map[string]interface{}, error) {
	var (
		jsonDocuments []map[string]interface{}
	)
	ctx := context.Background()
	cursor, err := collection.Find(ctx, query)
	if err != nil {
		return make([]map[string]interface{}, 0), err
	}

	if err = cursor.All(ctx, &jsonDocuments); err != nil {
		return make([]map[string]interface{}, 0), err
	}
	return jsonDocuments, nil
}

// GetAllWithHint is find all data in collection with hint
func (mongo Mongo) GetAllWithHint(collection *mongolib.Collection, query map[string]interface{}, hint []string) ([]map[string]interface{}, error) {
	var (
		jsonDocuments []map[string]interface{}
		err           error
	)
	ctx := context.Background()
	if len(hint) > 0 {
		findOptions := options.Find()
		findOptions.SetHint(hint)
		cursor, err := collection.Find(ctx, query, findOptions)
		if err != nil {
			return make([]map[string]interface{}, 0), err
		}
		if err = cursor.All(ctx, &jsonDocuments); err != nil {
			return make([]map[string]interface{}, 0), err
		}
	} else {
		jsonDocuments, err = mongo.GetAll(collection, query)
		if err != nil {
			return make([]map[string]interface{}, 0), err
		}
	}

	return jsonDocuments, nil
}

// GetAllWithPageLimit is find all data in collection with page and limit
func (mongo Mongo) GetAllWithPageLimit(collection *mongolib.Collection, query map[string]interface{}, page int, limit int) ([]map[string]interface{}, error) {
	var (
		jsonDocuments []map[string]interface{}
		err           error
	)
	ctx := context.Background()
	findOptions := options.Find()
	if page == 0 && limit == 0 {
		findOptions.SetSort(bson.D{{"created_at", -1}})
	} else {
		findOptions.SetSort(bson.D{{"created_at", -1}}).SetSkip(int64((page - 1) * limit)).SetLimit(int64(limit))
	}
	cursor, _ := collection.Find(nil, query, findOptions)
	if err = cursor.All(ctx, &jsonDocuments); err != nil {
		return make([]map[string]interface{}, 0), err
	}
	return jsonDocuments, nil
}

// GetAllWithPageLimitAndHint is find all data in collection with page and limit
func (mongo Mongo) GetAllWithPageLimitAndHint(collection *mongolib.Collection, query map[string]interface{}, page int, limit int, hint []string) ([]map[string]interface{}, error) {
	var (
		jsonDocuments []map[string]interface{}
		err           error
	)
	ctx := context.Background()
	findOptions := options.Find()
	if page == 0 && limit == 0 {
		findOptions.SetSort(bson.D{{"created_at", -1}})
		if len(hint) > 0 {
			findOptions.SetHint(hint)
		}
	} else {
		if len(hint) > 0 {
			findOptions.SetHint(hint)
		}
		findOptions.SetSort(bson.D{{"created_at", -1}}).SetSkip(int64((page - 1) * limit)).SetLimit(int64(limit))
	}
	cursor, _ := collection.Find(nil, query, findOptions)
	if err = cursor.All(ctx, &jsonDocuments); err != nil {
		return make([]map[string]interface{}, 0), err
	}
	return jsonDocuments, nil
}

// UpdateOne is update documents with query matched on collection
func (mongo Mongo) UpdateOne(collection *mongolib.Collection, query map[string]interface{}, updateInfo interface{}) (map[string]interface{}, error) {
	var (
		jsonDocument map[string]interface{}
	)
	ctx := context.Background()
	if _, err := collection.UpdateOne(ctx, query, updateInfo); err != nil {
		return jsonDocument, err
	}
	jsonDocument, err := mongo.GetOne(collection, query)
	return jsonDocument, err
}

// UpdateAll is update documents with query matched on collection
func (mongo Mongo) UpdateAll(collection *mongolib.Collection, query map[string]interface{}, updateInfo interface{}) ([]map[string]interface{}, error) {
	var (
		jsonDocuments []map[string]interface{}
	)
	ctx := context.Background()
	if _, err := collection.UpdateMany(ctx, query, updateInfo); err != nil {
		return jsonDocuments, err
	}
	jsonDocuments, errGetAll := mongo.GetAll(collection, query)
	return jsonDocuments, errGetAll
}

// Count is count data in collection
func (mongo Mongo) Count(collection *mongolib.Collection, query map[string]interface{}) (int64, error) {
	ctx := context.Background()
	count, err := collection.CountDocuments(ctx, query)
	if err != nil {
		return count, err
	}
	return count, err
}

// Insert data in collection
func (mongo Mongo) Insert(collection *mongolib.Collection, data map[string]interface{}) error {
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, data)
	if err != nil {
		return err
	}
	return err
}

// BulkWrite in collection
func (mongo Mongo) BulkWrite(collection *mongolib.Collection, models []mongo.WriteModel) error {
	ctx := context.Background()
	opts := options.BulkWrite().SetOrdered(false)
	_, err := collection.BulkWrite(ctx, models, opts)
	if err != nil {
		return err
	}
	return err
}
