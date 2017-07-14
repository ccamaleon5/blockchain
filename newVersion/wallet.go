/*
* Adrian Pareja
 */
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"

	"crypto/rand"
	"encoding/hex"
)

const (
	tableColumn     = "Movimientos"
	columnTime      = "Time"
	columnAccountID = "Account"
	columnBusiness  = "Business"
	columnAmount    = "Amount"
	columnBalance   = "Balance"
	columnType      = "Type"
)

const (
	tableWalletColumn     = "Wallet"
	columnWalletId  = "Id"
	columnWalletAccountID = "Account"
	columnWalletBalance   = "Balance"
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
	Limit    float64 `json:"limit"`
}

//Wallet - Structure for products used in buy goods
type WalletBalance struct {
	WalletId  string  `json:"walletid"`
	Balance   float64 `json:"balance"`
}

//Movimiento - Structure for movements
type Movement struct {
	Time     int64   `json:"time"`
	WalletId string  `json:"walletid"`
	Business string  `json:"business"`
	Amount   float64 `json:"amount"`
	Balance  float64 `json:"balance"`
	Type     string  `json:"type"`
}

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func main() {
	fmt.Printf("Iniciandooo Contrato Wallet....")
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error Iniciando Wallet Smart Contract: %s", err)
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

	stub.CreateTable(tableColumn, []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: columnAccountID, Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: columnTime, Type: shim.ColumnDefinition_INT64, Key: true},
		&shim.ColumnDefinition{Name: columnBusiness, Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: columnAmount, Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: columnBalance, Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: columnType, Type: shim.ColumnDefinition_STRING, Key: false},
	})
	
	stub.CreateTable(tableWalletColumn, []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: columnWalletId, Type: shim.ColumnDefinition_STRING, Key: true},	
		&shim.ColumnDefinition{Name: columnAccountID, Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: columnBalance, Type: shim.ColumnDefinition_STRING, Key: false},
	})

	return nil, nil
}

// Invoke Punto de entrada a cualquier función del ledger
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running..FUNCTION:" + function)

	if function == "createwallet" {
		return t.createWallet(stub, args)
	} else {
		if function == "transfer" {
			return t.transfer(stub, args)
		} else {
			if function == "putbalance" {
				return t.putBalance(stub, args)
			} else if function == "debitbalance" {
				return t.debitBalance(stub, args)
			}
		}
	}
	fmt.Println("invoke no encuentra la funcion: " + function)

	return nil, errors.New("Funcion invocada desconocida: " + function)
}

// Query es nuestro punto de entrada de querys
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running FUNCTION:" + function)

	// Manejar diferentes funciones
	if function == "getbalance" {
		return t.getBalance(stub, args)
	} else {
		if function == "gettotalcoin" {
			return t.getTotalCoin(stub, args)
		} else {
			if function == "getmovimientos" {
				return t.getMovimientos(stub, args)
			} else { 
				if function == "getdatos" {
					return t.getDatos(stub, args)
				} else if function == "getwallets" {
					return t.getWallets(stub, args)
				} 
			}
		}
	}
	fmt.Println("query no encuentra la funcion: " + function)

	return nil, errors.New("Funcion invocada desconocida: " + function)
}

