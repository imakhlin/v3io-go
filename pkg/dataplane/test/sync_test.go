package test

import (
	"fmt"
	"testing"

	"github.com/v3io/v3io-go/pkg/dataplane"
	"github.com/v3io/v3io-go/pkg/errors"

	"github.com/stretchr/testify/suite"
)

type syncTestSuite struct {
	testSuite
}

//
// Container tests
//

type syncContainerTestSuite struct {
	syncTestSuite
}

func (suite *syncContainerTestSuite) TestGetContainers() {
	suite.T().Skip()
	suite.containerName = ""

	getContainersInput := v3io.GetContainersInput{}

	// when run against a context
	suite.populateDataPlaneInput(&getContainersInput.DataPlaneInput)

	// get containers
	response, err := suite.container.GetContainersSync(&getContainersInput)
	suite.Require().NoError(err, "Failed to get containers")

	getContainersOutput := response.Output.(*v3io.GetContainersOutput)
	fmt.Println(getContainersOutput)

	response.Release()
}

func (suite *syncContainerTestSuite) TestGetContainerContents() {
	getContainerContentsInput := v3io.GetContainerContentsInput{}

	// when run against a context
	suite.populateDataPlaneInput(&getContainerContentsInput.DataPlaneInput)

	// get container contents
	response, err := suite.container.GetContainerContentsSync(&getContainerContentsInput)
	suite.Require().NoError(err, "Failed to get container contents")

	getContainerContentsOutput := response.Output.(*v3io.GetContainerContentsOutput)
	fmt.Println(getContainerContentsOutput)

	response.Release()
}

type syncContextContainerTestSuite struct {
	syncContainerTestSuite
}

func (suite *syncContextContainerTestSuite) SetupSuite() {
	suite.syncContainerTestSuite.SetupSuite()

	suite.createContext()
}

type syncContainerContainerTestSuite struct {
	syncContainerTestSuite
}

func (suite *syncContainerContainerTestSuite) SetupSuite() {
	suite.syncContainerTestSuite.SetupSuite()

	suite.createContainer()
}

//
// Object tests
//

type syncObjectTestSuite struct {
	syncTestSuite
}

func (suite *syncObjectTestSuite) TestObject() {
	path := "/object.txt"
	contents := "vegans are better than everyone"

	getObjectInput := &v3io.GetObjectInput{
		Path: path,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getObjectInput.DataPlaneInput)

	response, err := suite.container.GetObjectSync(getObjectInput)

	// get the underlying root error
	errWithStatusCode, errHasStatusCode := err.(v3ioerrors.ErrorWithStatusCode)
	suite.Require().True(errHasStatusCode)
	suite.Require().Equal(404, errWithStatusCode.StatusCode())

	//
	// PUT contents to some object
	//

	putObjectInput := &v3io.PutObjectInput{
		Path: path,
		Body: []byte(contents),
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&putObjectInput.DataPlaneInput)

	err = suite.container.PutObjectSync(putObjectInput)

	suite.Require().NoError(err, "Failed to put")

	//
	// Get the contents
	//

	getObjectInput = &v3io.GetObjectInput{
		Path: path,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getObjectInput.DataPlaneInput)

	response, err = suite.container.GetObjectSync(getObjectInput)

	suite.Require().NoError(err, "Failed to get")

	// make sure buckets is not empty
	suite.Require().Equal(contents, string(response.Body()))

	// release the response
	response.Release()

	//
	// Delete the object
	//

	deleteObjectInput := &v3io.DeleteObjectInput{
		Path: path,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&deleteObjectInput.DataPlaneInput)

	err = suite.container.DeleteObjectSync(deleteObjectInput)

	suite.Require().NoError(err, "Failed to delete")

	//
	// Get the contents again (should fail)
	//

	getObjectInput = &v3io.GetObjectInput{
		Path: path,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getObjectInput.DataPlaneInput)

	response, err = suite.container.GetObjectSync(getObjectInput)

	suite.Require().Error(err, "Failed to get")
	suite.Require().Nil(response)
}

type syncContextObjectTestSuite struct {
	syncObjectTestSuite
}

func (suite *syncContextObjectTestSuite) SetupSuite() {
	suite.syncObjectTestSuite.SetupSuite()

	suite.createContext()
}

type syncContainerObjectTestSuite struct {
	syncObjectTestSuite
}

func (suite *syncContainerObjectTestSuite) SetupSuite() {
	suite.syncObjectTestSuite.SetupSuite()

	suite.createContainer()
}

//
// KV tests
//

