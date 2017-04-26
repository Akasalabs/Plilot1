package main

import (
	"errors"
	"fmt"
	"strconv"
	//"strconv"
	"encoding/json"
	//"time"
	//"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var assestIndexstr = "_assestindex"

// AssetObject struct
type AssetObject struct {
	Serialno string
	Partno   string
	Owner    string
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
	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)
	err = stub.PutState(assestIndexstr, jsonAsBytes)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Invoke is our entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "initAssset" {
		return t.initAssset(stub, args)
	}
	fmt.Println("invoke did not find func: " + function) //error

	return nil, errors.New("Received unknown function invocation: " + function)
}

// Query queries the hyperledger
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function) //error

	return nil, errors.New("Received unknown function query " + function)
}

func (t *SimpleChaincode) initAssset(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//convert the arguments into an asset Object
	AssetObject, err := CreateAssetObject(args[0:])
	if err != nil {
		fmt.Println("initAsset(): Cannot create asset object ")
		return nil, errors.New("initAsset(): Cannot create asset object")
	}

	// check if the asset already exists
	assestAsBytes, err := stub.GetState(AssetObject.Serialno)
	if err != nil {
		fmt.Println("initAssset() : failed to get asset")
		return nil, errors.New("Failed to get asset")
	}
	if assestAsBytes != nil {
		fmt.Println("initAssset() : Asset already exists ", AssetObject.Serialno)
		jsonResp := "{\"Error\":\"Failed - Asset already exists " + AssetObject.Serialno + "\"}"
		return nil, errors.New(jsonResp)
	}

	buff, err := ARtoJSON(AssetObject)
	if err != nil {
		errorStr := "initAssset() : Failed Cannot create object buffer for write : " + args[1]
		fmt.Println(errorStr)
		return nil, errors.New(errorStr)
	} else {
		err = stub.PutState(args[0], buff)
		if err != nil {
			fmt.Println("initAssset() : write error while inserting record\n")
			return nil, errors.New("initAssset() : write error while inserting record : " + err.Error())
		}
	}
	return nil, nil
}

// read function return value
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
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

	return valAsbytes, nil
}

// CreateAssetObject creates an asset
func CreateAssetObject(args []string) (AssetObject, error) {
	// S001 LHTMO bosch
	var err error
	var myAsset AssetObject

	// Check there are 3 Arguments provided as per the the struct
	if len(args) != 3 {
		fmt.Println("CreateAssetObject(): Incorrect number of arguments. Expecting 3 ")
		return myAsset, errors.New("CreateAssetObject(): Incorrect number of arguments. Expecting 3 ")
	}

	// Validate Serialno is an integer

	_, err = strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("CreateAssetObject(): SerialNo should be an integer create failed! ")
		return myAsset, errors.New("CreateAssetbject(): SerialNo should be an integer create failed. ")
	}

	myAsset = AssetObject{args[0], args[1], args[2]}

	fmt.Println("CreateAssetObject(): Asset Object created: ", myAsset.Serialno, myAsset.Partno, myAsset.Owner)
	return myAsset, nil
}

// ARtoJSON Converts an Asset Object to a JSON String
func ARtoJSON(ast AssetObject) ([]byte, error) {

	ajson, err := json.Marshal(ast)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return ajson, nil
}