// createWallet - invocar esta funcion para crear un wallet con saldo inicial
func (t *SimpleChaincode) createWallet(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call---Funcion createWallet---")
	if len(args) != 6 {
		return nil, errors.New("Numero incorrecto de argumentos.Se espera 6 para createWallet")
	}
	
	bytesWallet, _ := stub.GetState(args[0])
	if bytesWallet != nil {
		fmt.Println("Ya existe el wallet con id %s",args[0])
		return nil,errors.New("El wallet ya existe")
	}

	walletId := NewV4()
	fmt.Printf("UUIDv4: %s\n", walletId)
	amt, err := strconv.ParseFloat("0", 64)

	if err != nil {
		fmt.Println("Error Float parsing")
		return nil, errors.New("Error marshaling wallet")
	}

	wallet := Wallet{
		Id:       args[0],
		Email:    args[1],
		Phone:    args[2],
		Document: args[3],
		Password: args[4],
		Amount:   amt,
		Limit:    100,
	}

	bytes, err := json.Marshal(wallet)
	if err != nil {
		fmt.Println("Error marshaling wallet")
		return nil, errors.New("Error marshaling wallet")
	}

	err = stub.PutState(wallet.Id, bytes)
	if err != nil {
		return nil, err
	}

	a := makeTimestamp()

	fmt.Printf("Time: %d \n", a)

	col1Val := args[0]
	col2Val := "Create"
	col3Val := "0"
	col5Val := "W"

	var columns []*shim.Column
	col0 := shim.Column{Value: &shim.Column_String_{String_: col1Val}}
	col1 := shim.Column{Value: &shim.Column_Int64{Int64: a}}
	col2 := shim.Column{Value: &shim.Column_String_{String_: col2Val}}
	col3 := shim.Column{Value: &shim.Column_String_{String_: col3Val}}
	col4 := shim.Column{Value: &shim.Column_String_{String_: col3Val}}
	col5 := shim.Column{Value: &shim.Column_String_{String_: col5Val}}
	columns = append(columns, &col0)
	columns = append(columns, &col1)
	columns = append(columns, &col2)
	columns = append(columns, &col3)
	columns = append(columns, &col4)
	columns = append(columns, &col5)

	row := shim.Row{Columns: columns}
	ok, err := stub.InsertRow("Movimientos", row)
	if err != nil {
		return nil, fmt.Errorf("Insert Row Movimientos operation failed. %s", err)
	}
	if !ok {
		return nil, errors.New("Fallo insertar Row with given key already exists")
	}
	
	//Se inserta el Wallet en la Tabla de Wallets
	fmt.Printf("Insertando Wallet en la Tabla")

	col11Val := args[0]

	var columns1 []*shim.Column
	col10 := shim.Column{Value: &shim.Column_String_{String_: "Wallet"}}
	col11 := shim.Column{Value: &shim.Column_String_{String_: col11Val}}
	col12 := shim.Column{Value: &shim.Column_String_{String_: "0.00"}}
	
	columns1 = append(columns1, &col10)
	columns1 = append(columns1, &col11)
	columns1 = append(columns1, &col12)

	row1 := shim.Row{Columns: columns1}
	ok1, err1 := stub.InsertRow("Wallet", row1)
	if err1 != nil {
		return nil, fmt.Errorf("Insert Row Wallet operation failed. %s", err1)
	}
	if !ok1 {
		return nil, errors.New("Fallo insertar Row with given key already exists")
	}
	
	//Se envia un evento de exito
	err = stub.SetEvent("createWallet", []byte("createWallet:OK"))
	if err != nil {
		return nil, errors.New("Fallo enviar el evento de Crear Wallet")
	}

	return []byte(`{"code":0,"response":null}`), nil
}

