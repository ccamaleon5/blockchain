/*
  Cineplanet Smart Contract
  Adrian Pareja
*/
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/util"

	"crypto/rand"
	"encoding/hex"
)

const business string = "Cineplanet"
const walletContract string = "095fb94438b9baa981095b175e7d93b20a3b7fdb91c825bf3cc7b85a5265427c3ad5cdde04cdac73216d6c9e9779fb050efc6404ef0223ebef8f366b39ff299d"
const change float64 = 1

const (
	tableColumn     = "CanjesCineplanet"
	columnTime      = "Time"
	columnAccountID = "Account"
	columnAmount    = "Amount"
	columnType      = "Type"
)

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
	Id       string  `json:"id"`
	Email    string  `json:"email"`
	Phone    string  `json:"phone"`
	Document string  `json:"document"`
	Password string  `json:"password"`
	Amount   float64 `json:"amount"`
}

//Movimiento - Structure for movements
type Movement struct {
	Time     int64   `json:"time"`
	WalletId string  `json:"walletid"`
	Amount   float64 `json:"amount"`
	Type     string  `json:"type"`
}

//Balance - Structure for balance
type Balance struct{
	Business string `json:"business"`
	Total float64 `json:"total"`
	Exchange float64 `json:"exchange"`
	Send float64 `json:"send"`   
}

//Response - Structure for response
type ResponseContract struct {
	Code     int32  `json:"code"`
	Balance string `json:"balance"`
	Limit string `json:"limit"`
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
	
	amt, err := strconv.ParseFloat(args[0], 64)

	if err != nil {
		fmt.Println("Error Float parsing")
		return nil, errors.New("Error marshaling wallet")
	}
	
	//Adquirir coins iniciales
	f := "debittotalcoin"
	invokeArgs := util.ToChaincodeArgs(f, args[0])
	response, err := stub.InvokeChaincode(walletContract, invokeArgs)
	
	if err != nil {
		errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", err.Error())
		fmt.Printf(errStr)
		return nil, errors.New(errStr)
	}

	fmt.Printf("Invoke chaincode successful. Got response %s", string(response))
	
	balance := Balance{
		Business: "Cineplanet",
		Total:    amt,
		Exchange: 0,
		Send:     0,
	}

	bytes, err1 := json.Marshal(balance)
	if err1 != nil {
		fmt.Println("Error marshaling wallet")
		return nil, errors.New("Error marshaling wallet")
	}

	err = stub.PutState("coinBalance", bytes)
	if err != nil {
		fmt.Println("Error creando el balance inicial del negocio")
		return nil, err
	}
	
	stub.CreateTable(tableColumn, []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: columnAccountID, Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: columnTime, Type: shim.ColumnDefinition_INT64, Key: true},
		&shim.ColumnDefinition{Name: columnAmount, Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: columnType, Type: shim.ColumnDefinition_STRING, Key: false},
	})

	return nil, nil
}

// Invoke Punto de entrada a cualquier función del ledger
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Cineplanet invoke is running..FUNCTION:" + function)

	if function == "createwallet" {
		return t.createWallet(stub, args)
	} else { 
		if function == "buy" {
			return t.buy(stub, args)
		} else if function == "getcoins" {
			return t.getCoins(stub, args)
		}
	}
	fmt.Println("invoke no encuentra la funcion: " + function)

	return nil, errors.New("Funcion invocada desconocida: " + function)
}

// Query es nuestro punto de entrada de querys
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Cineplanet query is running FUNCTION:" + function)

	// Manejar diferentes funciones
	if function == "getbalance" {
		return t.getBalance(stub, args)
	} else{ 
		if function == "gettotalcoin" {
			return t.getTotalCoin(stub, args)
		} else if function == "getmovimientos" {
			return t.getMovimientos(stub, args)
		}
	}
	
	fmt.Println("query no encuentra la funcion: " + function)

	return nil, errors.New("Funcion invocada desconocida: " + function)
}

