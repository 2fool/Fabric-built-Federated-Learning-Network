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
    Run the following scripts sequentially:
    ```sh
    ./1
    ./2
    ./3
    ./4
    ```

4. **Start the Gateway:**
    Open the gateway files and run the Go application for each node:
    ```sh
    go run .
    ```

You can now interact with the network and explore its functionalities!
