# \NamespacesApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddNamespaceMember**](NamespacesApi.md#AddNamespaceMember) | **Post** /api/namespaces/{tenant}/members | Add a member to a namespace
[**CreateNamespace**](NamespacesApi.md#CreateNamespace) | **Post** /api/namespaces/{tenant} | Create namespace
[**DeleteNamespace**](NamespacesApi.md#DeleteNamespace) | **Delete** /api/namespaces/{tenant} | Delete namespace
[**EditNamespace**](NamespacesApi.md#EditNamespace) | **Put** /api/namespaces/{tenant} | Edit namespace
[**GetNamespace**](NamespacesApi.md#GetNamespace) | **Get** /api/namespaces/{tenant} | Get a namespace
[**GetNamespaceToken**](NamespacesApi.md#GetNamespaceToken) | **Get** /api/auth/token/{tenant} | Get a new namespace&#39;s token
[**GetNamespaces**](NamespacesApi.md#GetNamespaces) | **Get** /api/namespaces | Get namespaces list
[**RemoveNamespaceMember**](NamespacesApi.md#RemoveNamespaceMember) | **Delete** /api/namespaces/{tenant}/members/{uid} | Remove a member from a namespace
[**UpdateNamespaceMember**](NamespacesApi.md#UpdateNamespaceMember) | **Patch** /api/namespaces/{tenant}/members/{uid} | Update a member from a namespace



## AddNamespaceMember

> Namespace AddNamespaceMember(ctx, tenant).AddNamespaceMemberRequest(addNamespaceMemberRequest).Execute()

Add a member to a namespace



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
    addNamespaceMemberRequest := *openapiclient.NewAddNamespaceMemberRequest("Username_example", openapiclient.namespaceMemberRole("administrator")) // AddNamespaceMemberRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NamespacesApi.AddNamespaceMember(context.Background(), tenant).AddNamespaceMemberRequest(addNamespaceMemberRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NamespacesApi.AddNamespaceMember``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AddNamespaceMember`: Namespace
    fmt.Fprintf(os.Stdout, "Response from `NamespacesApi.AddNamespaceMember`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenant** | **string** | Namespace&#39;s tenant ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiAddNamespaceMemberRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **addNamespaceMemberRequest** | [**AddNamespaceMemberRequest**](AddNamespaceMemberRequest.md) |  | 

### Return type

[**Namespace**](Namespace.md)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateNamespace

> Namespace CreateNamespace(ctx, tenant).EditNamespaceRequest(editNamespaceRequest).Execute()

Create namespace



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
    editNamespaceRequest := *openapiclient.NewEditNamespaceRequest("examplespace") // EditNamespaceRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NamespacesApi.CreateNamespace(context.Background(), tenant).EditNamespaceRequest(editNamespaceRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NamespacesApi.CreateNamespace``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateNamespace`: Namespace
    fmt.Fprintf(os.Stdout, "Response from `NamespacesApi.CreateNamespace`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenant** | **string** | Namespace&#39;s tenant ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateNamespaceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **editNamespaceRequest** | [**EditNamespaceRequest**](EditNamespaceRequest.md) |  | 

### Return type

[**Namespace**](Namespace.md)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteNamespace

> DeleteNamespace(ctx, tenant).Execute()

Delete namespace



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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NamespacesApi.DeleteNamespace(context.Background(), tenant).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NamespacesApi.DeleteNamespace``: %v\n", err)
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

Other parameters are passed through a pointer to a apiDeleteNamespaceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## EditNamespace

> Namespace EditNamespace(ctx, tenant).EditNamespaceRequest(editNamespaceRequest).Execute()

Edit namespace



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
    editNamespaceRequest := *openapiclient.NewEditNamespaceRequest("examplespace") // EditNamespaceRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NamespacesApi.EditNamespace(context.Background(), tenant).EditNamespaceRequest(editNamespaceRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NamespacesApi.EditNamespace``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `EditNamespace`: Namespace
    fmt.Fprintf(os.Stdout, "Response from `NamespacesApi.EditNamespace`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenant** | **string** | Namespace&#39;s tenant ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiEditNamespaceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **editNamespaceRequest** | [**EditNamespaceRequest**](EditNamespaceRequest.md) |  | 

### Return type

[**Namespace**](Namespace.md)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNamespace

> Namespace GetNamespace(ctx, tenant).Execute()

Get a namespace



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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NamespacesApi.GetNamespace(context.Background(), tenant).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NamespacesApi.GetNamespace``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetNamespace`: Namespace
    fmt.Fprintf(os.Stdout, "Response from `NamespacesApi.GetNamespace`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenant** | **string** | Namespace&#39;s tenant ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNamespaceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Namespace**](Namespace.md)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNamespaceToken

> UserAuth GetNamespaceToken(ctx, tenant).Execute()

Get a new namespace's token



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
    tenant := "tenant_example" // string | Namespace Tenant (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NamespacesApi.GetNamespaceToken(context.Background(), tenant).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NamespacesApi.GetNamespaceToken``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetNamespaceToken`: UserAuth
    fmt.Fprintf(os.Stdout, "Response from `NamespacesApi.GetNamespaceToken`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenant** | **string** | Namespace Tenant | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNamespaceTokenRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**UserAuth**](UserAuth.md)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNamespaces

> []Namespace GetNamespaces(ctx).Filter(filter).Page(page).PerPage(perPage).Execute()

Get namespaces list



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
    filter := "W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lIiwib3BlcmF0b3IiOiJjb250YWlucyIsInZhbHVlIjoiZXhhbXBsZXNwYWNlIn19XQ==" // string | Namespaces's filter.   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called `type`, it will filter by a `property` called `name` where the value should `contains` `examplespace`.  If you want get only Namespaces name as `examplespace`, the JSON object will looks like this   ```json [   {     \"type\":\"property\",     \"params\":{       \"name\":\"name\",       \"operator\":\"contains\",       \"value\":\"examplespace\"     }   } ] ```  So, the output encoded string will result on: `W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lIiwib3BlcmF0b3IiOiJjb250YWlucyIsInZhbHVlIjoiZXhhbXBsZXNwYWNlIn19XQ==`  (optional)
    page := int32(56) // int32 | Pagination page number (optional) (default to 1)
    perPage := int32(56) // int32 | Pagination items per page (optional) (default to 10)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NamespacesApi.GetNamespaces(context.Background()).Filter(filter).Page(page).PerPage(perPage).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NamespacesApi.GetNamespaces``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetNamespaces`: []Namespace
    fmt.Fprintf(os.Stdout, "Response from `NamespacesApi.GetNamespaces`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetNamespacesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filter** | **string** | Namespaces&#39;s filter.   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called &#x60;type&#x60;, it will filter by a &#x60;property&#x60; called &#x60;name&#x60; where the value should &#x60;contains&#x60; &#x60;examplespace&#x60;.  If you want get only Namespaces name as &#x60;examplespace&#x60;, the JSON object will looks like this   &#x60;&#x60;&#x60;json [   {     \&quot;type\&quot;:\&quot;property\&quot;,     \&quot;params\&quot;:{       \&quot;name\&quot;:\&quot;name\&quot;,       \&quot;operator\&quot;:\&quot;contains\&quot;,       \&quot;value\&quot;:\&quot;examplespace\&quot;     }   } ] &#x60;&#x60;&#x60;  So, the output encoded string will result on: &#x60;W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lIiwib3BlcmF0b3IiOiJjb250YWlucyIsInZhbHVlIjoiZXhhbXBsZXNwYWNlIn19XQ&#x3D;&#x3D;&#x60;  | 
 **page** | **int32** | Pagination page number | [default to 1]
 **perPage** | **int32** | Pagination items per page | [default to 10]

### Return type

[**[]Namespace**](Namespace.md)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RemoveNamespaceMember

> Namespace RemoveNamespaceMember(ctx, tenant, uid).Execute()

Remove a member from a namespace



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
    uid := "uid_example" // string | Member's ID

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NamespacesApi.RemoveNamespaceMember(context.Background(), tenant, uid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NamespacesApi.RemoveNamespaceMember``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `RemoveNamespaceMember`: Namespace
    fmt.Fprintf(os.Stdout, "Response from `NamespacesApi.RemoveNamespaceMember`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenant** | **string** | Namespace&#39;s tenant ID | 
**uid** | **string** | Member&#39;s ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiRemoveNamespaceMemberRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Namespace**](Namespace.md)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateNamespaceMember

> UpdateNamespaceMember(ctx, tenant, uid).UpdateNamespaceMemberRequest(updateNamespaceMemberRequest).Execute()

Update a member from a namespace



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
    uid := "uid_example" // string | Member's ID
    updateNamespaceMemberRequest := *openapiclient.NewUpdateNamespaceMemberRequest() // UpdateNamespaceMemberRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NamespacesApi.UpdateNamespaceMember(context.Background(), tenant, uid).UpdateNamespaceMemberRequest(updateNamespaceMemberRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NamespacesApi.UpdateNamespaceMember``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenant** | **string** | Namespace&#39;s tenant ID | 
**uid** | **string** | Member&#39;s ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateNamespaceMemberRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **updateNamespaceMemberRequest** | [**UpdateNamespaceMemberRequest**](UpdateNamespaceMemberRequest.md) |  | 

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

