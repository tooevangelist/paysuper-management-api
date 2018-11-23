package model

type Log struct {
	RequestHeaders  string `bson:"request_headers"`
	RequestBody     string `bson:"request_body"`
	ResponseHeaders string `bson:"response_headers"`
	ResponseBody    string `bson:"response_body"`
}
