package sqlconnect

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	// "github.com/joho/godotenv"
)

func ConnectDb( ) (*sql.DB, error) {
	// err:=godotenv.Load()
	// if err!=nil{
	// 	fmt.Println("Error while loading dotenv")
	// 	return nil, err
	// }
	user:=os.Getenv("DB_USER")
	password:=os.Getenv("DB_PASSWORD")
	dbname:=os.Getenv("DB_NAME")
	dbport:=os.Getenv("DB_PORT")
	dbhost:=os.Getenv("DB_HOST")




	fmt.Println("Trying to connect mariadb")
	// connectionString:="root:root@tcp(127.0.0.1:3307)/" +dbname
	connectionString:=fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, dbhost, dbport, dbname)


	db, err:=sql.Open("mysql", connectionString)
	if err!=nil{
return nil, err


	}
	fmt.Println("Connected to Mariadb")


	return db, nil

}