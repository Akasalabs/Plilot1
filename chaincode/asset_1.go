package main

import (
	"errors"
	"fmt"
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

// Assest struct
type Assest struct {
	Serialno string `json:"serialno"`
	Partno   string `json:"partno"`
	Owner    string `json:"owner"`
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

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

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "init_assset" {
		return t.init_assset(stub, args)
	} else if function == "write_owner" {
		return t.write_owner(stub, args)
	}
	fmt.Println("invoke did not find func: " + function) //error

	return nil, errors.New("Received unknown function invocation: " + function)
}

func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function) //error

	return nil, errors.New("Received unknown function query")
}

func (t *SimpleChaincode) init_assset(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	// S001 LHTMO bosch
	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	serialno := args[0]
	partno := args[1]
	owner := args[2]

	//check assest already exist
	/*assestAsBytes, err := stub.GetState(serialno)
	if err != nil {
		return nil, errors.New("Failed to get assest")
	}
	res := Assest{}
	json.Unmarshal (assestAsBytes, &res)
	if res.Serialno == serialno {
		fmt.Println("This assest arleady exists: " + serialno)
		fmt.Println(res);
		return nil, errors.New("This assest arleady exists")
	}*/

	str := `{"serialno": "` + serialno + `", "partno": "` + partno + `", "owner": "` + owner + `"}`
	err = stub.PutState(serialno, []byte(str))
	if err != nil {
		return nil, err
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
