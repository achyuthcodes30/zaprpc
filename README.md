# ZapRPC
A minimal and easy-to-use RPC framework over QUIC for Go Developers!

ZapRPC uses the Go language itself as an IDL, eliminating the need for protobuf compilation and providing a seamless experience to go developers!

## TODO
- [x] Calls over QUIC transport
- [x] Gob serialisation
- [ ]  Service builder using reflection for a local call like interface (eg. Calculator.Add())
- [ ]  Improve error handling and interface
- [ ]  Improve context flow
- [ ]  Streaming patterns (Unary Client, Unary Server)
- [ ]  Struct Tags support
- [ ]  Concurrency Support
- [ ]  Inerceptors/Middleware
- [ ]  Improve reliability
- [ ]  Auth
- [ ]  Load Balancing
- [ ]  Optimisations
- [ ]  Benchmarking
