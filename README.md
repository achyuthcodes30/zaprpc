# ZapRPC ⚡
A minimal and easy-to-use RPC framework over QUIC and HTTP/2 for Go Developers!

ZapRPC utilizes the Go language itself as an IDL, eliminating the need for protobuf compilation and providing a seamless experience for Go developers.

## Setup ⚙️

**NOTE:** Do ensure you have the latest versions of **Golang** available locally.

```bash
git clone https://github.com/acmpesuecc/zaprpc.git
cd zaprpc
go mod download
go mod tidy # if adding new modules 
```
## Usage 🚀

The `example/` directory contains basic client and server implementations.

Running the client:
```bash
go run example/client/main.go
```
Running the server:
```bash
go run example/server/main.go
```
## Install and use in your project ♻️

Add Dependency:
```bash
go get github.com/achyuthcodes30/zaprpc
```

Import:
```go
import "github.com/achyuthcodes30/zaprpc"
```

Then setup client and server as shown in `example/`

## Contributing ⭐

Want to get involved? Check out the [CONTRIBUTING.md](CONTRIBUTING.md) guide to learn how you can contribute code, suggest improvements, or report issues.

## License 📜

This project is licensed under the MIT License — free for personal and commercial use with attribution.

See the [LICENSE](LICENSE) file for more details.

## Acknowledgements 🤝

- [quic-go](https://quic-go.net) - QUIC library
- [zap](https://github.com/uber-go/zap) - Structured Logging

## Maintainer(s)

[**Achyuth Yogesh Sosale**](https://github.com/achyuthcodes30) - achyuthyogesh0@gmail.com