// putBalance - invocar esta funcion incrementar los coins en el balance
func (t *SimpleChaincode) putBalance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call---Funcion PutBalance---")
	if len(args) != 3 {
		return nil, errors.New("Numero incorrecto de argumentos.Se espera 3 para putBalance")
	}

	fmt.Printf("WalletId 1: %s\n", args[0])
	fmt.Printf("Business: %s\n", args[1])
	fmt.Printf("Monto: %s\n", args[2])

	bytesWallet1, err1 := stub.GetState(args[0])

	walletReceiver := Wallet{}
	err := json.Unmarshal(bytesWallet1, &walletReceiver)

	fmt.Println(walletReceiver)
	if err1 != nil {
		fmt.Println("Error retrieving " + args[0])
		return nil, errors.New("Error retrieving " + args[0])
	}

	amt, err := strconv.ParseFloat(args[2], 64)

	walletReceiver.Amount = walletReceiver.Amount + amt //carga coins al balance

	walletReceiverJSONasBytes, _ := json.Marshal(walletReceiver)
	err = stub.PutState(args[0], walletReceiverJSONasBytes) //rewrite the wallet

	if err != nil {
		return nil, err
	}

	a := makeTimestamp()

	fmt.Printf("Time: %d \n", a)

	col1Val := args[0]
	col2Val := args[1]
	col3Val := strconv.FormatFloat(amt, 'f', 6, 64)
	col4Val := strconv.FormatFloat(walletReceiver.Amount, 'f', 6, 64)
	col5Val := "C"

	var columns []*shim.Column
	col0 := shim.Column{Value: &shim.Column_String_{String_: col1Val}}
	col1 := shim.Column{Value: &shim.Column_Int64{Int64: a}}
	col2 := shim.Column{Value: &shim.Column_String_{String_: col2Val}}
	col3 := shim.Column{Value: &shim.Column_String_{String_: col3Val}}
	col4 := shim.Column{Value: &shim.Column_String_{String_: col4Val}}
	col5 := shim.Column{Value: &shim.Column_String_{String_: col5Val}}
	columns = append(columns, &col0)
	columns = append(columns, &col1)
	columns = append(columns, &col2)
	columns = append(columns, &col3)
	columns = append(columns, &col4)
	columns = append(columns, &col5)

	row := shim.Row{Columns: columns}
	ok, err := stub.InsertRow("Movimientos", row)
	if err != nil {
		return nil, fmt.Errorf("Insert Row Movimientos operation failed. %s", err)
	}
	if !ok {
		return nil, errors.New("Fallo insertar Row with given key already exists")
	}
	
	//Se actualiza el row de Wallet
	var columns1 []*shim.Column
	col10 := shim.Column{Value: &shim.Column_String_{String_: "Wallet"}}
	col11 := shim.Column{Value: &shim.Column_String_{String_: col1Val}}
	col12 := shim.Column{Value: &shim.Column_String_{String_: col4Val}}
	
	columns1 = append(columns1, &col10)
	columns1 = append(columns1, &col11)
	columns1 = append(columns1, &col12)

	row1 := shim.Row{Columns: columns1}
	ok2, err2 := stub.ReplaceRow("Wallet",row1)
	if err2 != nil {
		return nil, fmt.Errorf("Insert Row Wallet operation failed. %s", err2)
	}
	if !ok2 {
		return nil, errors.New("Fallo insertar Row Wallet with given key already exists")
	}
	
	//Se envia un evento de exito
	err = stub.SetEvent("debitEvent", []byte("debitEvent:4"))
	if err != nil {
		return nil, errors.New("Fallo enviar el evento de debito")
	}

	return []byte(`{"code":0,"response":null}`), nil
}

// debitBalance - invocar esta funcion debitar coins del balance
func (t *SimpleChaincode) debitBalance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call---Funcion DebitBalance---")
	if len(args) != 3 {
		return nil, errors.New("Numero incorrecto de argumentos.Se espera 2 para debitBalance")
	}

	fmt.Printf("WalletId 1: %s\n", args[0])
	fmt.Printf("Monto: %s\n", args[1])

	bytesWallet1, err1 := stub.GetState(args[0])

	walletReceiver := Wallet{}
	err := json.Unmarshal(bytesWallet1, &walletReceiver)

	fmt.Println(walletReceiver)
	if err1 != nil {
		fmt.Println("Error retrieving " + args[0])
		return nil, errors.New("Error retrieving " + args[0])
	}

	amt, err := strconv.ParseFloat(args[2], 64)

	walletReceiver.Amount = walletReceiver.Amount - amt //debita coins del balance
	walletReceiver.Limit = walletReceiver.Limit - amt

	walletReceiverJSONasBytes, _ := json.Marshal(walletReceiver)
	err = stub.PutState(args[0], walletReceiverJSONasBytes) //rewrite the wallet

	if err != nil {
		return nil, err
	}

	a := makeTimestamp()

	fmt.Printf("Time: %d \n", a)

	col1Val := args[0]
	col2Val := args[1]
	col3Val := strconv.FormatFloat(amt, 'f', 6, 64)
	col4Val := strconv.FormatFloat(walletReceiver.Amount, 'f', 6, 64)
	col5Val := "D"

	var columns []*shim.Column
	col0 := shim.Column{Value: &shim.Column_String_{String_: col1Val}}
	col1 := shim.Column{Value: &shim.Column_Int64{Int64: a}}
	col2 := shim.Column{Value: &shim.Column_String_{String_: col2Val}}
	col3 := shim.Column{Value: &shim.Column_String_{String_: col3Val}}
	col4 := shim.Column{Value: &shim.Column_String_{String_: col4Val}}
	col5 := shim.Column{Value: &shim.Column_String_{String_: col5Val}}
	columns = append(columns, &col0)
	columns = append(columns, &col1)
	columns = append(columns, &col2)
	columns = append(columns, &col3)
	columns = append(columns, &col4)
	columns = append(columns, &col5)

	row := shim.Row{Columns: columns}
	ok, err := stub.InsertRow("Movimientos", row)
	if err != nil {
		return nil, fmt.Errorf("Insert Row Movimientos operation failed. %s", err)
	}
	if !ok {
		return nil, errors.New("Fallo insertar Row with given key already exists")
	}
	
	//Se actualiza el row de Wallet
	var columns1 []*shim.Column
	col10 := shim.Column{Value: &shim.Column_String_{String_: "Wallet"}}
	col11 := shim.Column{Value: &shim.Column_String_{String_: col1Val}}
	col12 := shim.Column{Value: &shim.Column_String_{String_: col4Val}}
	
	columns1 = append(columns1, &col10)
	columns1 = append(columns1, &col11)
	columns1 = append(columns1, &col12)

	row1 := shim.Row{Columns: columns1}
	ok3, err3 := stub.ReplaceRow("Wallet",row1)
	if err3 != nil {
		return nil, fmt.Errorf("Insert Row Wallet operation failed. %s", err3)
	}
	if !ok3 {
		return nil, errors.New("Fallo insertar Row Wallet with given key already exists")
	}
	
	//Se actualiza el balance global de coin
	
	coinBalance, err2 := stub.GetState("coinBalance")
	fmt.Println(coinBalance)
	if err2 != nil {
		fmt.Println("Error retrieving coinBalance")
		return nil, errors.New("Error retrieving coinBalance")
	}
	
	newCoinBalance,_ := strconv.ParseFloat(string(coinBalance), 64)
	newCoinBalance = newCoinBalance + amt
	
	err = stub.PutState("coinBalance", []byte(strconv.FormatFloat(newCoinBalance, 'f', 6, 64)))
	if err != nil {
		fmt.Println("Error setting new coinBalance")
		return nil, err
	}

	return []byte(`{"code":0,"response":null}`), nil
}

