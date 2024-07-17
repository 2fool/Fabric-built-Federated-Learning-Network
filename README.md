# Basic Network of Hyperledger Fabric Integrated with Federated Learning

## Setup Instructions

Follow these steps to set up and run the network:

1. **Configure DNS Settings:**
    ```sh
    ./setDns.sh
    ```

2. **Source Environment Variables:**
    ```sh
    source envpeer1soft
    ```

3. **Execute Setup Scripts:**
    Run the following scripts sequentially (in terminal!):
    ```sh
    ./1
    ./2
    ./3
    ./4
    ```
    there might be several permission error make sure u give enough permission for it!

4. **Start the Gateway:**
    Open the gateway files and run the Go application for each node:
    ```sh
    go run .
    ```
    the gateway files are in "contract-gateway", "node_hard" and "node_web". it mostly interacts with chaincode in order to savely pass data.
    connection files are sepreated in names like "connection_xxx.go", make sure you change it to fit your local path:)

You can now interact with the network and explore its functionalities!
(feel free to post issues or send me email)