// createWallet - invocar esta funcion para crear un wallet con saldo inicial
func (t *SimpleChaincode) createWallet(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Cineplanet Call---Funcion createWallet---")

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

// createWallet - invocar esta funcion para compras y canjes de coins
func (t *SimpleChaincode) buy(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Cineplanet Call---Funcion Buy---")

	if len(args) != 3 {
		return nil, errors.New("Numero incorrecto de argumentos.Se espera 3 para buy")
	}

	solesTotal, err := strconv.ParseFloat(args[1], 64)
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

	solesSubtotal := (solesTotal * change) - coins //Cambiando a Coins

	//Compra soles subtotal y canje coins
	if solesSubtotal > 0 && coins > 0 {
		coinBalance, _ := strconv.ParseFloat(responseContract.Balance, 64)

		if coinBalance <= coins {
			return nil, errors.New("El cliente no cuenta con coins suficientes")
		}

		//Debitar Coins Usuario
		invokeArgs2 := util.ToChaincodeArgs(f, args[0], business, strconv.FormatFloat(coins, 'f', 6, 64))
		response2, err6 := stub.InvokeChaincode(walletContract, invokeArgs2)
		if err6 != nil {
			errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", err6.Error())
			fmt.Printf(errStr)
			return nil, errors.New(errStr)
		}

		fmt.Printf("Invoke chaincode successful. Got response %s", string(response2))
		
		insertRow(stub,strconv.FormatFloat(coins, 'f', 6, 64),"C")

		//Cargar Coins Usuario
		coins = solesSubtotal //- coins
		if coins < 0 {
			coins = coins * -1
			f = "debitbalance"
		} else {
			f = "putbalance"
		}
	} else {
		if solesSubtotal > 0 { //Compra Soles
			f = "putbalance"
			coins = solesTotal * change
		} else if coins > 0 { //Canje Coins
			coinBalance, _ := strconv.ParseFloat(responseContract.Balance, 64)
			if coinBalance <= coins {
				return nil, errors.New("El cliente no cuenta con coins suficientes")
			}
			f = "debitbalance"
		}
	}

	invokeArgs := util.ToChaincodeArgs(f, args[0], business, strconv.FormatFloat(coins, 'f', 6, 64))
	response, err4 := stub.InvokeChaincode(walletContract, invokeArgs)
	if err4 != nil {
		errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", err4.Error())
		fmt.Printf(errStr)
		return nil, errors.New(errStr)
	}

	fmt.Printf("Invoke chaincode successful. Got response %s", string(response))
	
	if f == "putbalance" {
		insertRow(stub,strconv.FormatFloat(coins, 'f', 6, 64),"D")
	} else{
		insertRow(stub,strconv.FormatFloat(coins, 'f', 6, 64),"C")
	}
	
	coins,_ = strconv.ParseFloat(args[2], 64)
	
	if false == updateBalance(stub, coins, solesSubtotal) {
		errStr := fmt.Sprintf("Failed update balance")
		return nil, errors.New(errStr)
	} 

	return nil, nil
}

func (t *SimpleChaincode) getBalance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Cineplanet----getBalance() is running----")

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

	fmt.Println(responseContract.Balance)
	if err1 != nil {
		fmt.Println("Error parseando a Json" + args[0])
		return nil, errors.New("Error retrieving Balance" + args[0])
	}

	return []byte(fmt.Sprintf(`{"code":0,"balance":"%s","limit":"%s"}`, responseContract.Balance,responseContract.Limit)), nil
}

func (t *SimpleChaincode) getTotalCoin(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----getTotalCoin() is running----")

	bytesWallet1, err1 := stub.GetState("coinBalance")

	balance := Balance{}
	err := json.Unmarshal(bytesWallet1, &balance)
	if err != nil {
		fmt.Println("Error TotalCoin parsing")
		return nil, errors.New("Error marshaling totalBalance")
	}
	

	fmt.Println(balance)
	if err1 != nil {
		fmt.Println("Error retrieving balance")
		return nil, errors.New("Error retrieving coinBalance")
	}

	return []byte(fmt.Sprintf(`{"business":"%s","balance":%v,"spend":%v,"sents":%v}`, balance.Business,balance.Total,balance.Exchange,balance.Send)), nil
}