// transfer - invocar esta funcion para transferir coins de un wallet a otro
func (t *SimpleChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call---Funcion Transfer---")
	if len(args) != 3 {
		return nil, errors.New("Numero incorrecto de argumentos.Se espera 3 para createWallet")
	}

	fmt.Printf("WalletId 1: %s\n", args[1])
	fmt.Printf("WalletId 2: %s\n", args[0])
	fmt.Printf("Monto: %s\n", args[2])

	bytesWallet1, err1 := stub.GetState(args[1])

	walletSender := Wallet{}
	err := json.Unmarshal(bytesWallet1, &walletSender)

	if err1 != nil {
		fmt.Println("Error retrieving " + args[1])
		return nil, errors.New("Error retrieving " + args[1])
	}
	fmt.Println(walletSender)

	bytesWallet2, err2 := stub.GetState(args[0])
	walletReceiver := Wallet{}
	err = json.Unmarshal(bytesWallet2, &walletReceiver)
	if err2 != nil {
		fmt.Println("Error retrieving " + args[0])
		return nil, errors.New("Error retrieving " + args[0])
	}
	fmt.Println(walletReceiver)

	amt, err := strconv.ParseFloat(args[2], 64)

	if amt <= walletSender.Amount {
		walletSender.Amount = walletSender.Amount - amt     //debita el monto
		walletReceiver.Amount = walletReceiver.Amount + amt //carga el monto

		walletSenderJSONasBytes, _ := json.Marshal(walletSender)
		err = stub.PutState(args[1], walletSenderJSONasBytes) //rewrite the wallet

		if err != nil {
			fmt.Println("Error guardar el Sender")
			return nil, err
		}

		walletReceiverJSONasBytes, _ := json.Marshal(walletReceiver)
		err = stub.PutState(args[0], walletReceiverJSONasBytes) //rewrite the wallet
		if err != nil {
			fmt.Println("Error guardar el Recceiver")
			return nil, err
		}

		a := makeTimestamp()

		fmt.Printf("Time: %d \n", a)

		col1Val := args[0]
		col2Val := args[1]
		col3Val := strconv.FormatFloat(amt, 'f', 6, 64)
		col4Val := strconv.FormatFloat(walletReceiver.Amount, 'f', 6, 64)
		col5Val := "C"

		var columns []*shim.Column
		var columns2 []*shim.Column
		col0 := shim.Column{Value: &shim.Column_String_{String_: col1Val}}
		col1 := shim.Column{Value: &shim.Column_Int64{Int64: a}}
		col2 := shim.Column{Value: &shim.Column_String_{String_: col2Val}}
		col3 := shim.Column{Value: &shim.Column_String_{String_: col3Val}}
		col4 := shim.Column{Value: &shim.Column_String_{String_: col4Val}}
		col5 := shim.Column{Value: &shim.Column_String_{String_: col5Val}}
		columns = append(columns, &col0)
		columns = append(columns, &col1)
		columns = append(columns, &col2)
		columns = append(columns, &col3)
		columns = append(columns, &col4)
		columns = append(columns, &col5)
		
		row := shim.Row{Columns: columns}
		ok, err := stub.InsertRow("Movimientos", row)
		if err != nil {
			fmt.Println("Error al insertar la fila de sender")
			return nil, fmt.Errorf("Insert Row Movimientos operation failed. %s", err)
		}
		if !ok {
			return nil, errors.New("Fallo insertar Row with given key already exists")
		}

		fmt.Println("Inserto Fila de Sender")

		b := a+1
		
		fmt.Printf("Time: %d \n", b)
		
		col1Val = args[1]
		col2Val = args[0]
		col3Val = strconv.FormatFloat(amt, 'f', 6, 64)
		col4Val = strconv.FormatFloat(walletSender.Amount, 'f', 6, 64)
		col5Val = "D"

		col0 = shim.Column{Value: &shim.Column_String_{String_: col1Val}}
		col1 = shim.Column{Value: &shim.Column_Int64{Int64: b}}
		col2 = shim.Column{Value: &shim.Column_String_{String_: col2Val}}
		col3 = shim.Column{Value: &shim.Column_String_{String_: col3Val}}
		col4 = shim.Column{Value: &shim.Column_String_{String_: col4Val}}
		col5 = shim.Column{Value: &shim.Column_String_{String_: col5Val}}

		columns2 = append(columns2, &col0)
		columns2 = append(columns2, &col1)
		columns2 = append(columns2, &col2)
		columns2 = append(columns2, &col3)
		columns2 = append(columns2, &col4)
		columns2 = append(columns2, &col5)

		row2 := shim.Row{Columns: columns2}
		ok2, err2 := stub.InsertRow("Movimientos", row2)
		if err2 != nil {
			fmt.Println("Error al insertar la fila de receiver")
			return nil, fmt.Errorf("Insert Row2 Movimientos operation failed. %s", err2)
		}
		if !ok2 {
			return nil, errors.New("Fallo insertar Row2 with given key already exists")
		}

		fmt.Println("Inserto fila de receiver")

		return []byte(`{"code":0,"response":null}`), nil
	} else {
		return nil, errors.New("No cuentas con suficientes coins para esta transferencia")
	}
}

