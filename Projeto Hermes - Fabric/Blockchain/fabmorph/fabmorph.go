/////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////           Chaincode            /////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////

/*
	This chaincode aims to make the requisition, reading and storage
	of data .json that will be sent by another application, written in pyhton, via http. For
	the communication of the application written in python with this chaincode, was used
	the Pyhton SDK.

	@author: Carlos Augusto R. de Oliveira.
	@date: July/2019
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	//These imports are for the hyperledger interface.
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

/*
	SmartContract defines the base structure of the chaincode. All methods are implemented to
	return a SmartContract type.
*/
type SmartContract struct {
}

/*
	Data is our key-value structure (digital asset) and implements a single record to manage
	the measures. All transactions of blockchain operate with this type.
	IMPORTANT: All file names must begin with a capital letter.
*/
type Data struct {
	ID                int64   `json:"id"`
	Sensor            string  `json:"sensor_model"`
	Value             float32 `json:"value"`
	Timestamp         float32 `json:"timestamp"`
	Location          string  `json:"location"`
	Physical_quantity string  `json:"physical_quantity"`
}

/*
	This method is called when the chaincode is started.
*/
func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
	The invoke function is called on each transaction invoking the chaincode. It follows a structure
	of switching calls, so each valid feature need to have a proper entry-point.
*/
func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	// extract the function name and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	//implements a switch for each acceptable function

	if fn == "insertMeasurement" {
		//inserts a measurement which increases the meter consumption counter.
		return s.insertMeasurement(stub, args)

	} else if fn == "queryLedger" {
		//execute a CouchDB query, args must include query expression
		return s.queryLedger(stub, args)

	} else if fn == "getConsumption" {
		//retrieves the accumulated consumption
		return s.getConsumption(stub, args)

	} else if fn == "queryHistory" {
		//look for a specific fill up record and brings its changing history
		return s.queryHistory(stub, args)
	}

	//function fn not implemented, notify error
	return shim.Error("Chaincode do not support this function")
}

func (s *SmartContract) insertMeasurement(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	//validate args vector lenght
	if len(args) != 2 {
		return shim.Error("It was expected 2 parameter: <meter ID> <measurement>")
	}

	//gets the parameter associated with the meter ID
	meterid := args[0]

	//try to convert the informed measurement into the format []byte, required by Gomorph
	measurement, err := json.Marshal(args[1])

	if err != nil {
		panic(err)
	}

	//check if we have success
	if measurement == nil {
		//measurement is not a proper number
		return shim.Error("Error on veryfing measurement, it is not a proper input, deu ruim")
	}

	MyMeter := Data{}

	//convert bytes into a Data object
	json.Unmarshal(measurement, &MyMeter)

	//update Data state in the ledger
	stub.PutState(meterid, measurement)

	//loging...
	fmt.Println("Updating Data consumption:", measurement)

	//notify procedure success
	return shim.Success(nil)
}

func (s *SmartContract) queryHistory(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	//validate args vector lenght
	if len(args) != 1 {
		return shim.Error("It was expected 1 parameter: <key>")
	}

	historyIer, err := stub.GetHistoryForKey(args[0])

	//verifies if the history exists
	if err != nil {
		//fmt.Println(errMsg)
		return shim.Error("Fail on getting ledger history")
	}

	// buffer is a JSON array containing records
	var buffer bytes.Buffer
	var counter = 0
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for historyIer.HasNext() {
		//increments iterator
		queryResponse, err := historyIer.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		//generates a formated result
		buffer.WriteString("{\"Value\":")
		buffer.WriteString("\"")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("\"")
		buffer.WriteString(", \"Counter\":")
		buffer.WriteString(strconv.Itoa(counter))
		//buffer.WriteString(queryResponse.Timestamp)
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true

		//increases counter
		counter++
	}
	buffer.WriteString("]")
	historyIer.Close()

	//loging...
	fmt.Printf("Consulting ledger history, found %d\n records", counter)

	//notify procedure success
	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) queryLedger(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	//validate args vector lenght
	if len(args) != 1 {
		return shim.Error("It was expected 1 parameter: <query string>")
	}

	//using auxiliar variable
	queryString := args[0]

	//loging...
	fmt.Printf("Executing the following query: %s\n", queryString)

	//try to execute query and obtain records iterator
	resultsIterator, err := stub.GetQueryResult(queryString)
	//test if iterator is valid
	if err != nil {
		return shim.Error(err.Error())
	}
	//defer iterator closes at the end of the function
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		//increments iterator
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		//generates a formated result
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")
		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	//loging...
	fmt.Printf("Obtained the following fill up records: %s\n", buffer.String())

	//notify procedure success
	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) getConsumption(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	//validate args vector lenght
	if len(args) != 1 {
		return shim.Error("It was expected 1 parameter: <meter ID>")
	}

	//gets the parameter associated with the meter ID and the incremental measurement
	meterid := args[0]

	//retrive Data record
	meterAsBytes, err := stub.GetState(meterid)

	//test if we receive a valid meter ID
	if err != nil || meterAsBytes == nil {
		return shim.Error("Error on retrieving meter ID register")
	}

	//return payload with bytes related to the meter state
	return shim.Success(meterAsBytes)
}

/*
 * The main function starts up the chaincode in the container during instantiate
 */
func main() {
	////////////////////////////////////////////////////////
	////// USE THIS BLOCK TO COMPILE THE CHAINCODE /////////
	if err := shim.Start(new(SmartContract)); err != nil {
		fmt.Printf("Error starting SmartContract chaincode: %s\n", err)
	}
}
