# Aria

**Aria** is a lightweight, event-driven WebSocket framework for Go.
Aria is inspired by [olahol/melody](https://github.com/olahol/melody), and based on [coder/websocket](https://github.com/coder/websocket).
It provides a simple API for managing WebSocket connections, broadcasting messages, and handling events like `OnConnect`, `OnMessage`, `OnClose`, and `OnDisconnect`.  

The goal of Aria is to make building real-time applications (chat servers, dashboards, games, etc.) as straightforward as possible while keeping performance and code clarity.

---

## Features

- ✅ Simple API inspired by event-driven frameworks  
- ✅ Hook functions: `OnConnect`, `OnMessage`, `OnClose`, `OnDisconnect`, `OnMessageBinary`, `OnError`  
- ✅ Broadcast support (`BroadCast` and `BroadCastFilter`)  
- ✅ Connection lifecycle management with graceful cleanup  
- ✅ Context-aware connection handling (`Handle` and `HandleWithContext`)  
- ✅ Support for WebSocket compression and subprotocols via options  

---

## Installation

```bash
go get github.com/n9te9/aria
```

## Quick Start

See the [_example/chat](https://github.com/n9te9/aria/tree/main/_example/chat) directory for a complete chat application using Aria.
This includes both a Go WebSocket server and a simple HTML/JavaScript client.

## Contribution

Contributions are welcome! 🎉
If you’d like to improve Aria, please follow these steps:

1. Fork the repository
2. Create a feature branch (git checkout -b feature/my-feature)
3. Write your code and add/update tests
4. Run go test ./... to ensure all tests pass
5. Open a Pull Request with a clear description of your changes

Please also check for consistency in naming conventions, comments, and code style before submitting.