type syncKVTestSuite struct {
	syncTestSuite
	items map[string]map[string]interface{}
}

func (suite *syncKVTestSuite) TestEMD() {
	itemsToCreate := map[string]map[string]interface{}{
		"bob":    {"age": 42, "feature": "mustache"},
		"linda":  {"age": 41, "feature": "singing"},
		"louise": {"age": 9, "feature": "bunny ears"},
		"tina":   {"age": 14, "feature": "butts"},
	}

	//
	// Create items one by one
	//

	// create the items
	for itemToCreateKey, itemToCreateAttributes := range itemsToCreate {
		input := v3io.PutItemInput{
			Path:       "/emd0/" + itemToCreateKey,
			Attributes: itemToCreateAttributes,
		}

		// when run against a context, will populate fields like container name
		suite.populateDataPlaneInput(&input.DataPlaneInput)

		// get a specific bucket
		err := suite.container.PutItemSync(&input)
		suite.Require().NoError(err, "Failed to put item")
	}

	suite.verifyItems(itemsToCreate)

	//
	// Update item and verify
	//

	// update louise item
	updateItemInput := v3io.UpdateItemInput{
		Path: "/emd0/louise",
		Attributes: map[string]interface{}{
			"height": 130,
			"quip":   "i can smell fear on you",
		},
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&updateItemInput.DataPlaneInput)

	err := suite.container.UpdateItemSync(&updateItemInput)
	suite.Require().NoError(err, "Failed to update item")

	// get louise
	getItemInput := v3io.GetItemInput{
		Path:           "/emd0/louise",
		AttributeNames: []string{"__size", "age", "quip", "height"},
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getItemInput.DataPlaneInput)

	response, err := suite.container.GetItemSync(&getItemInput)
	suite.Require().NoError(err, "Failed to get item")

	getItemOutput := response.Output.(*v3io.GetItemOutput)

	// make sure we got the age and quip correctly
	suite.Require().Equal(0, getItemOutput.Item["__size"].(int))
	suite.Require().Equal(130, getItemOutput.Item["height"].(int))
	suite.Require().Equal("i can smell fear on you", getItemOutput.Item["quip"].(string))
	suite.Require().Equal(9, getItemOutput.Item["age"].(int))

	// release the response
	response.Release()

	// get all items whose age is over 15
	getItemsInput := v3io.GetItemsInput{
		Path:           "/emd0/",
		AttributeNames: []string{"age", "feature"},
		Filter:         "age > 15",
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getItemsInput.DataPlaneInput)

	cursor, err := v3io.NewItemsCursor(suite.container, &getItemsInput)
	suite.Require().NoError(err, "Failed to get items")

	cursorItems, err := cursor.AllSync()
	suite.Require().NoError(err)
	suite.Require().Len(cursorItems, 2)

	// iterate over age, make sure it's over 15
	for _, cursorItem := range cursorItems {
		age, err := cursorItem.GetFieldInt("age")
		suite.Require().NoError(err)
		suite.Require().True(age > 15)
	}

	// release the response
	response.Release()

	//
	// Increment age
	//

	incrementAgeExpression := "age = age + 1"

	// update louise's age
	updateItemInput = v3io.UpdateItemInput{
		Path:       "/emd0/louise",
		Expression: &incrementAgeExpression,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&updateItemInput.DataPlaneInput)

	err = suite.container.UpdateItemSync(&updateItemInput)
	suite.Require().NoError(err, "Failed to update item")

	// get tina
	getItemInput = v3io.GetItemInput{
		Path:           "/emd0/louise",
		AttributeNames: []string{"age"},
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getItemInput.DataPlaneInput)

	response, err = suite.container.GetItemSync(&getItemInput)
	suite.Require().NoError(err, "Failed to get item")

	getItemOutput = response.Output.(*v3io.GetItemOutput)

	// check that age incremented
	suite.Require().Equal(10, getItemOutput.Item["age"].(int))

	// release the response
	response.Release()

	//
	// Delete everything
	//

	suite.deleteItems(itemsToCreate)
}

func (suite *syncKVTestSuite) TestPutItems() {
	items := map[string]map[string]interface{}{
		"bob":   {"age": 42, "feature": "mustache", "married": false, "timestamp": int64(1556450700000)},
		"linda": {"age": 41, "feature": "singing", "married": true, "timestamp": int64(1556450700000)},
	}

	putItemsInput := &v3io.PutItemsInput{
		Path:  "/emd0",
		Items: items,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&putItemsInput.DataPlaneInput)

	// get a specific bucket
	response, err := suite.container.PutItemsSync(putItemsInput)
	suite.Require().NoError(err, "Failed to put items")

	putItemsOutput := response.Output.(*v3io.PutItemsOutput)

	// must succeed - everything was valid
	suite.Require().True(putItemsOutput.Success)
	suite.Require().Nil(putItemsOutput.Errors)

	response.Release()

	suite.verifyItems(items)

	suite.deleteItems(items)
}

