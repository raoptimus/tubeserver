package mongodb

import (
	"crypto/sha1"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

type Sequence struct {
	Id     string `bson:"_id"`
	LastId int    `bson:"lastId"`
}

func GetNewIncId(c *mgo.Collection) (id int, err error) {
	err = IncAndGet(c.Database.C("Sequence").Find(bson.M{"_id": c.Name + "__id"}), "lastId", 1, &id)
	return
}

func IncAndGet(q *mgo.Query, field string, inc int, ret *int) error {
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{field: inc}},
		ReturnNew: true,
		Upsert:    true,
	}

	var result map[string]interface{}
	sel := bson.M{field: 1}
	if field != "_id" {
		sel["_id"] = -1
	}
	_, err := q.Select(sel).Apply(change, &result)
	if err != nil {
		return err
	}

	*ret = result[field].(int)
	return nil
}

func GenerateObjectId(args ...string) bson.ObjectId {
	data := strings.Join(args, "|")
	b := sha1.Sum([]byte(data))

	return bson.ObjectId(b[:12])
}
