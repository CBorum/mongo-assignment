package main

import (
	"time"
	"log"
	"regexp"
	"sort"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	dbConn *mgo.Session
	dbURL = "mongodb://localhost:27017/social_net"
)

const (
	database = "social_net"
	collection = "tweets"
)

func main() {
	log.Println("starting...")

	if os.Getenv("DB_URL") != "" {
		dbURL = os.Getenv("DB_URL")
	}

	var err error
	dbConn, err = mgo.Dial(dbURL)
	defer dbConn.Close()
	if err != nil {
		panic(err)
	}

	go setIndexes()

	server := gin.Default()
	server.GET("/", endpoints)
	server.GET("/users", users)
	server.GET("/mentioners", topMentioners)
	server.GET("/mentioned", topMentioned)
	server.GET("/active", mostActive)
	server.GET("/polarity", topPolarity)
	server.Run()
}

func endpoints(c *gin.Context) {
	c.JSON(200, gin.H{
		"users": "/users",
		"top mentioners": "/mentioners",
		"top mentioned":  "/mentioned",
		"most active": "/active",
		"most grumpy and most happy": "/polarity",
	})
}

// How many Twitter users are in the database?
// Which Twitter users link the most to other Twitter users? (Provide the top ten.)
// Who is are the most mentioned Twitter users? (Provide the top five.)
// Who are the most active Twitter users (top ten)?
// Who are the five most grumpy (most negative tweets) and the most happy (most positive tweets)? (Provide five users for each group)

func users(c *gin.Context) {
	var result []interface{}
	err := tweetsColl().Find(bson.M{}).Distinct("user", &result) // returns all the distict users in an array, not just the length which would be optimal
	if err != nil {
		log.Println(err)
		c.JSON(500, "error")
		return
	}

	c.JSON(200, gin.H{"users": len(result)})
}

func topMentioners(c *gin.Context) {
	var result []interface{}
	pipe := tweetsColl().Pipe([]bson.M{
		{
			"$unwind": "$text",
		},
		{
			"$match": bson.M{
				"text": bson.M{"$regex": `@\w+`},
			},
		},
		{
			"$group": bson.M{
				"_id":   bson.M{"user": "$user"},
				"count": bson.M{"$sum": 1},
			},
		},
		{"$sort": bson.M{"count": -1}},
		{"$limit": 10},
	})

	err := pipe.All(&result)
	if err != nil {
		log.Println(err)
		c.JSON(500, "error")
		return
	}

	c.JSON(200, result)
}

func topMentioned(c *gin.Context) {
	pipe := tweetsColl().Pipe(mentionedQuery)

	r := regexp.MustCompile(`@(\w+)`)
	mentions := map[string]int{}
	iter := pipe.Iter()
	tStart := time.Now()	

	var res map[string]interface{}
	result := mentionResult{}
	for iter.Next(&res) { // mgo's way of iterating over the results returned
		str := res["text"].(string) // assert the type of the value returned
		matches := r.FindAllStringSubmatch(str, -1)
		for _, v := range matches { // iterate over matches. There can be multiple mentions in one tweet
			if index, ok := mentions[v[1]]; ok {
				result[index].Count++
			} else {
				mentions[v[1]] = len(result)
				result = append(result, &mention{User: v[1], Count: 1})
			}
		}
	}

	// cost: O(n*log(n))
	sort.Sort(result)
	log.Println("process time", time.Since(tStart))
	c.JSON(200, result[:10])
}

// func topMentioned2(c *gin.Context) {
// 	pipe := tweetsColl().Pipe(mentionedQuery)
	
// 	r := regexp.MustCompile(`@(\w+)`)
// 	var res []map[string]string

// 	err := pipe.All(&res)
// 	if err != nil {
// 		log.Println(err)
// 		c.JSON(500, "error")
// 		return
// 	}
// 	tStart := time.Now()

// 	mentions := map[string]int{}
// 	result := mentionResult{}

// 	for _, value := range res {
// 		str := value["text"] // assert the type of the value returned
// 		matches := r.FindAllStringSubmatch(str, -1)
// 		for _, v := range matches { // iterate over matches. There can be multiple mentions in one tweet
// 			if index, ok := mentions[v[1]]; ok {
// 				result[index].Count++
// 			} else {
// 				mentions[v[1]] = len(result)
// 				result = append(result, &mention{User: v[1], Count: 1})
// 			}
// 		}
// 	}

// 	// cost: O(n*log(n))
// 	sort.Sort(result)
// 	log.Println("process time", time.Since(tStart))
// 	c.JSON(200, result[:10])
// }

func mostActive(c *gin.Context) {
	var result []interface{}
	pipe := tweetsColl().Pipe([]bson.M{
		{
			"$group": bson.M{
				"_id":   bson.M{"user": "$user"},
				"count": bson.M{"$sum": 1},
			},
		},
		{"$sort": bson.M{"count": -1}},
		{"$limit": 10},
	})

	err := pipe.All(&result)
	if err != nil {
		log.Println(err)
		c.JSON(500, "error")
		return
	}

	log.Println(result)
	c.JSON(200, result)
}

func topPolarity(c *gin.Context) {
	nChan := make(chan []interface{})
	pChan := make(chan []interface{})

	go pipeExec(tweetsColl().Pipe(getPolarityQuery(0)), nChan)
	go pipeExec(tweetsColl().Pipe(getPolarityQuery(4)), pChan)

	negative := <-nChan
	positive := <-pChan

	result := map[string]interface{}{
		"negative": negative,
		"positive": positive,
	}

	c.JSON(200, result)
}

func pipeExec(pipe *mgo.Pipe, done chan []interface{}) {
	var res []interface{}
	err := pipe.All(&res)
	if err != nil {
		log.Println(err)
	}
	done <- res
}

func getPolarityQuery(polarity int) []bson.M {
	return []bson.M{
		{
			"$match": bson.M{
				"polarity": polarity,
			},
		},
		{
			"$group": bson.M{
				"_id":   bson.M{"user": "$user"},
				"count": bson.M{"$sum": 1},
			},
		},
		{"$sort": bson.M{"count": -1}},
		{"$limit": 5},
	}
}

var mentionedQuery = []bson.M{
	{
		"$unwind": "$text",
	},
	{
		"$match": bson.M{
			"text": bson.M{"$regex": `@\w+`},
		},
	},
	{
		"$project": bson.M{
			"text": 1,
		},
	},
}

func tweetsColl() *mgo.Collection {
	return dbConn.DB(database).C(collection)
}

type mention struct {
	User  string
	Count int
}

// for sorting

type mentionResult []*mention

func (s mentionResult) Len() int {
	return len(s)
}

func (s mentionResult) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s mentionResult) Less(i, j int) bool {
	return s[i].Count > s[j].Count
}

// indexes

func setIndexes() {
	if dbConn == nil {
		panic("no dbConn")
	}

	indexes, err := tweetsColl().Indexes()
	if err != nil {
		log.Println(err)
	}

	if len(indexes) < 2 {
		log.Println("setting indexes")
		tweetsColl().EnsureIndexKey("user")
		tweetsColl().EnsureIndexKey("$text:text")
		tweetsColl().EnsureIndexKey("polarity")
	}
}
