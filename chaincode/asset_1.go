package main

import (
	"errors"
	"fmt"
	"strconv"
	//"strconv"
	"encoding/json"
	"time"
	//"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var dispatchOrderIndexstr = "_dispatchOrderindex"

//==============================================================================================================================
//	 Status types - contract lifecycle is broken down into 5 statuses, this is part of the business logic to determine what can
//					be done to the vehicle at points in it's lifecycle
//==============================================================================================================================
const STATE_OBD_REQUEST_CREATED = 0
const STATE_READY_FOR_DISPATCH = 1
const STATE_ARRIVAL_OF_TRANSPORTER = 2
const STATE_READY_FOR_SHIPMENT = 3
const STATE_IN_TRANSIT = 4
const STATE_SHIPMENT_DELIVERED = 5
const STATE_AMENDED = 6
const STATE_DROPPED = 7

const SELLER = "seller"
const TRANSPORTER = "transporter"
const BUYER = "lease_company"

type DispatchOrderObject struct {
	dispatchOrderId                string
	stage                          int
	customer                       string
	transporter                    string
	seller                         string
	assetIDs                       string
	asnNumber                      string
	source                         string
	shipmentType                   string
	contractType                   string
	deliveryTerm                   string
	dispatchDate                   string
	transporterRef                 string
	loadingType                    string
	vehicleType                    string
	weight                         string
	consignment                    string
	quantity                       string
	partNumber                     string
	partName                       string
	orderRefNum                    string
	createdOn                      string
	documentID1                    string
	documentID2                    string
	documentID3                    string
	documentID4                    string
	dropDescription                string
	deliverydescription            string
	inTransitDisptachOfficerSigned string
	inTransitTransporterSigned     string
	transactionDescription         string
	timeStamp                      string // This is the time stamp
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
	if function == "keys" {
		return t.getAllKeys(stub, args)
	} else if function == "readContract" { //read a contract
		return t.readContract(stub, args)
	}

	fmt.Println("query did not find func: " + function) //error
	return nil, errors.New("Received unknown function query " + function)
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

	buff, err := doToJSON(dispatchObject)
	if err != nil {
		errorStr := "initContract() : Failed Cannot create object buffer for write : " + args[1]
		fmt.Println(errorStr)
		return nil, errors.New(errorStr)
	}
	err = stub.PutState(dispatchObject.dispatchOrderId, buff)
	if err != nil {
		fmt.Println("initContract() : write error while inserting record\n")
		return nil, errors.New("initContract() : write error while inserting record : " + err.Error())
	}

	// make an entry into transaction history table

	return nil, nil
}

// read function return value
func (t *SimpleChaincode) readContract(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
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

// CreateContractObject creates an contract
func CreateDispatchOrderObject(args []string) (DispatchOrderObject, error) {
	// S001 LHTMO bosch
	var err error
	var myDispatchOrder DispatchOrderObject

	// Check there are 31 Arguments provided as per the the struct, time is computed
	if len(args) != 31 {
		fmt.Println("CreateDispatchOrderObject(): Incorrect number of arguments. Expecting 31 ")
		return myDispatchOrder, errors.New("CreateDispatchOrderObject(): Incorrect number of arguments. Expecting 31 ")
	}

	//check whether the dispatch order already exists
	myDispatchOrder = DispatchOrderObject{args[0], STATE_OBD_REQUEST_CREATED, args[2], args[3], args[4], args[5], args[6], args[7], args[8], args[9], args[10], args[11], args[12], args[13], args[14], args[15], args[16], args[17], args[18], args[19], args[20], args[21], args[22], args[23], args[24], args[25], args[26], args[27], args[28], args[29], args[30], time.Now().Format("20060102150405")}
	if err != nil {
		fmt.Println(err)
		return myDispatchOrder, err
	}
	//fmt.Println("CreateContractObject(): Contract Object created: ", myContract.Contractid, myContract.Stage, myContract.Buyer, myContract.Transporter, myContract.Seller, myContract.AssetID, myContract.DocumentID, time.Now().Format("20060102150405"))
	return myDispatchOrder, nil
}

// doToJSON Converts an dispatch Object to a JSON String
func doToJSON(c DispatchOrderObject) ([]byte, error) {
	cjson, err := json.Marshal(c)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return cjson, nil
}
