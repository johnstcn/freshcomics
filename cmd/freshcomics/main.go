package main

//go:generate go-bindata -prefix "web/templates/" -pkg web -o web/templates.go web/templates/

func main() {
	root().Execute()
}
