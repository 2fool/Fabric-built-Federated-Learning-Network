package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Constants for the timeouts and other configurations
const (
	AggregationTimeout = 60 * time.Second
	CheckInterval      = 10 * time.Second
)

// SmartContract provides functions for managing the aggregation
type SmartContract struct {
	contractapi.Contract
}

// LSTMParameters represents the parameters uploaded by nodes
type LSTMParameters struct {
	NodeID string      `json:"nodeId"`
	Wi     [][]float64 `json:"Wi"`
	Wf     [][]float64 `json:"Wf"`
	Wo     [][]float64 `json:"Wo"`
	Wc     [][]float64 `json:"Wc"`
	Bi     []float64   `json:"bi"`
	Bf     []float64   `json:"bf"`
	Bo     []float64   `json:"bo"`
	Bc     []float64   `json:"bc"`
	Round  int         `json:"round"`
}

// AggregatedResult represents the result of aggregation
type AggregatedResult struct {
	NodeID string      `json:"nodeId"`
	Wi     [][]float64 `json:"Wi"`
	Wf     [][]float64 `json:"Wf"`
	Wo     [][]float64 `json:"Wo"`
	Wc     [][]float64 `json:"Wc"`
	Bi     []float64   `json:"bi"`
	Bf     []float64   `json:"bf"`
	Bo     []float64   `json:"bo"`
	Bc     []float64   `json:"bc"`
	Round  int         `json:"round"`
}

// Temporary storage for parameters
var localParams = struct {
	sync.RWMutex
	m map[string]LSTMParameters
}{m: make(map[string]LSTMParameters)}

// UploadParameter allows a node to upload its LSTM parameters
func (s *SmartContract) UploadParameter(ctx contractapi.TransactionContextInterface, nodeId string, wi, wf, wo, wc [][]float64, bi, bf, bo, bc []float64, round int) error {
	localParams.Lock()
	defer localParams.Unlock()

	param := LSTMParameters{
		NodeID: nodeId,
		Wi:     wi,
		Wf:     wf,
		Wo:     wo,
		Wc:     wc,
		Bi:     bi,
		Bf:     bf,
		Bo:     bo,
		Bc:     bc,
		Round:  round,
	}
	localParams.m[nodeId+"_"+strconv.Itoa(round)] = param
	return nil
}

// Check if all parameters are submitted
func areAllParametersSubmitted(keys []string, round int) bool {
	localParams.RLock()
	defer localParams.RUnlock()
	count := 0
	var missingNodes []string
	for _, key := range keys {
		if _, exists := localParams.m[key+"_"+strconv.Itoa(round)]; exists {
			count++
		} else {
			missingNodes = append(missingNodes, key)
		}
	}
	if count == len(keys) {
		fmt.Println("All parameters are submitted")
	} else {
		fmt.Printf("Missing parameters for node(s): %v\n", missingNodes)
	}
	return count == len(keys)
}

// Use previous round parameters if current round parameters are not complete
func usePreviousRoundParameters(keys []string, round int) error {
	localParams.Lock()
	defer localParams.Unlock()
	for _, key := range keys {
		prevParam, exists := localParams.m[key+"_"+strconv.Itoa(round-1)]
		if exists {
			localParams.m[key+"_"+strconv.Itoa(round)] = prevParam
		} else {
			return fmt.Errorf("parameter for node %s not found in previous round", key)
		}
	}
	return nil
}

// Initialize sum matrices and vectors
func initializeSumMatrices(keys []string, round int) ([][]float64, [][]float64, [][]float64, [][]float64) {
	localParams.RLock()
	defer localParams.RUnlock()

	var sumWi, sumWf, sumWo, sumWc [][]float64
	for i := range localParams.m[keys[0]+"_"+strconv.Itoa(round)].Wi {
		sumWi = append(sumWi, make([]float64, len(localParams.m[keys[0]+"_"+strconv.Itoa(round)].Wi[i])))
		sumWf = append(sumWf, make([]float64, len(localParams.m[keys[0]+"_"+strconv.Itoa(round)].Wf[i])))
		sumWo = append(sumWo, make([]float64, len(localParams.m[keys[0]+"_"+strconv.Itoa(round)].Wo[i])))
		sumWc = append(sumWc, make([]float64, len(localParams.m[keys[0]+"_"+strconv.Itoa(round)].Wc[i])))
	}
	return sumWi, sumWf, sumWo, sumWc
}

