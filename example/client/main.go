package main

import (
	"context"
	"crypto/tls"
	"fmt"
	zaprpc "github.com/achyuthcodes30/zaprpc"
)

func clientMain() {
	client := zaprpc.NewClient(nil)
	CalcConn, _ := zaprpc.NewConn(context.Background(), "localhost:5000", &zaprpc.ConnectionConfig{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos: []string{"zaprpc"},
		},
	})
	
	additionResult, err := client.Zap(context.Background(),CalcConn, "Calculator.Add", 10, 20)
	fmt.Printf("Result is %d\n",additionResult)
	fmt.Printf("Errors: %v",err)
}
func main() {
	clientMain()
}