func (suite *syncKVTestSuite) TestPutItemsWithError() {
	items := map[string]map[string]interface{}{
		"bob":     {"age": 42, "feature": "mustache"},
		"linda":   {"age": 41, "feature": "singing"},
		"invalid": {"__name": "foo", "feature": "singing"},
	}

	putItemsInput := &v3io.PutItemsInput{
		Path:  "/emd0",
		Items: items,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&putItemsInput.DataPlaneInput)

	// get a specific bucket
	response, err := suite.container.PutItemsSync(putItemsInput)
	suite.Require().NoError(err, "Failed to put items")

	putItemsOutput := response.Output.(*v3io.PutItemsOutput)

	// must succeed - everything was valid
	suite.Require().False(putItemsOutput.Success)
	suite.Require().NotNil(putItemsOutput.Errors)
	suite.Require().NotNil(putItemsOutput.Errors["invalid"])

	response.Release()

	// remove invalid because it shouldn't be verified / deleted
	delete(items, "invalid")

	suite.verifyItems(items)

	suite.deleteItems(items)
}

func (suite *syncKVTestSuite) verifyItems(items map[string]map[string]interface{}) {

	// get all items
	getItemsInput := v3io.GetItemsInput{
		Path:           "/emd0/",
		AttributeNames: []string{"*"},
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getItemsInput.DataPlaneInput)

	cursor, err := v3io.NewItemsCursor(suite.container, &getItemsInput)
	suite.Require().NoError(err, "Failed to create cursor")

	receivedItems, err := cursor.AllSync()
	suite.Require().NoError(err)
	suite.Require().Len(receivedItems, len(items))

	// TODO: test values

	// release the response
	cursor.Release()
}

func (suite *syncKVTestSuite) deleteItems(items map[string]map[string]interface{}) {

	// delete the items
	for itemKey, _ := range items {
		input := v3io.DeleteObjectInput{
			Path: "/emd0/" + itemKey,
		}

		// when run against a context, will populate fields like container name
		suite.populateDataPlaneInput(&input.DataPlaneInput)

		// get a specific bucket
		err := suite.container.DeleteObjectSync(&input)
		suite.Require().NoError(err, "Failed to delete item")
	}

	input := &v3io.DeleteObjectInput{
		Path: "/emd0/",
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&input.DataPlaneInput)

	// delete the directory
	err := suite.container.DeleteObjectSync(input)

	suite.Require().NoError(err, "Failed to delete")
}

type syncContextKVTestSuite struct {
	syncKVTestSuite
}

func (suite *syncContextKVTestSuite) SetupSuite() {
	suite.syncKVTestSuite.SetupSuite()

	suite.createContext()
}

type syncContainerKVTestSuite struct {
	syncKVTestSuite
}

func (suite *syncContainerKVTestSuite) SetupSuite() {
	suite.syncKVTestSuite.SetupSuite()

	suite.createContainer()
}

//
// Stream tests
//

type syncStreamTestSuite struct {
	syncTestSuite
	testPath string
}

func (suite *syncStreamTestSuite) SetupTest() {
	suite.testPath = "/stream-test"
	suite.deleteAllStreamsInPath(suite.testPath)
}

func (suite *syncStreamTestSuite) TearDownTest() {
	suite.deleteAllStreamsInPath(suite.testPath)
}