// Aggregate parameters
func aggregateParameters(keys []string, round int) ([][]float64, [][]float64, [][]float64, [][]float64, []float64, []float64, []float64, []float64) {
	sumWi, sumWf, sumWo, sumWc := initializeSumMatrices(keys, round)
	sumBi := make([]float64, len(localParams.m[keys[0]+"_"+strconv.Itoa(round)].Bi))
	sumBf := make([]float64, len(localParams.m[keys[0]+"_"+strconv.Itoa(round)].Bf))
	sumBo := make([]float64, len(localParams.m[keys[0]+"_"+strconv.Itoa(round)].Bo))
	sumBc := make([]float64, len(localParams.m[keys[0]+"_"+strconv.Itoa(round)].Bc))

	localParams.RLock()
	defer localParams.RUnlock()

	for _, key := range keys {
		param := localParams.m[key+"_"+strconv.Itoa(round)]
		for i := range param.Wi {
			for j := range param.Wi[i] {
				sumWi[i][j] += param.Wi[i][j]
				sumWf[i][j] += param.Wf[i][j]
				sumWo[i][j] += param.Wo[i][j]
				sumWc[i][j] += param.Wc[i][j]
			}
		}
		for i := range param.Bi {
			sumBi[i] += param.Bi[i]
			sumBf[i] += param.Bf[i]
			sumBo[i] += param.Bo[i]
			sumBc[i] += param.Bc[i]
		}
	}

	return sumWi, sumWf, sumWo, sumWc, sumBi, sumBf, sumBo, sumBc
}

// Calculate average parameters
func calculateAverage(sumWi, sumWf, sumWo, sumWc [][]float64, sumBi, sumBf, sumBo, sumBc []float64, count int) {
	for i := range sumWi {
		for j := range sumWi[i] {
			sumWi[i][j] /= float64(count)
			sumWf[i][j] /= float64(count)
			sumWo[i][j] /= float64(count)
			sumWc[i][j] /= float64(count)
		}
	}
	for i := range sumBi {
		sumBi[i] /= float64(count)
		sumBf[i] /= float64(count)
		sumBo[i] /= float64(count)
		sumBc[i] /= float64(count)
	}
}

// StartAggregation performs the aggregation of the parameters if all nodes have submitted
func (s *SmartContract) StartAggregation(ctx contractapi.TransactionContextInterface, round int) (string, error) {
	keys := []string{"soft", "web", "hard"} // NodeIDs, assumed to be known identifiers

	startTime := time.Now()
	for time.Since(startTime) < AggregationTimeout {
		if areAllParametersSubmitted(keys, round) {
			break
		}
		time.Sleep(CheckInterval)
	}

	// Check if all parameters are submitted within the timeout period
	if !areAllParametersSubmitted(keys, round) {
		// If not all parameters are submitted, use the previous round parameters
		err := usePreviousRoundParameters(keys, round)
		if err != nil {
			return "", err
		}
	}

	sumWi, sumWf, sumWo, sumWc, sumBi, sumBf, sumBo, sumBc := aggregateParameters(keys, round)
	calculateAverage(sumWi, sumWf, sumWo, sumWc, sumBi, sumBf, sumBo, sumBc, len(keys))

	timestamp := time.Now().Unix()
	selectedNodeIndex := int(timestamp % int64(len(keys)))
	selectedNode := keys[selectedNodeIndex]

	result := AggregatedResult{
		NodeID: selectedNode,
		Wi:     sumWi,
		Wf:     sumWf,
		Wo:     sumWo,
		Wc:     sumWc,
		Bi:     sumBi,
		Bf:     sumBf,
		Bo:     sumBo,
		Bc:     sumBc,
		Round:  round,
	}
	resultAsBytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %v", err)
	}
	err = ctx.GetStub().PutState("RESULT_Aggregated_"+strconv.Itoa(round), resultAsBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put state: %v", err)
	}

	// 将全局模型参数返回给节点
	for _, key := range keys {
		err := sendGlobalModelToNode(ctx, key, result)
		if err != nil {
			return "", fmt.Errorf("failed to send global model to node %s: %v", key, err)
		}
	}

	// 释放当前轮次参数的内存
	localParams.Lock()
	defer localParams.Unlock()
	for _, key := range keys {
		delete(localParams.m, key+"_"+strconv.Itoa(round))
	}

	return string(resultAsBytes), nil
}

// sendGlobalModelToNode sends the aggregated global model to a specific node
func sendGlobalModelToNode(ctx contractapi.TransactionContextInterface, nodeID string, result AggregatedResult) error {
	// 使用链码事件发送结果到节点
	resultAsBytes, err := json.Marshal(result)
	if err != nil {
		return err
	}
	err = ctx.GetStub().SetEvent("GlobalModelUpdate", resultAsBytes)
	if err != nil {
		return err
	}
	fmt.Printf("Sending global model to node %s: %+v\n", nodeID, result)
	return nil
}

func (s *SmartContract) CheckWorking(ctx contractapi.TransactionContextInterface) (string, error) {
	return fmt.Sprintf("chaincode is working"), nil
}

func main() {
	rand.Seed(time.Now().UnixNano())
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating smart contract: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting smart contract: %s", err.Error())
	}
}