func (t *SimpleChaincode) getBalance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----getBalance() is running----")

	if len(args) != 1 {
		return nil, errors.New("Incorrecto numero de argumentos. Se esperaba 1")
	}

	walletId := args[0] // wallet id
	fmt.Println("wallet id is ")
	fmt.Println(walletId)
	bytes, err := stub.GetState(args[0])
	if err != nil {
		fmt.Println("Error retrieving " + walletId)
		return nil, errors.New("Error retrieving " + walletId)
	}
	wallet := Wallet{}
	err1 := json.Unmarshal(bytes, &wallet)

	fmt.Println(wallet.Amount)
	if err1 != nil {
		fmt.Println("Error parseando a Json" + args[0])
		return nil, errors.New("Error retrieving Balance" + args[0])
	}

	return []byte(fmt.Sprintf(`{"code":0,"balance":"%s","limit":"%s"}`, strconv.FormatFloat(wallet.Amount, 'f', 6, 64), strconv.FormatFloat(wallet.Limit, 'f', 6, 64))), nil
}

//Funcion que obtiene el total de Coins en el sistema
func (t *SimpleChaincode) getTotalCoin(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----getTotalCoin() is running----")

	coinBalance, err := stub.GetState("coinBalance")
	fmt.Println(coinBalance)
	if err != nil {
		fmt.Println("Error retrieving coinBalance")
		return nil, errors.New("Error retrieving coinBalance")
	}

	return []byte(fmt.Sprintf(`{"code":0,"response":"%s"}`, coinBalance)), nil
}

