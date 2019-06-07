//////////////////////////////////////
//           Chaincode            ////
//////////////////////////////////////
/*
    Esse contrato inteligente tem como objetivo fazer a requisição, leitura e armazenamento
    de dados .json que estão alocados em um servidor local.

	@autor: Carlos Augusto R. de Oliveira.
	@data: Junho/2019
*/
package main

import (
	//the majority of the imports are trivial...

	"bytes"
	"encoding/json"
	"fmt"

	//these imports are for Hyperledger Fabric interface
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

/*
 * SmartContract define a estrutura base do chaincode. Todos os métodos são implementados para
 * retornar um SmartContract type.
 */
type SmartContract struct {
}

/*
 O Meter constitui nossa estrutura de valor-chave (ativo digital) e implementa
 um único registro para gerenciar a chave pública e medidas. Todas as transações de
 blockchain operam com esse tipo.
 IMPORTANTE: Todos os nomes dos arquivos devem começar com letra maiúscula.
*/
type Meter struct {
	PublicKey    string `json:"publickey"`
	PlainMeasure int64  `json:"plainmeasure"`
}

/*
 * Esse método é chamado quando o chaincode é iniciado.
 */
func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The invoke function is called on each transaction invoking the chaincode. It
 * follows a structure of switching calls, so each valid feature need to
 * have a proper entry-point.
 */
func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	// extract the function name and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	//implements a switch for each acceptable function

	if fn == "insertMeasurement" {
		//inserts a measurement which increases the meter consumption counter. The measurement
		return s.insertMeasurement(stub, args)

	} else if fn == "queryLedger" {
		//execute a CouchDB query, args must include query expression
		return s.queryLedger(stub, args)
	}

	//function fn not implemented, notify error
	return shim.Error("Chaincode do not support this function")
}

func (s *SmartContract) insertMeasurement(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	//validate args vector lenght
	if len(args) != 2 {
		return shim.Error("It was expected 2 parameter: <meter ID> <measurement>")
	}

	//gets the parameter associated with the meter ID and the incremental measurement
	meterid := args[0]

	//creates Meter struct to manipulate returned bytes
	MyMeter := Meter{}

	meterAsBytes, err := stub.GetState(meterid)

	if err != nil || meterAsBytes == nil {
		return shim.Error("Error on retrieving meter ID register")
	}

	//convert bytes into a Meter object
	json.Unmarshal(meterAsBytes, &MyMeter)

	//encapsulates meter back into the JSON structure
	newMeterAsBytes, _ := json.Marshal(MyMeter)

	//update meter state in the ledger
	stub.PutState(meterid, newMeterAsBytes)

	//loging...
	fmt.Println("Updating meter consumption:", MyMeter)

	//notify procedure success
	return shim.Success(nil)
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

func TestPaillier() {

}

/*
 * The main function starts up the chaincode in the container during instantiate
 */
func main() {

	////////////////////////////////////////////////////////
	// USE THIS BLOCK TO COMPILE THE CHAINCODE
	//if err := shim.Start(new(SmartContract)); err != nil {
	//	fmt.Printf("Error starting SmartContract chaincode: %s\n", err)
	//}
	////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////
	// USE THIS BLOCK TO PERFORM ANY TEST WITH PAILLIER
	TestPaillier()
	////////////////////////////////////////////////////////

}