func (suite *syncStreamTestSuite) TestStream() {
	streamPath := fmt.Sprintf("%s/mystream/", suite.testPath)

	//
	// Create the stream
	//

	createStreamInput := v3io.CreateStreamInput{
		Path:                 streamPath,
		ShardCount:           4,
		RetentionPeriodHours: 1,
	}

	suite.populateDataPlaneInput(&createStreamInput.DataPlaneInput)

	err := suite.container.CreateStreamSync(&createStreamInput)

	suite.Require().NoError(err, "Failed to create stream")

	//
	// Put some records
	//

	firstShardID := 1
	secondShardID := 2
	invalidShardID := 10

	records := []*v3io.StreamRecord{
		{ShardID: &firstShardID, Data: []byte("first shard record #1")},
		{ShardID: &firstShardID, Data: []byte("first shard record #2")},
		{ShardID: &invalidShardID, Data: []byte("invalid shard record #1")},
		{ShardID: &secondShardID, Data: []byte("second shard record #1")},
		{Data: []byte("some shard record #1")},
	}

	putRecordsInput := v3io.PutRecordsInput{
		Path:    streamPath,
		Records: records,
	}

	suite.populateDataPlaneInput(&putRecordsInput.DataPlaneInput)

	response, err := suite.container.PutRecordsSync(&putRecordsInput)
	suite.Require().NoError(err, "Failed to put records")

	putRecordsResponse := response.Output.(*v3io.PutRecordsOutput)

	// should have one failure
	suite.Require().Equal(1, putRecordsResponse.FailedRecordCount)

	// verify record results
	for recordIdx, record := range putRecordsResponse.Records {

		// third record should've failed
		if recordIdx == 2 {
			suite.Require().NotEqual(0, record.ErrorCode)
		} else {
			suite.Require().Equal(0, record.ErrorCode)
		}
	}

	response.Release()

	//
	// Seek
	//

	seekShardInput := v3io.SeekShardInput{
		Path: streamPath + "1",
		Type: v3io.SeekShardInputTypeEarliest,
	}

	suite.populateDataPlaneInput(&seekShardInput.DataPlaneInput)

	response, err = suite.container.SeekShardSync(&seekShardInput)

	suite.Require().NoError(err, "Failed to seek shard")
	location := response.Output.(*v3io.SeekShardOutput).Location

	suite.Require().NotEqual("", location)

	response.Release()

	//
	// Get records
	//

	getRecordsInput := v3io.GetRecordsInput{
		Path:     streamPath + "1",
		Location: location,
		Limit:    100,
	}

	suite.populateDataPlaneInput(&getRecordsInput.DataPlaneInput)

	response, err = suite.container.GetRecordsSync(&getRecordsInput)

	suite.Require().NoError(err, "Failed to get records")

	getRecordsOutput := response.Output.(*v3io.GetRecordsOutput)

	suite.Require().Equal("first shard record #1", string(getRecordsOutput.Records[0].Data))
	suite.Require().Equal("first shard record #2", string(getRecordsOutput.Records[1].Data))

	response.Release()

	//
	// Delete stream
	//

	deleteStreamInput := v3io.DeleteStreamInput{
		Path: streamPath,
	}

	suite.populateDataPlaneInput(&deleteStreamInput.DataPlaneInput)

	err = suite.container.DeleteStreamSync(&deleteStreamInput)
	suite.Require().NoError(err, "Failed to delete stream")
}

func (suite *syncStreamTestSuite) deleteAllStreamsInPath(path string) error {

	getContainerContentsInput := v3io.GetContainerContentsInput{
		Path: path,
	}

	suite.populateDataPlaneInput(&getContainerContentsInput.DataPlaneInput)

	// get all streams in the test path
	response, err := suite.container.GetContainerContentsSync(&getContainerContentsInput)

	if err != nil {
		return err
	}

	defer response.Release()

	// iterate over streams (prefixes) and delete them
	for _, commonPrefix := range response.Output.(*v3io.GetContainerContentsOutput).CommonPrefixes {
		deleteStreamInput := v3io.DeleteStreamInput{
			Path: "/" + commonPrefix.Prefix,
		}

		suite.populateDataPlaneInput(&deleteStreamInput.DataPlaneInput)

		suite.container.DeleteStreamSync(&deleteStreamInput)
	}

	return nil
}

type syncContextStreamTestSuite struct {
	syncStreamTestSuite
}

func (suite *syncContextStreamTestSuite) SetupSuite() {
	suite.syncStreamTestSuite.SetupSuite()

	suite.createContext()
}

type syncContainerStreamTestSuite struct {
	syncStreamTestSuite
}

func (suite *syncContainerStreamTestSuite) SetupSuite() {
	suite.syncStreamTestSuite.SetupSuite()

	suite.createContainer()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSyncSuite(t *testing.T) {
	suite.Run(t, new(syncContextContainerTestSuite))
	suite.Run(t, new(syncContextObjectTestSuite))
	suite.Run(t, new(syncContainerObjectTestSuite))
	suite.Run(t, new(syncContextKVTestSuite))
	suite.Run(t, new(syncContainerKVTestSuite))
	suite.Run(t, new(syncContextStreamTestSuite))
	suite.Run(t, new(syncContainerStreamTestSuite))
}
