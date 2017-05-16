package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	//"strconv"
	"encoding/json"
	//"time"
	//"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

const STATE_OBD_REQUEST_CREATED = 0

var dispatchOrderIndexstr = "_dispatchOrderindex"

type DispatchOrderObject struct {
	dispatchOrderId string
	stage           string
	customer        string
	timeStamp       string // This is the time stamp
}

var tables = []string{"AssetTable", "TransactionHistory", "DocumentTable"}

func GetNumberOfKeys(tname string) int {
	TableMap := map[string]int{
		"AssetTable":         2,
		"TransactionHistory": 3,
		"DocumentTable":      2,
	}
	return TableMap[tname]
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes the chain and two tables - one for asset and other for transaction history
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	fmt.Println("Application Init")
	var err error

	for _, val := range tables {
		err = stub.DeleteTable(val)
		if err != nil {
			return nil, fmt.Errorf("Init(): DeleteTable of %s  Failed ", val)
		}
		err = InitLedger(stub, val)
		if err != nil {
			return nil, fmt.Errorf("Init(): InitLedger of %s  Failed ", val)
		}
	}

	err = stub.PutState("_dispatchOrderindex", []byte(dispatchOrderIndexstr))
	if err != nil {
		return nil, err
	}

	fmt.Println("Init() Initialization Complete  : ", args)
	return []byte("Init(): Initialization Complete"), nil
}

func InitLedger(stub shim.ChaincodeStubInterface, tableName string) error {

	// Generic Table Creation Function - requires Table Name and Table Key Entry
	// Create Table - Get number of Keys the tables supports
	// This version assumes all Keys are String and the Data is Bytes

	nKeys := GetNumberOfKeys(tableName)
	if nKeys < 1 {
		fmt.Println("Atleast 1 Key must be provided \n")
		fmt.Println("Auction_Application: Failed creating Table ", tableName)
		return errors.New("Auction_Application: Failed creating Table " + tableName)
	}

	var columnDefsForTbl []*shim.ColumnDefinition

	for i := 0; i < nKeys; i++ {
		columnDef := shim.ColumnDefinition{Name: "keyName" + strconv.Itoa(i), Type: shim.ColumnDefinition_STRING, Key: true}
		columnDefsForTbl = append(columnDefsForTbl, &columnDef)
	}

	columnLastTblDef := shim.ColumnDefinition{Name: "Details", Type: shim.ColumnDefinition_BYTES, Key: false}
	columnDefsForTbl = append(columnDefsForTbl, &columnLastTblDef)

	// Create the Table (Nil is returned if the Table exists or if the table is created successfully
	err := stub.CreateTable(tableName, columnDefsForTbl)

	if err != nil {
		fmt.Println("Auction_Application: Failed creating Table ", tableName)
		return errors.New("Auction_Application: Failed creating Table " + tableName)
	}

	return err
}

// Invoke is our entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "createDispatchOrder" {
		return t.createDispatchOrder(stub, args)
	}
	fmt.Println("invoke did not find func: " + function) //error

	return nil, errors.New("Received unknown function invocation: " + function)
}

// Query queries the hyperledger
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "readState" { //read a variable
		return t.readState(stub, args)
	}
	if function == "keys" {
		return t.getAllKeys(stub, args)
	}
	fmt.Println("query did not find func: " + function) //error

	return nil, errors.New("Received unknown function query " + function)
}

// read function return value
func (t *SimpleChaincode) readState(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}
	fmt.Println("valAsBytes", valAsbytes)

	obj, err := JSONtoDO(valAsbytes)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to conver into object" + name + "\"}"
		return nil, errors.New(jsonResp)
	}
	fmt.Println("json object trying to convert is", obj)

	return valAsbytes, nil
}

