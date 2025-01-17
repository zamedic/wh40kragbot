package db

type Embedding struct {
	Vector   []float32 `bson:"vector"`
	Page     int       `bson:"page"`
	Document string    `bson:"document"`
	Index    int       `bson:"index"`
}
