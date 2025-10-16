package main

import (
	"context"
	"fmt"
	"sync"
	zaprpc "github.com/achyuthcodes30/zaprpc"
	"go.uber.org/zap"
)

// CalculatorService defines the interface for our calculator service
type CalculatorService interface {
	Add(a, b int) int
	Subtract(a, b int) int
	Multiply(a, b int) int
	Divide(a, b int) (float64, error)
}

// CalculatorServiceImpl implements the CalculatorService interface
type CalculatorServiceImpl struct{}

func (c *CalculatorServiceImpl) Add(a, b int) int {
	return a + b
}

func (c *CalculatorServiceImpl) Subtract(a, b int) int {
	return a - b
}

func (c *CalculatorServiceImpl) Multiply(a, b int) int {
	return a * b
}

func (c *CalculatorServiceImpl) Divide(a, b int) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return float64(a) / float64(b), nil
}



func serverMain() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	tr, _ := zaprpc.NewTransport(":5000", logger)
	CalcServer := zaprpc.NewServer(&zaprpc.ServerConfig{Logger: logger, QUICTransport: tr})
	CalcServer.RegisterService("Calculator", new(CalculatorServiceImpl))
	CalcServer.Serve(context.Background())
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func(){
		defer wg.Done()
		serverMain()
	}()
	wg.Wait()
}