func (t *SimpleChaincode) getAllKeys(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) < 2 {
		return nil, errors.New("put operation must include two arguments, a key and value")
	}

	startKey := args[0]
	endKey := args[1]

	keysIter, err := stub.RangeQueryState(startKey, endKey)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("keys operation failed. Error accessing state: %s", err))
	}
	defer keysIter.Close()
	var keys []string
	for keysIter.HasNext() {
		response, _, iterErr := keysIter.Next()
		if iterErr != nil {
			return nil, errors.New(fmt.Sprintf("keys operation failed. Error accessing state: %s", err))
		}
		keys = append(keys, response)
	}

	for key, value := range keys {
		fmt.Printf("key %d contains %s\n", key, value)
	}

	jsonKeys, err := json.Marshal(keys)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("keys operation failed. Error accessing state: %s", err))
	}

	return jsonKeys, nil
}

func (t *SimpleChaincode) createDispatchOrder(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//convert the arguments into an Diapatch order Object
	dispatchObject, err := CreateDispatchOrderObject(args[0:])
	if err != nil {
		fmt.Println("createDispatchOrder(): Cannot create dispatch object ")
		return nil, errors.New("createDispatchOrder(): Cannot create dipatch object")
	}

	// check if the DispatchOrder already exists
	contractAsBytes, err := stub.GetState(dispatchObject.dispatchOrderId)
	if err != nil {
		fmt.Println("createDispatchOrder() : failed to get contract")
		return nil, errors.New("Failed to get dispatchOrder")
	}
	if contractAsBytes != nil {
		fmt.Println("initContract() : contract already exists for ", dispatchObject.dispatchOrderId)
		jsonResp := "{\"Error\":\"Failed - contract already exists " + dispatchObject.dispatchOrderId + "\"}"
		return nil, errors.New(jsonResp)
	}
	fmt.Println("dispatchObject is ", dispatchObject)
	buff, err := doToJSON(dispatchObject)
	if err != nil {
		errorStr := "initContract() : Failed Cannot create object buffer for write : " + args[0]
		fmt.Println(errorStr)
		return nil, errors.New(errorStr)
	}
	fmt.Println("createDispatchOrder() : buffer", buff)
	err = stub.PutState(args[0], buff)
	if err != nil {
		fmt.Println("initContract() : write error while inserting record\n")
		return nil, errors.New("initContract() : write error while inserting record : " + err.Error())
	}

	// make an entry into transaction history table

	return nil, nil
}

// CreateContractObject creates an contract
func CreateDispatchOrderObject(args []string) (DispatchOrderObject, error) {
	// S001 LHTMO bosch
	var err error
	var myDispatchOrder DispatchOrderObject

	// Check there are 31 Arguments provided as per the the struct, time is computed
	if len(args) != 3 {
		fmt.Println("CreateDispatchOrderObject(): Incorrect number of arguments. Expecting 31 ")
		return myDispatchOrder, errors.New("CreateDispatchOrderObject(): Incorrect number of arguments. Expecting 31 ")
	}

	//check whether the dispatch order already exists
	myDispatchOrder = DispatchOrderObject{args[0], strconv.Itoa(STATE_OBD_REQUEST_CREATED), args[2], time.Now().Format("20060102150405")}
	if err != nil {
		fmt.Println(err)
		return myDispatchOrder, err
	}
	fmt.Println("CreateDispatchOrderObject(): dispatch Object created: ", myDispatchOrder.dispatchOrderId, myDispatchOrder.stage, myDispatchOrder.customer, myDispatchOrder.timeStamp)
	return myDispatchOrder, nil
}

// doToJSON Converts an dispatch Object to a JSON String
func doToJSON(c DispatchOrderObject) ([]byte, error) {
	cjson, err := json.Marshal(c)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println("dispatch object as bytes ", cjson)
	return cjson, nil
}

func JSONtoDO(data []byte) (DispatchOrderObject, error) {

	do := DispatchOrderObject{}
	err := json.Unmarshal([]byte(data), &do)
	if err != nil {
		fmt.Println("Unmarshal failed : ", err)
	}

	return do, err
}
