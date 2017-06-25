package main

import "os"

func main() {
	a := App{}
	a.Initialize(os.Getenv("APP_ROOT_FOLDER_PATH"))
	a.Run(":8080")
}
