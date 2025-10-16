package main

import (
	"context"
	"crypto/tls"
	"fmt"
	zaprpc "github.com/achyuthcodes30/zaprpc"
	"go.uber.org/zap"
)

func clientMain() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	client := zaprpc.NewClient(&zaprpc.ClientConfig{
		Logger: logger.Named("zaprpc-client"),
	})
	CalcConn, _ := zaprpc.NewConn(context.Background(), "localhost:5000", &zaprpc.ConnectionConfig{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos: []string{"zaprpc"},
		},
	})
	
	additionResult, err := client.Zap(CalcConn, "Calculator.Add", 10, 20)
	fmt.Printf("Result is %d\n",additionResult)
	fmt.Printf("Errors: %v",err)
}
func main() {
	clientMain()
}