//Obtener los movimientos de los coins en Cineplanet
func (t *SimpleChaincode) getMovimientos(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----getMovimientos() is running----")

	if len(args) != 1 {
		return nil, errors.New("Incorrecto numero de argumentos. Se esperaba 1")
	}

	walletId := args[0] // wallet id
	fmt.Println("Business id is ")
	fmt.Println(walletId)
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: walletId}}
	columns = append(columns, col1)

	rowChannel, err := stub.GetRows(tableColumn, columns)
	if err != nil {
		return nil, fmt.Errorf("getRowTableOne operation failed. %s", err)
	}

	movimientos := []Movement{}
	for {
		select {
		case row, ok := <-rowChannel:
			if !ok {
				rowChannel = nil
			} else {
				columnas := row.GetColumns()
				amountRow, _ := strconv.ParseFloat(columnas[2].GetString_(), 64)
				movimiento := Movement{Time: columnas[1].GetInt64(), WalletId: columnas[0].GetString_(), Amount: amountRow, Type: columnas[3].GetString_()}

				movimientos = append(movimientos, movimiento)
			}
		}
		if rowChannel == nil {
			break
		}
	}

	jsonRows, err := json.Marshal(movimientos)
	if err != nil {
		return nil, fmt.Errorf("getRows Movimientos operation failed. Error marshaling JSON: %s", err)
	}

	return jsonRows, nil
}

//Insertar Row de Retorno y Entrega de Coins al Usuario
func insertRow(stub shim.ChaincodeStubInterface, amount string, tipo string) bool {
	
	//Insertar Row de Retorno de Coins al Negocio
		a := makeTimestamp()

		fmt.Printf("Time: %d \n", a)
	
		var columns []*shim.Column
		col0 := shim.Column{Value: &shim.Column_String_{String_: business}}
		col1 := shim.Column{Value: &shim.Column_Int64{Int64: a}}
		col2 := shim.Column{Value: &shim.Column_String_{String_: amount}}
		col3 := shim.Column{Value: &shim.Column_String_{String_: tipo}}
		
		columns = append(columns, &col0)
		columns = append(columns, &col1)
		columns = append(columns, &col2)
		columns = append(columns, &col3)
	
		row := shim.Row{Columns: columns}
		ok, err := stub.InsertRow(tableColumn, row)
		if err != nil {
			return false
		}
		if !ok {
			return false
		}
	return true;
}

//Cambiar el balance de coins del negocio
func updateBalance(stub shim.ChaincodeStubInterface, coins float64, subtotalSoles float64) bool {
	bytesWallet1, err1 := stub.GetState("coinBalance")

	balance := Balance{}
	err := json.Unmarshal(bytesWallet1, &balance)

	fmt.Println(balance)
	if err1 != nil {
		fmt.Println("Error retrieving balance")
		return false
	}

	//amt, err := strconv.ParseFloat(args[2], 64)

	balance.Exchange = balance.Exchange + coins
	balance.Send = balance.Send + subtotalSoles 
	balance.Total = balance.Total - subtotalSoles

	balanceJSONasBytes, _ := json.Marshal(balance)
	err = stub.PutState("coinBalance", balanceJSONasBytes) //rewrite the wallet

	if err != nil {
		return false
	}

	return true
}

//Obtener los movimientos de los coins en Vivanda
func (t *SimpleChaincode) getCoins(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----getCoins() is running----")

	if len(args) != 1 {
		return nil, errors.New("Incorrecto numero de argumentos. Se esperaba 1")
	}

	amt, err := strconv.ParseFloat(args[0], 64)

	if err != nil {
		fmt.Println("Error Float parsing")
		return nil, errors.New("Error marshaling wallet")
	}
	
	//Adquirir coins adicionales
	f := "debittotalcoin"
	invokeArgs := util.ToChaincodeArgs(f, args[0])
	response, err1 := stub.InvokeChaincode(walletContract, invokeArgs)
	
	if err1 != nil {
		errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", err.Error())
		fmt.Printf(errStr)
		return nil, errors.New(errStr)
	}

	fmt.Printf("Invoke chaincode successful. Got response %s", string(response))
	
	//Cambiar el balance
	bytesWallet1, err2 := stub.GetState("coinBalance")

	balance := Balance{}
	err3 := json.Unmarshal(bytesWallet1, &balance)
	if err3 != nil {
		fmt.Println("Error parsing json")
		return nil, errors.New("Error unmarshaling wallet")
	}
	
	fmt.Println(balance)
	if err2 != nil {
		fmt.Println("Error retrieving balance")
		return nil, errors.New("Error")
	}
 
	balance.Total = balance.Total + amt

	balanceJSONasBytes, _ := json.Marshal(balance)
	err = stub.PutState("coinBalance", balanceJSONasBytes) //rewrite the wallet

	if err != nil {
		fmt.Printf("Error actualizando el balance del negocio")
		return nil, errors.New("Error")
	}
	
	return nil, nil
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

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
