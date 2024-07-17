package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

const (
	channelName   = "testchannel"
	chaincodeName = "basic"
)

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

func main() {
	// 连接到节点Soft
	connectionSoft := newGrpcConnectionSoft()
	defer connectionSoft.Close()

	// 使用Soft节点
	gatewaySoft, err := client.Connect(
		newIdentitySoft(),
		client.WithSign(newSignSoft()),
		client.WithClientConnection(connectionSoft),
		client.WithEvaluateTimeout(10*time.Second),
		client.WithEndorseTimeout(30*time.Second),
		client.WithSubmitTimeout(30*time.Second),
		client.WithCommitStatusTimeout(2*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer gatewaySoft.Close()

	networkSoft := gatewaySoft.GetNetwork(channelName)
	contractSoft := networkSoft.GetContract(chaincodeName)

	// 调用CheckWorking方法，检查链码是否正常工作
	// checkWorking(contractSoft)

	// 上传参数
	uploadParameters(contractSoft, "soft", getMockParameters(), 1)

	time.Sleep(60 * time.Second)

	// 启动聚合
	startAggregation(contractSoft, 1)
}

func checkWorking(contract *client.Contract) {
	fmt.Println("Evaluating Transaction: CheckWorking")
	result, err := contract.EvaluateTransaction("CheckWorking")
	if err != nil {
		fmt.Println("Error evaluating transaction:", err)
		return
	}
	fmt.Println("CheckWorking result:", string(result))
}

func uploadParameters(contract *client.Contract, nodeId string, params LSTMParameters, round int) {
	fmt.Printf("Submitting Transaction: UploadParameter, node ID %s for round %d\n", nodeId, round)

	params.Round = round

	// 序列化每个参数为字符串并传递给链码
	args := []string{
		nodeId,
		string(toBytes(params.Wi)),
		string(toBytes(params.Wf)),
		string(toBytes(params.Wo)),
		string(toBytes(params.Wc)),
		string(toBytes(params.Bi)),
		string(toBytes(params.Bf)),
		string(toBytes(params.Bo)),
		string(toBytes(params.Bc)),
		strconv.Itoa(round),
	}

	_, err := contract.SubmitTransaction("UploadParameter", args...)
	if err != nil {
		fmt.Println("Error submitting transaction:", err)
	}
}

func toBytes(data interface{}) []byte {
	bytes, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Errorf("Error marshaling data: %s", err))
	}
	return bytes
}

func startAggregation(contract *client.Contract, round int) {
	fmt.Printf("Submitting Transaction: StartAggregation, starts aggregation process for round %d\n", round)

	result, err := contract.SubmitTransaction("StartAggregation", strconv.Itoa(round))
	if err != nil {
		fmt.Println("Error submitting transaction:", err)
		return
	}

	fmt.Println("Aggregation result:", string(result))
}

func formatJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, data, "", "  ")
	return prettyJSON.String()
}

// Mock function to generate sample LSTM parameters
func getMockParameters() LSTMParameters {
	return LSTMParameters{
		NodeID: "soft",
		Wi:     [][]float64{{0.1, 0.2}, {0.3, 0.4}},
		Wf:     [][]float64{{0.5, 0.6}, {0.7, 0.8}},
		Wo:     [][]float64{{0.9, 1.0}, {1.1, 1.2}},
		Wc:     [][]float64{{1.3, 1.4}, {1.5, 1.6}},
		Bi:     []float64{0.1, 0.2},
		Bf:     []float64{0.3, 0.4},
		Bo:     []float64{0.5, 0.6},
		Bc:     []float64{0.7, 0.8},
		Round:  1,
	}
}