//Funcion que otorga coins a los negocios
func (t *SimpleChaincode) debitTotalCoin(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----debitTotalCoin() is running----")

	coinBalance, err := stub.GetState("coinBalance")
	fmt.Println(coinBalance)
	if err != nil {
		fmt.Println("Error retrieving coinBalance")
		return nil, errors.New("Error retrieving coinBalance")
	}
	
	newCoinBalance,_ := strconv.ParseFloat(string(coinBalance), 64)
	newCoinBalance = newCoinBalance - 250000
	
	err = stub.PutState("coinBalance", []byte(strconv.FormatFloat(newCoinBalance, 'f', 6, 64)))
	if err != nil {
		fmt.Println("Error setting new coinBalance")
		return nil, err
	}

	return []byte(fmt.Sprintf(`{"code":0,"response":"%s"}`, coinBalance)), nil
}

//Funcion que obtiene todos los movimientos de un usuario
func (t *SimpleChaincode) getMovimientos(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----getMovimientos() is running----")

	if len(args) != 1 {
		return nil, errors.New("Incorrecto numero de argumentos. Se esperaba 1")
	}

	walletId := args[0] // wallet id
	fmt.Println("wallet id is ")
	fmt.Println(walletId)
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: walletId}}
	columns = append(columns, col1)

	rowChannel, err := stub.GetRows("Movimientos", columns)
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
				amountRow, _ := strconv.ParseFloat(columnas[3].GetString_(), 64)
				balanceRow, _ := strconv.ParseFloat(columnas[4].GetString_(), 64)
				movimiento := Movement{Time: columnas[1].GetInt64(), WalletId: columnas[0].GetString_(), Business: columnas[2].GetString_(), Amount: amountRow, Balance: balanceRow, Type: columnas[5].GetString_()}

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

//Funcion que obtiene todos los wallets ubicados en la blockchain
func (t *SimpleChaincode) getWallets(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----getWallets() is running----")

	if len(args) != 1 {
		return nil, errors.New("Incorrecto numero de argumentos. Se esperaba 1")
	}

	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "Wallet"}}
	columns = append(columns, col1)

	rowChannel, err := stub.GetRows("Wallet", columns)
	if err != nil {
		return nil, fmt.Errorf("getRow TableWallet operation failed. %s", err)
	}

	wallets := []WalletBalance{}
	for {
		select {
		case row, ok := <-rowChannel:
			if !ok {
				rowChannel = nil
			} else {
				columnas := row.GetColumns()
				balanceRow, _ := strconv.ParseFloat(columnas[2].GetString_(), 64)
				wallet := WalletBalance{WalletId: columnas[1].GetString_(), Balance: balanceRow}

				wallets = append(wallets, wallet)
			}
		}
		if rowChannel == nil {
			break
		}
	}

	jsonRows, err := json.Marshal(wallets)
	if err != nil {
		return nil, fmt.Errorf("getRows Wallet operation failed. Error marshaling JSON: %s", err)
	}

	return jsonRows, nil
}

//getData - Obtiene los datos generales del usuario
func (t *SimpleChaincode) getDatos(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----getDatos() is running----")

	if len(args) != 1 {
		return nil, errors.New("Incorrecto numero de argumentos. Se esperaba 1")
	}

	walletId := args[0] // wallet id
	fmt.Println("wallet id is")
	fmt.Println(walletId)
	bytes, err := stub.GetState(args[0])
	if err != nil {
		fmt.Println("Error retrieving " + walletId)
		return nil, errors.New("Error retrieving " + walletId)
	}
	wallet := Wallet{}
	err1 := json.Unmarshal(bytes, &wallet)

	fmt.Println(wallet)
	if err1 != nil {
		fmt.Println("Error parseando a Json" + args[0])
		return nil, errors.New("Error retrieving Balance" + args[0])
	}

	return bytes, nil
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
