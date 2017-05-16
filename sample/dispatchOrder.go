package main

import (
	"errors"
	"fmt"
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

var dispatchOrderIndexstr = "_dispatchOrderindex"

// AssetObject struct
type DispatchOrderObject struct {
	DispatchOrderID string `json:"dispatchOrderId"`
	Stage           string `json:"stage"`
	Customer        string `json:"customer"`
	TimeStamp       string `json:"timeStamp"` // This is the time stamp
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes the chain
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	var err error
	err = stub.PutState("hello world!", []byte(dispatchOrderIndexstr))
	if err != nil {
		return nil, err
	}

	return nil, nil
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

func (t *SimpleChaincode) createDispatchOrder(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//convert the arguments into an asset Object
	dispatchObject, err := CreateDispatchObject(args[0:])
	if err != nil {
		fmt.Println("createDispatchOrder(): Cannot create asset object ")
		return nil, errors.New("createDispatchOrder(): Cannot create asset object")
	}

	// check if the asset already exists
	dispatchObjectAsBytes, err := stub.GetState(dispatchObject.DispatchOrderID)
	if err != nil {
		fmt.Println("createDispatchOrder() : failed to get dispatch order")
		return nil, errors.New("Failed to get dispatch order")
	}
	if dispatchObjectAsBytes != nil {
		fmt.Println("createDispatchOrder() : dispatch order exists ", dispatchObject.DispatchOrderID)
		jsonResp := "{\"Error\":\"Failed - dispatch order exists " + dispatchObject.DispatchOrderID + "\"}"
		return nil, errors.New(jsonResp)
	}

	buff, err := DOtoJSON(dispatchObject)
	if err != nil {
		errorStr := "createDispatchOrder() : Failed Cannot create object buffer for write : " + args[0]
		fmt.Println(errorStr)
		return nil, errors.New(errorStr)
	}
	err = stub.PutState(args[0], buff)
	if err != nil {
		fmt.Println("createDispatchOrder() : write error while inserting record\n ")
		return nil, errors.New("createDispatchOrder() : write error while inserting record : " + err.Error())
	}
	return nil, nil
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

// CreateAssetObject creates an asset
func CreateDispatchObject(args []string) (DispatchOrderObject, error) {
	// S001 LHTMO bosch
	var err error
	var myDispatchObject DispatchOrderObject

	// Check there are 3 Arguments provided as per the the struct
	if len(args) != 3 {
		fmt.Println("CreateAssetObject(): Incorrect number of arguments. Expecting 3 ")
		return myDispatchObject, errors.New("CreateDispatchObject(): Incorrect number of arguments. Expecting 3 ")
	}

	// Validate Serialno is an integer

	myDispatchObject = DispatchOrderObject{args[0], args[1], args[2], time.Now().Format("20060102150405")}
	if err != nil {
		fmt.Println("CreateDispatchObject(): Dispatch order object create failed! ")
		return myDispatchObject, errors.New("DispatchOrderObject(): Dispatch order object create failed!. ")
	}

	fmt.Println("CreateDispatchObject(): Dispatch Object created: ", myDispatchObject.DispatchOrderID, myDispatchObject.Stage, myDispatchObject.Customer, myDispatchObject.TimeStamp)
	return myDispatchObject, nil
}

// ARtoJSON Converts an Asset Object to a JSON String
func DOtoJSON(do DispatchOrderObject) ([]byte, error) {

	ajson, err := json.Marshal(do)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println("DOtoJSON() :", ajson)
	return ajson, nil
}
