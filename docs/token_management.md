How is token/credential generated and managed in etcd?
======
<span style="color: #808080; font-family: Babas; font-size: 1em;">
ahrtr@github <br>
January 23, 2023
</span>

# Table of Contents
- **[Background](#background)**
- **[How is each user credentials persisted](#how-is-each-user-credentials-persisted)**
- **[How is each token generated](#how-is-each-token-generated)**
  - [Simple Token](#simple-token)
  - [JWT Token](#jwt-token)
- **[How is the credentials and token exchanged/transported](#how-is-the-credentials-and-token-exchangedtransported)**
- **[How is the token used and managed](#how-is-the-token-used-and-managed)**
  - [Simple Token](#simple-token-1)
  - [JWT Token](#jwt-token-1)
- **[Possible minor improvement](#possible-minor-improvement)**

# Background
etcd needs to handle token/credentials only when [auth](https://etcd.io/docs/v3.5/learning/design-auth-v3/#design-and-implementation) is enabled.

This post describes how etcd generates and manages the tokens. There are two kinds of tokens, which are simple and JWT. The simple token is designed
for development testing, so it isn't recommended to use simple token in production use cases. Please use JWT token in production. Use the flag `--auth-token`
to specify the token type, and the valid options are "simple" or "jwt".

# How is each user credentials persisted?
No matter which token type you are going to use, you need to setup [RBAC](https://etcd.io/docs/v3.5/op-guide/authentication/rbac/) firstly.
Creating users using command `etcdctl user add` or client SDK. Specifically, 
you need to set both username and password when creating a user. 

etcd uses [bcrypt.GenerateFromPassword](https://pkg.go.dev/golang.org/x/crypto/bcrypt@v0.0.0-20220525230936-793ad666bf5e#GenerateFromPassword)
(see also [v3_server.go#L490](https://github.com/etcd-io/etcd/blob/ee566c492bb2e0962068a531666c68e1c39d3723/server/etcdserver/v3_server.go#L490))
to generate a bcrypt hash of the password at the given cost specified in `--bcrypt-cost`. Eventually the **hashed password is persisted in the
[bbolt](https://github.com/etcd-io/bbolt/) database**.

Note if [using TLS CN based auth](https://etcd.io/docs/v3.5/op-guide/authentication/rbac/#using-tls-common-name), then no password is needed;
accordingly nothing to save in this case.

# How is each token generated?
etcd server generates a token when it successfully authenticates a user. The client side needs to provide both username and password to authenticate a user. 

When authenticating a user, etcd server uses 
[bcrypt.CompareHashAndPassword](https://pkg.go.dev/golang.org/x/crypto/bcrypt@v0.0.0-20220525230936-793ad666bf5e#CompareHashAndPassword)
(see also [store.go#L375](https://github.com/etcd-io/etcd/blob/ee566c492bb2e0962068a531666c68e1c39d3723/server/auth/store.go#L375))
to compare a bcrypt hashed password with the given plaintext password. If both the username and password are correct, then etcd server generates a token per 
the token type you choose.

## Simple Token
If simple token is configured, etcd server generates a simple token
in the format [`fmt.Sprintf("%s.%d", simpleTokenPrefix, index)`](https://github.com/etcd-io/etcd/blob/ee566c492bb2e0962068a531666c68e1c39d3723/server/auth/simple_token.go#L222).
The `simpleTokenPrefix` is [a random string of 16 bytes](https://github.com/etcd-io/etcd/blob/ee566c492bb2e0962068a531666c68e1c39d3723/server/auth/simple_token.go#L110-L123), and
the `index` is just the current [consistent_index](https://github.com/etcd-io/etcd/blob/ee566c492bb2e0962068a531666c68e1c39d3723/server/etcdserver/apply/apply.go#L293).

Note **simple tokens are not cryptographically signed**.

## JWT Token
Note you need to start etcd with flags something like below to use JWT token,
```
--auth-token=jwt,ttl=300s,pub-key=/srv/jwt_RS256.pub,priv-key=/srv/jwt_RS256,sign-method=RS256
```

If JWT token is configured, etcd server [generates a JWT token](https://github.com/etcd-io/etcd/blob/21e21fe36b89d6ab51e1d72d3b232e782e21512e/server/auth/jwt.go#L89-L96), 
which is **cryptographically signed** by the provided private key. 
Note etcd depends on [golang-jwt/jwt](https://github.com/golang-jwt/jwt) to generate JWT tokens.

Refer to [#signing-methods-and-key-types](https://github.com/golang-jwt/jwt#signing-methods-and-key-types) to learn more about Signing Methods and Key Types.

# How is the credentials and token exchanged/transported?
When adding a user, the client side populates [AuthUserAddRequest](https://github.com/etcd-io/etcd/blob/ee566c492bb2e0962068a531666c68e1c39d3723/api/etcdserverpb/rpc.pb.go#L4540),
and the request is marshaled at client side and unmarshalled at server side by gRPC automatically. 

When authenticating a user, the client side populates [AuthenticateRequest](https://github.com/etcd-io/etcd/blob/ee566c492bb2e0962068a531666c68e1c39d3723/api/etcdserverpb/rpc.pb.go#L4485),
and the request is marshaled at client side and unmarshalled at server side by gRPC automatically. If the user is successfully authenticated, then the generated token is
returned to the client side via [AuthenticateResponse](https://github.com/etcd-io/etcd/blob/ee566c492bb2e0962068a531666c68e1c39d3723/api/etcdserverpb/rpc.pb.go#L5383), which is
marshaled at server side and unmarshalled at client side by gRPC automatically.

**The transport layer is secured by TLS** if configured, see 
[client.go#L352](https://github.com/etcd-io/etcd/blob/ee566c492bb2e0962068a531666c68e1c39d3723/client/v3/client.go#L352) and
[client.go#L234](https://github.com/etcd-io/etcd/blob/ee566c492bb2e0962068a531666c68e1c39d3723/client/v3/client.go#L234).

# How is the token used and managed?
When auth is enabled and the client already gets a valid token, it needs to get the token included in the context for each following gRPC request.
See [credentials.go#L118](https://github.com/etcd-io/etcd/blob/ee566c492bb2e0962068a531666c68e1c39d3723/client/v3/credentials/credentials.go#L118)
and [client.go#L296](https://github.com/etcd-io/etcd/blob/ee566c492bb2e0962068a531666c68e1c39d3723/client/v3/client.go#L296) for client side. 
See [store.go#L1044-L1058](https://github.com/etcd-io/etcd/blob/ee566c492bb2e0962068a531666c68e1c39d3723/server/auth/store.go#L1044-L1058) for server side.

Note if [using TLS CN based auth](https://etcd.io/docs/v3.5/op-guide/authentication/rbac/#using-tls-common-name), etcd server 
[gets the username from the Common Name in peer's certificate](https://github.com/etcd-io/etcd/blob/ee566c492bb2e0962068a531666c68e1c39d3723/server/etcdserver/v3_server.go#L913).

etcd server parses and manages the token differently per the token type.

## Simple Token
Note the **simple tokens are stateful**,
etcd [caches all token-username mappings in memory](https://github.com/etcd-io/etcd/blob/ee566c492bb2e0962068a531666c68e1c39d3723/server/auth/simple_token.go#L141).
The lifetime of each simple token is specified by `--auth-token-ttl`, which defaults to 300 seconds. etcd removes a token from the cache when it expires.

When etcd server receives a simple token, it just checks whether the token exists in the cache, and regards it as valid if it exists.

## JWT Token
The **JWT tokens are stateless**, and etcd depends on [jwt.Parse](https://github.com/golang-jwt/jwt/blob/v4.4.3/token.go#L98) to
parse, validate, verify the signature of a JWT token and return the parsed token. Each JWT token's lifetime is specified in
`--auth-token=jwt,ttl=300s,...`, and defaults to 300 seconds.

# Possible minor improvement
Note `--auth-token-ttl` is only used by simple tokens. Personally I think it should be used by the JWT token as well. In order to be 
backward compatible, JWT should use TTL below in descending priority order,
- `--auth-token=jwt,ttl=300s,...`
- `--auth-token-ttl`
- The default value 300s.

Please anyone feel free to raise an issue in etcd community and gets it resolved.
