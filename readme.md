# Load Balancer in Go

This is a simple Load Balancer implemented in Golang that uses **Round-Robin** as the load balancing algorithm. It also includes basic **health checks** to ensure that requests are only sent to healthy backend servers.

## 🚧 Project Status

This project is still in the **learning and development phase**. Currently, it only supports **Round-Robin** for distributing requests. Future improvements will include:

- Support for additional load balancing algorithms (Least Connections, Weighted Load Balancing, etc.)
- Advanced health checks (TCP, multiple failure thresholds)
- Path-based and Header-based routing
- Session persistence (Sticky Sessions)
- TLS Termination (HTTPS support)
- Auto-scaling & failover
- Detailed metrics & logging

## ⚙️ Configuration

The load balancer reads its configuration from a YAML file. Below is an example configuration:

```yaml
port: 8080
serverList:
  - http://httpbin.org
  - http://httpbin.org
  - http://httpbin.org
healthCheck:
  endpoint: "/"
  intervalInSeconds: 5
```

### Explanation:

- `port`: The port on which the load balancer listens.
- `serverList`: A list of backend servers where requests will be forwarded.
- `healthCheck`:
  - `endpoint`: The health check endpoint to verify server availability.
  - `intervalInSeconds`: The interval at which health checks are performed.

## 🚀 How to Run

### 1. Clone the Repository

```sh
git clone https://github.com/PillaiManish/PiLB
cd PiLB
```

### 2. Install Dependencies

Ensure that you have Golang installed. If not, download it from [golang.org](https://golang.org/dl/).

### 3. Run the Load Balancer

```sh
go run main.go
```

### 4. Test the Load Balancer

You can send requests using `curl` or Postman:

```sh
curl http://localhost:8080
```

This will forward the request to one of the backend servers in a **Round-Robin** fashion.

## 🛠️ Future Enhancements

As the project evolves, the following improvements will be added:

- More load balancing strategies
- Enhanced health checks
- Logging & monitoring support
- Rate limiting and security features
- Auto-scaling capabilities

Stay tuned for updates! 🚀

## 🤝 Contributing

This is an ongoing learning project. If you have suggestions or improvements, feel free to open an issue or a pull request.

## 📜 License

This project is open-source and available under the [MIT License](LICENSE).

---



