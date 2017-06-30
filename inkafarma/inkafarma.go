/*
  Inkafarma Smart Contract
  Adrian Pareja
*/
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/util"

	"crypto/rand"
	"encoding/hex"
)

const business string = "Inkafarma"
const walletContract string = "e966457f41cb31cb21dd1f69d7809f9b1ac9437f2daf913f8c94d0047694e4ccf6ccd298a526978af8fef9d942ef90e3cdfd4fd2acf7f2a35121f5540d23d6fd"

// UUID layout variants.
const (
	VariantNCS = iota
	VariantRFC4122
	VariantMicrosoft
	VariantFuture
)

// Used in string method conversion
const dash byte = '-'

// UUID representation compliant with specification
// described in RFC 4122.
type UUID [16]byte

//Wallet - Structure for products used in buy goods
type Wallet struct {
	Id        string    `json:"id"`
	Email     string  `json:"email"`
	Phone     string  `json:"phone"`
	Document  string `json:"document"`
	Password  string `json:"password"` 
	Amount    float64 `json:"amount"` 
}

//Response - Structure for response
type ResponseContract struct {
	Code        int32    `json:"code"`
	Response    string  `json:"response"` 
}

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func main() {
	fmt.Printf("Iniciandooo Contrato Cineplanet....")
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error Iniciando Cineplanet Smart Contract: %s", err)
	}
}

// Init reinicia los estados del ledger
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Número de Argumentos incorrecto. Se esperaba 1 argumento")
	}

	err := stub.PutState("coinBalance", []byte(args[0]))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Invoke Punto de entrada a cualquier función del ledger
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Vivanda invoke is running..FUNCTION:" + function)

	if function == "createwallet" {
		return t.createWallet(stub, args)
	} else if function == "buy" {
		return t.buy(stub, args)
	}
	fmt.Println("invoke no encuentra la funcion: " + function)

	return nil, errors.New("Funcion invocada desconocida: " + function)
}

// Query es nuestro punto de entrada de querys
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Vivanda query is running FUNCTION:" + function)

	// Manejar diferentes funciones
	if function == "getbalance" {
		return t.getBalance(stub, args)
	}
	fmt.Println("query no encuentra la funcion: " + function)

	return nil, errors.New("Funcion invocada desconocida: " + function)
}

// createWallet - invocar esta funcion para crear un wallet con saldo inicial
func (t *SimpleChaincode) createWallet(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Vivanda Call---Funcion createWallet---")
	
	if len(args) != 5 {
		return nil, errors.New("Numero incorrecto de argumentos.Se espera 5 para createWallet")
	}

	f := "createwallet"
	invokeArgs := util.ToChaincodeArgs(f, args[0], args[1], args[2], args[3], "123456", args[4])
	response, err := stub.InvokeChaincode(walletContract, invokeArgs)
	if err != nil {
		errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", err.Error())
		fmt.Printf(errStr)
		return nil, errors.New(errStr)
	}

	fmt.Printf("Invoke chaincode successful. Got response %s", string(response))

	return nil, nil
}

