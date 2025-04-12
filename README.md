## Generic REST client

Provides wrappers around http.Client to simplify REST request preparation & execution.

### Usage
#### 1. Import lib
```shell
go get github.com/slink-go/rc
```
or
```shell
go get go.slink.ws/rc
```
#### 2. Create client
The only required argument to `NewClient(...)` is `WithBaseUrl(...)`. Other acceptable are `WithUserAgent(value string)` and 
`WithHttpClient(custom *http.Client)`. If non-required options are not passed, default values will be used.
```go
cl, err := NewClient(
    WithBaseUrl("https://test.com"),
    WithUserAgent("test-agent"),
)
```

#### 3. Prepare request
Available options:
- `WithMethod` (default one is `http.MethodGet`)
- `WithPath` (deafult is `""`). Can be mentioned multiple times. All values will be joined with `/`.
- `WithQueryParam` to set single value for a key
- `WithQueryParams` to set multiple value for a key
- `WithBody` to set request body
```go
req, err := cl.NewRequest(
    WithMethod(http.MethodHead),
    WithQueryPath("/endpoint"),
    WithQueryPath("sub"),
    WithQueryPath("path"),
    WithQueryParam("k1", "v1"),
    WithQueryParam("k2", "v2"),
)
```
#### 4. Execute request
Execute request with `client.Do(ctx, req, acc)`. `Acc` here is response accumulator. Can be nil. If set, library 
will try to convert response to it. If it is `*io.Writer`, response buffer will be copied to it. 
```go
var b bytes.Buffer
w := bufio.NewWriter(&b)
if _, err := cl.Do(context.Background(), req, w); err != nil {
    return nil, err
}
```
