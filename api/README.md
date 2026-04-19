# API Directory

This directory contains all API definitions and contracts for the GoConnect project.

## Structure

```
api/
└── shared/              # Shared API contracts across services
    ├── proto/          # Protocol Buffer (.proto) definition files
    │   └── auth.proto  # Authentication service API definitions
    └── proto_gen/      # Generated Go code from protobuf
        └── shared/
            └── proto/
                ├── auth.pb.go       # Generated message types
                └── auth_grpc.pb.go  # Generated gRPC service code
```

## Purpose

The `api/shared/` folder organizes all protobuf definitions and generated code for shared services:

- **Better organization**: Groups related proto files together
- **Scalability**: Easy to add versioned APIs (e.g., `api/v1/`, `api/v2/`)
- **Clear separation**: Distinguishes between different API types if needed

## Generating Protobuf Code

### Using Script (Recommended)
```bash
.\scripts\generate-proto.bat
```

### Using Makefile
```bash
make proto
```

### Manual Generation
```bash
protoc --go_out=api/shared/proto_gen --go_opt=paths=source_relative \
    --go-grpc_out=api/shared/proto_gen --go-grpc_opt=paths=source_relative \
    api/shared/proto/auth.proto
```

## Import in Go Code

```go
import (
    pb "github.com/goconnect/api/shared/proto_gen/shared/proto"
)

// Use in your code
func main() {
    client := pb.NewAuthServiceClient(conn)
    // ...
}
```

## Adding New API Definitions

1. Create `.proto` file in `api/shared/proto/`
2. Run generation: `.\scripts\generate-proto.bat`
3. Import in your service code
4. Commit both `.proto` and generated files

## Future Expansion

This structure allows for:

- **Versioned APIs**: `api/v1/`, `api/v2/`
- **Service-specific APIs**: `api/auth/`, `api/user/`
- **Third-party integrations**: `api/external/`
- **GraphQL schemas**: `api/graphql/`

## Best Practices

1. **Version your APIs** - Use semantic versioning for breaking changes
2. **Document your protobuf** - Add comments to `.proto` files
3. **Commit generated code** - Keep `proto_gen/` in version control
4. **Regenerate after changes** - Always run proto generation after editing `.proto`
5. **Backward compatibility** - Avoid breaking changes when possible

## References

- [Protocol Buffers Documentation](https://protobuf.dev/)
- [gRPC Go Quickstart](https://grpc.io/docs/languages/go/quickstart/)
- [Buf Documentation](https://buf.build/docs/) - Advanced protobuf tooling