// createWallet - invocar esta funcion para crear un wallet con saldo inicial
func (t *SimpleChaincode) buy(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Vivanda Call---Funcion Buy---")
	
	if len(args) != 3 {
		return nil, errors.New("Numero incorrecto de argumentos.Se espera 3 para buy")
	}

	var change float64 = 3
	
	soles, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		errStr := fmt.Sprintf("Fallo convertir cadena a float: %s", err.Error())
		return nil, errors.New(errStr)
	}
	
	coins, err1 := strconv.ParseFloat(args[2], 64)
	if err1 != nil {
		errStr := fmt.Sprintf("Fallo convertir cadena a float: %s", err.Error())
		return nil, errors.New(errStr)
	}
	
	f := "getbalance"
	queryArgs := util.ToChaincodeArgs(f, args[0])
	responseQuery, err2 := stub.QueryChaincode(walletContract, queryArgs)
	if err2 != nil {
		errStr := fmt.Sprintf("Failed to query chaincode. Got error: %s", err.Error())
		return nil, errors.New(errStr)
	}

	fmt.Printf("Invoke chaincode successful. Got response %s", string(responseQuery))

	responseContract := ResponseContract{}
	err1 = json.Unmarshal(responseQuery, &responseContract)
	
	f = "debitbalance"
	
	//Compra soles y canje coins
	if soles > 0 && coins > 0 {
		
		coinBalance,_ := strconv.ParseFloat(responseContract.Response, 64)
		
		if  coinBalance <= coins {
			return nil,errors.New("El cliente no cuenta con coins suficientes")
		}
			
		coins = (soles * change) - coins
		if coins < 0 {
			coins = coins*-1
			f = "debitbalance"
		} else {
			f = "putbalance"
		}
	} else {
		if soles > 0 {  //Compra Soles
			f = "putbalance"					
			coins = soles * change
		}else if coins > 0 {  //Canje Coins
			coinBalance,_ := strconv.ParseFloat(responseContract.Response, 64)
			if coinBalance <= coins {
				return nil,errors.New("El cliente no cuenta con coins suficientes")
			}
			f = "debitbalance"
		}
	}

	invokeArgs := util.ToChaincodeArgs(f, args[0], business, strconv.FormatFloat(coins,'f',6,64))
	response, err4 := stub.InvokeChaincode(walletContract, invokeArgs)
	if err4 != nil {
		errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", err4.Error())
		fmt.Printf(errStr)
		return nil, errors.New(errStr)
	}

	fmt.Printf("Invoke chaincode successful. Got response %s", string(response))

	return nil, nil
}

func (t *SimpleChaincode) getBalance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Vivanda----getBalance() is running----")
	
	if len(args) != 1 {
		return nil, errors.New("Numero incorrecto de argumentos.Se espera 1 para getBalance")
	}

	f := "getbalance"
	invokeArgs := util.ToChaincodeArgs(f, args[0])
	response, err := stub.QueryChaincode(walletContract, invokeArgs)
	if err != nil {
		errStr := fmt.Sprintf("Failed to query chaincode. Got error: %s", err.Error())
		fmt.Printf(errStr)
		return nil, errors.New(errStr)
	}

	fmt.Printf("Invoke chaincode successful. Got response %s", string(response))

	responseContract := ResponseContract{}
	err1 := json.Unmarshal(response, &responseContract)
	
	fmt.Println(responseContract.Response)
	if err1 != nil {
		fmt.Println("Error parseando a Json" + args[0])
		return nil, errors.New("Error retrieving Balance" + args[0])
	}
	
	return []byte(fmt.Sprintf(`{"code":0,"response":"%s"}`,responseContract.Response)), nil
}

func safeRandom(dest []byte) {
	if _, err := rand.Read(dest); err != nil {
		panic(err)
	}
}

// SetVersion sets version bits.
func (u *UUID) SetVersion(v byte) {
	u[6] = (u[6] & 0x0f) | (v << 4)
}

// SetVariant sets variant bits as described in RFC 4122.
func (u *UUID) SetVariant() {
	u[8] = (u[8] & 0xbf) | 0x80
}

func (u UUID) Version() uint {
	return uint(u[6] >> 4)
}

// Variant returns UUID layout variant.
func (u UUID) Variant() uint {
	switch {
	case (u[8] & 0x80) == 0x00:
		return VariantNCS
	case (u[8]&0xc0)|0x80 == 0x80:
		return VariantRFC4122
	case (u[8]&0xe0)|0xc0 == 0xc0:
		return VariantMicrosoft
	}
	return VariantFuture
}

// Returns canonical string representation of UUID:
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func (u UUID) String() string {
	buf := make([]byte, 36)

	hex.Encode(buf[0:8], u[0:4])
	buf[8] = dash
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = dash
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = dash
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = dash
	hex.Encode(buf[24:], u[10:])

	return string(buf)
}

func NewV4() UUID {
	u := UUID{}
	safeRandom(u[:])
	u.SetVersion(4)
	u.SetVariant()

	return u
}
