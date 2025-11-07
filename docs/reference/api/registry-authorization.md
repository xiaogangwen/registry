# Registry Authorization Specification

A standardized authentication and authorization mechanism for MCP registries, based on the [MCP Authorization Specification](https://modelcontextprotocol.io/specification/draft/basic/authorization).

**Scope:** This specification is intended for downstream sub-registries (e.g., private company registries). The official modelcontextprotocol.io registry remains public for reading, and uses a different auth system for publishing for legacy reasons (this may change in future).

## Architecture

Sub-registries implementing this specification act as OAuth 2.1 Resource Servers, identical to how MCP servers work:

- **Registry** = OAuth 2.1 Resource Server (validates access tokens)
- **MCP Client** = OAuth 2.1 Client (requests registry resources with tokens)
- **Authorization Server** = Issues tokens and manages authentication

This allows clients to reuse their existing MCP authorization implementation without modification.

## Client Requirements

Clients MUST follow the [MCP Authorization Specification](https://modelcontextprotocol.io/specification/draft/basic/authorization) when authenticating to registries. The flow is identical to MCP server authentication.

### Discovery

Registries requiring auth SHOULD return discovery information via a 401 Unauthorized response with the WWW-Authenticate header:

```http
GET /v0/servers HTTP/1.1
Host: registry.example.com

HTTP/1.1 401 Unauthorized
WWW-Authenticate: Bearer realm="MCP Registry", scope="registry:read", resource_metadata="https://registry.example.com/.well-known/oauth-protected-resource"
```

(but for compatibility with the MCP spec, registries MAY serve metadata at a well-known URI)

Clients supporting auth MUST then retrieve the Protected Resource Metadata from the provided URL:

```json
{
  "resource": "https://registry.example.com",
  "authorization_servers": ["https://auth.example.com"],
  "scopes_supported": ["registry:read", "registry:write"],
  "bearer_methods_supported": ["header"]
}
```

### Authorization Server Discovery

Clients MUST discover the authorization server using standard OAuth 2.0 / OpenID Connect well-known endpoints, following the same priority order specified in the MCP Authorization Specification.

### Token Requests

Clients MUST include the `resource` parameter (RFC 8707) identifying the registry in all token requests:

```http
POST /token HTTP/1.1
Host: auth.example.com
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code
&code=AUTHORIZATION_CODE
&redirect_uri=http://localhost:3000/callback
&client_id=CLIENT_ID
&code_verifier=CODE_VERIFIER
&resource=https%3A%2F%2Fregistry.example.com
```

Clients MUST use PKCE with the S256 challenge method.

### Authenticated Requests

Clients MUST include the access token in the Authorization header for all authenticated requests:

```http
GET /v0/servers HTTP/1.1
Host: registry.example.com
Authorization: Bearer eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9...
```

Clients MUST NOT include tokens in query strings.

### Scope Selection

Clients MUST follow this priority hierarchy for scope selection:

1. Use scopes from the `WWW-Authenticate` header's `scope` parameter if provided
2. Use all scopes from `scopes_supported` in Protected Resource Metadata if available
3. Omit the scope parameter if neither option exists

### Step-Up Authorization

When clients receive a 403 response with `error="insufficient_scope"`, clients SHOULD:

1. Parse the required scopes from the `scope` parameter in the WWW-Authenticate header
2. Initiate a new authorization flow with the expanded scope set
3. Retry the original request with the new token

Clients MUST implement retry limits to prevent infinite loops.

## Security Requirements

Registries and clients implementing this specification MUST follow all security requirements defined in the [MCP Authorization Specification](https://modelcontextprotocol.io/specification/draft/basic/authorization).

### Scopes

Registries SHOULD use scopes aligned with their API operations:

- `registry:read` - List and read server metadata
- `registry:write` - Publish and update servers
- `registry:admin` - Administrative operations (optional)

Registries MAY define additional scopes for fine-grained access control (e.g., `registry:read:acme-internal` for organization-specific private servers).

### Permission-Based Server Visibility

Registries MAY implement permission-based filtering of server lists. When a client provides an access token, registries SHOULD filter the response based on the token's claims, excluding servers the user does not have permission to access.

Example approach:

```json
{
  "name": "io.github.acme/internal-tools",
  "title": "ACME Internal Tools",
  "visibility": "private",
  "allowed_scopes": ["registry:read:acme-internal"]
}
```

When listing servers, the registry filters results based on the token's scope claims. Users without appropriate scopes do not see private servers in responses.

### Step-Up Authorization Responses

When a user attempts an operation requiring additional permissions, registries SHOULD return a 403 response with:

```http
HTTP/1.1 403 Forbidden
WWW-Authenticate: Bearer error="insufficient_scope", scope="registry:read registry:write", resource_metadata="https://registry.example.com/.well-known/oauth-protected-resource"
```

## Benefits

**For clients:** Zero new code required - reuse existing MCP authorization implementation with identical flow and libraries.

**For sub-registries:** Industry-standard OAuth 2.1 with flexible permissions and proven security model.

**For users:** Consistent login experience across MCP servers and registries.

## Comparison with Official Registry Authentication

The official modelcontextprotocol.io registry uses a custom JWT-based authentication system for publishing servers. Reading servers remains public and does not require authentication.

We may in future look at moving the official registry to this standardized OAuth 2.1-based system for consistency.

## References

- [MCP Authorization Specification](https://modelcontextprotocol.io/specification/draft/basic/authorization)
- [OAuth 2.1 Authorization Framework](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1)
- [RFC 8707: Resource Indicators for OAuth 2.0](https://datatracker.ietf.org/doc/html/rfc8707)
- [RFC 7591: OAuth 2.0 Dynamic Client Registration](https://datatracker.ietf.org/doc/html/rfc7591)
- [Issue #751](https://github.com/modelcontextprotocol/registry/issues/751)
