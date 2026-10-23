package main

import (
	"fmt"
	"reflect"
)

type Student struct{
	FirstName string `db:"first_name"`
	LastName string
	Email string
	Class string
}

func main(){
	s1:=Student{FirstName: "Chetan", LastName: "paudel", Email: "example@gmail.com", Class:"8"}

	modelValue:=reflect.ValueOf(s1)
	modelType:=reflect.TypeOf(s1)
	
	fmt.Println(modelValue.Type())
	fmt.Println(modelValue.Kind())
	fmt.Println(modelValue.CanSet())
	fmt.Println(modelValue.CanConvert(modelType))

fmt.Println(modelValue.NumField())
fmt.Println(modelValue.Field(0))
fmt.Println(modelValue.Field(1))
fmt.Println(modelType.NumField())
fmt.Println(modelType.Field(0))
fmt.Printf("%T, %T",modelValue.Field(0), modelValue.Field(0).Interface())
for i:=0;i<modelType.NumField();i++{
		dbTag:=modelType.Field(i).Tag.Get("db")
		fmt.Println(dbTag)



}
}
