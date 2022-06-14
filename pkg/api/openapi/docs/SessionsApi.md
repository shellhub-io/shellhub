# \SessionsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetSession**](SessionsApi.md#GetSession) | **Get** /api/sessions/{uid} | Get session
[**GetSessionRecord**](SessionsApi.md#GetSessionRecord) | **Get** /api/users/security | Get session record
[**GetSessions**](SessionsApi.md#GetSessions) | **Get** /api/sessions | Get sessions
[**SetSessionAuthenticationStatus**](SessionsApi.md#SetSessionAuthenticationStatus) | **Post** /api/sessions/{uid} | Set session authentication status
[**SetSessionRecord**](SessionsApi.md#SetSessionRecord) | **Put** /api/users/security/{tenant} | Set session record



## GetSession

> Session GetSession(ctx, uid).Execute()

Get session



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    uid := "uid_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SessionsApi.GetSession(context.Background(), uid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SessionsApi.GetSession``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSession`: Session
    fmt.Fprintf(os.Stdout, "Response from `SessionsApi.GetSession`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uid** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetSessionRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Session**](Session.md)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSessionRecord

> bool GetSessionRecord(ctx).Execute()

Get session record



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SessionsApi.GetSessionRecord(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SessionsApi.GetSessionRecord``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSessionRecord`: bool
    fmt.Fprintf(os.Stdout, "Response from `SessionsApi.GetSessionRecord`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetSessionRecordRequest struct via the builder pattern


### Return type

**bool**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSessions

> []Session GetSessions(ctx).Page(page).PerPage(perPage).Execute()

Get sessions



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    page := int32(56) // int32 | Pagination page number (optional) (default to 1)
    perPage := int32(56) // int32 | Pagination items per page (optional) (default to 10)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SessionsApi.GetSessions(context.Background()).Page(page).PerPage(perPage).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SessionsApi.GetSessions``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSessions`: []Session
    fmt.Fprintf(os.Stdout, "Response from `SessionsApi.GetSessions`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetSessionsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int32** | Pagination page number | [default to 1]
 **perPage** | **int32** | Pagination items per page | [default to 10]

### Return type

[**[]Session**](Session.md)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetSessionAuthenticationStatus

> SetSessionAuthenticationStatus(ctx, uid).SetSessionAuthenticationStatusRequest(setSessionAuthenticationStatusRequest).Execute()

Set session authentication status



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    uid := "uid_example" // string | 
    setSessionAuthenticationStatusRequest := *openapiclient.NewSetSessionAuthenticationStatusRequest() // SetSessionAuthenticationStatusRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SessionsApi.SetSessionAuthenticationStatus(context.Background(), uid).SetSessionAuthenticationStatusRequest(setSessionAuthenticationStatusRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SessionsApi.SetSessionAuthenticationStatus``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uid** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetSessionAuthenticationStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **setSessionAuthenticationStatusRequest** | [**SetSessionAuthenticationStatusRequest**](SetSessionAuthenticationStatusRequest.md) |  | 

### Return type

 (empty response body)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetSessionRecord

> SetSessionRecord(ctx, tenant).SetSessionRecordRequest(setSessionRecordRequest).Execute()

Set session record



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    tenant := "tenant_example" // string | Namespace's tenant ID
    setSessionRecordRequest := *openapiclient.NewSetSessionRecordRequest() // SetSessionRecordRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SessionsApi.SetSessionRecord(context.Background(), tenant).SetSessionRecordRequest(setSessionRecordRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SessionsApi.SetSessionRecord``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenant** | **string** | Namespace&#39;s tenant ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetSessionRecordRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **setSessionRecordRequest** | [**SetSessionRecordRequest**](SetSessionRecordRequest.md) |  | 

### Return type

 (empty response body)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

