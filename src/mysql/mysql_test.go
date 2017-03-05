package mysql

import "testing"

func TestDataBase_Connect_Close(t *testing.T) {
	db := DataBase{}
	db.Connect()
	defer db.Close()
}

func TestDataBase_Execute(t *testing.T) {
	db := DataBase{}
	db.Connect()
	defer db.Close()

	db.Execute("INSERT INTO test (value1, value2) VALUES (?, ?)", 99, "testing")
}

func TestDataBase_Execute2(t *testing.T) {
	db := DataBase{}
	db.Connect()
	defer db.Close()

	db.Execute("UPDATE test SET value1 = ?, value2 = ?", 20, "pass")
}

func TestDataBase_Execute3(t *testing.T) {
	db := DataBase{}
	db.Connect()
	defer db.Close()

	db.Execute("DELETE FROM test")
}

// TODO : add test methods for all functions.