package main

import (
	"context"
	"crypto/tls"
	"fmt"
	zaprpc "github.com/achyuthcodes30/zaprpc"
)

func clientMain() {
	CalcConn, _ := zaprpc.NewConn(context.Background(), "localhost:5000", &zaprpc.ClientConfig{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos: []string{"zaprpc"},
		},
	})
	additionResult, err := zaprpc.Zap(CalcConn, "Calculator.Add", 10, 20)
	fmt.Printf("Result is %d\n",additionResult)
	fmt.Printf("Errors: %v",err)
}
func main() {
	clientMain()
}
