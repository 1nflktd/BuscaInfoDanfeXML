package main

import "os"

func main() {
	a := App{}
	a.Initialize(os.Getenv("APP_ROOT_FOLDER_PATH"), os.Getenv("APP_DATABASE_NAME"))
	defer a.Close()
	a.Run(":8080")
}
