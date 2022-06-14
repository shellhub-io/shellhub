# \SshApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddTagPublicKey**](SshApi.md#AddTagPublicKey) | **Post** /api/sshkeys/public-keys/{fingerprint}/tags | Add tag public key
[**CreatePublicKey**](SshApi.md#CreatePublicKey) | **Post** /api/sshkeys/public-keys | Create public key
[**DeletePublicKey**](SshApi.md#DeletePublicKey) | **Delete** /api/sshkeys/public-keys/{fingerprint} | Delete public key
[**GetPublicKeys**](SshApi.md#GetPublicKeys) | **Get** /api/sshkeys/public-keys | Get public keys
[**RemoveTagPublicKey**](SshApi.md#RemoveTagPublicKey) | **Delete** /api/sshkeys/public-keys/{fingerprint}/tags/{tag} | Remove tag public key
[**SetSSHKey**](SshApi.md#SetSSHKey) | **Post** /api/auth/ssh | Set SSH key
[**UpdatePublicKey**](SshApi.md#UpdatePublicKey) | **Put** /api/sshkeys/public-keys/{fingerprint} | Update public key
[**UpdateTagsPublicKey**](SshApi.md#UpdateTagsPublicKey) | **Put** /api/sshkeys/public-keys/{fingerprint}/tags | Update tags public key



## AddTagPublicKey

> AddTagPublicKey(ctx, fingerprint).AddTagPublicKeyRequest(addTagPublicKeyRequest).Execute()

Add tag public key



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
    fingerprint := "fingerprint_example" // string | Public key's fingerprint.
    addTagPublicKeyRequest := *openapiclient.NewAddTagPublicKeyRequest("tag") // AddTagPublicKeyRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SshApi.AddTagPublicKey(context.Background(), fingerprint).AddTagPublicKeyRequest(addTagPublicKeyRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SshApi.AddTagPublicKey``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fingerprint** | **string** | Public key&#39;s fingerprint. | 

### Other Parameters

Other parameters are passed through a pointer to a apiAddTagPublicKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **addTagPublicKeyRequest** | [**AddTagPublicKeyRequest**](AddTagPublicKeyRequest.md) |  | 

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


## CreatePublicKey

> PublicKeyResponse CreatePublicKey(ctx).PublicKeyRequest(publicKeyRequest).Execute()

Create public key



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
    publicKeyRequest := *openapiclient.NewPublicKeyRequest("c3NoLXJzYSBBQUFBQjNOemFDMXljMkVBQUFBREFRQUJBQUFCQVFDWWdqRkNQUWdPejBEZ0VQQUh3blEyMGYzRUlGYjd2SkNtd1YxR25uRTU2K0htaGgyY295c3o5MnZqMW9GeElxQUlKZUZxU3lQNWwzbDZjbkFUVmxhZ2MxR21OQm5vQ0NZSlpicXdOVUFiM3RMTXdiOXBaSGVWMFczWVl4OERBSVVsL2ZYaVVhQTNpQk5BcTFrczFzYjZjbVN1VmYwTVNTSjdoOXU3c2Y2RnkyVmQ0U1FqSGd3YmNvSUY1Q0kyWkZlMEx6NTNWeGQwVlZRZG5ISGNBeldRVFlTMDIxcmVXeG5QR2RRdytmWXpCRWdRMG5sTmFzQXBRc1pVUXRPZ0t4TlNFcVJ0VnJiRUR4WisrTllQaWFuNUdSZ0huZWNUUzBaVGNjZjM4SDZYTms1Qm5XWGlEN2RCWlJBRnZ1UjBkWEF1cU9mYUM3SVl5MVJnS1lkdEsrUnY=", openapiclient.publicKeyFilter{PublicKeyFilterOneOf: openapiclient.NewPublicKeyFilterOneOf(".*")}, "example", ".*") // PublicKeyRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SshApi.CreatePublicKey(context.Background()).PublicKeyRequest(publicKeyRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SshApi.CreatePublicKey``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreatePublicKey`: PublicKeyResponse
    fmt.Fprintf(os.Stdout, "Response from `SshApi.CreatePublicKey`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreatePublicKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **publicKeyRequest** | [**PublicKeyRequest**](PublicKeyRequest.md) |  | 

### Return type

[**PublicKeyResponse**](PublicKeyResponse.md)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeletePublicKey

> DeletePublicKey(ctx, fingerprint).Execute()

Delete public key



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
    fingerprint := "fingerprint_example" // string | Public key's fingerprint.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SshApi.DeletePublicKey(context.Background(), fingerprint).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SshApi.DeletePublicKey``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fingerprint** | **string** | Public key&#39;s fingerprint. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeletePublicKeyRequest struct via the builder pattern


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


## GetPublicKeys

> []PublicKeyResponse GetPublicKeys(ctx).Filter(filter).Page(page).PerPage(perPage).Execute()

Get public keys



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
    filter := string(BYTE_ARRAY_DATA_HERE) // string | Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called `type`, it will filter by a `property` called `confirmed` where the value is `eq` to `true`.   This is a example to filter and get only the confirmed users. ```json  [   {   \"type\": \"property\",   \"params\": {       \"name\": \"confirmed\",       \"operator\": \"eq\",       \"value\": true       }   } ]  ```    So, the output enconded string will result on:     ```WwogICAgewogICAgInR5cGUiOiAicHJvcGVydHkiLAogICAgInBhcmFtcyI6IHsKICAgICAgICAibmFtZSI6ICJjb25maXJtZWQiLAogICAgICAgICJvcGVyYXRvciI6ICJlcSIsCiAgICAgICAgInZhbHVlIjogdHJ1ZQogICAgICAgIH0KICAgIH0KXQ==```  (optional)
    page := int32(56) // int32 | Pagination page number (optional) (default to 1)
    perPage := int32(56) // int32 | Pagination items per page (optional) (default to 10)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SshApi.GetPublicKeys(context.Background()).Filter(filter).Page(page).PerPage(perPage).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SshApi.GetPublicKeys``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetPublicKeys`: []PublicKeyResponse
    fmt.Fprintf(os.Stdout, "Response from `SshApi.GetPublicKeys`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetPublicKeysRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filter** | **string** | Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called &#x60;type&#x60;, it will filter by a &#x60;property&#x60; called &#x60;confirmed&#x60; where the value is &#x60;eq&#x60; to &#x60;true&#x60;.   This is a example to filter and get only the confirmed users. &#x60;&#x60;&#x60;json  [   {   \&quot;type\&quot;: \&quot;property\&quot;,   \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;confirmed\&quot;,       \&quot;operator\&quot;: \&quot;eq\&quot;,       \&quot;value\&quot;: true       }   } ]  &#x60;&#x60;&#x60;    So, the output enconded string will result on:     &#x60;&#x60;&#x60;WwogICAgewogICAgInR5cGUiOiAicHJvcGVydHkiLAogICAgInBhcmFtcyI6IHsKICAgICAgICAibmFtZSI6ICJjb25maXJtZWQiLAogICAgICAgICJvcGVyYXRvciI6ICJlcSIsCiAgICAgICAgInZhbHVlIjogdHJ1ZQogICAgICAgIH0KICAgIH0KXQ&#x3D;&#x3D;&#x60;&#x60;&#x60;  | 
 **page** | **int32** | Pagination page number | [default to 1]
 **perPage** | **int32** | Pagination items per page | [default to 10]

### Return type

[**[]PublicKeyResponse**](PublicKeyResponse.md)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RemoveTagPublicKey

> RemoveTagPublicKey(ctx, fingerprint, tag).Execute()

Remove tag public key



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
    fingerprint := "fingerprint_example" // string | Public key's fingerprint.
    tag := "tag_example" // string | Tag's name.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SshApi.RemoveTagPublicKey(context.Background(), fingerprint, tag).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SshApi.RemoveTagPublicKey``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fingerprint** | **string** | Public key&#39;s fingerprint. | 
**tag** | **string** | Tag&#39;s name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiRemoveTagPublicKeyRequest struct via the builder pattern


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


## SetSSHKey

> SetSSHKey200Response SetSSHKey(ctx).SetSSHKeyRequest(setSSHKeyRequest).Execute()

Set SSH key



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
    setSSHKeyRequest := *openapiclient.NewSetSSHKeyRequest() // SetSSHKeyRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SshApi.SetSSHKey(context.Background()).SetSSHKeyRequest(setSSHKeyRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SshApi.SetSSHKey``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SetSSHKey`: SetSSHKey200Response
    fmt.Fprintf(os.Stdout, "Response from `SshApi.SetSSHKey`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSetSSHKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **setSSHKeyRequest** | [**SetSSHKeyRequest**](SetSSHKeyRequest.md) |  | 

### Return type

[**SetSSHKey200Response**](SetSSHKey200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdatePublicKey

> PublicKeyResponse UpdatePublicKey(ctx, fingerprint).UpdatePublicKeyRequest(updatePublicKeyRequest).Execute()

Update public key



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
    fingerprint := "fingerprint_example" // string | Public key's fingerprint.
    updatePublicKeyRequest := *openapiclient.NewUpdatePublicKeyRequest("example", "example", openapiclient.publicKeyFilter{PublicKeyFilterOneOf: openapiclient.NewPublicKeyFilterOneOf(".*")}) // UpdatePublicKeyRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SshApi.UpdatePublicKey(context.Background(), fingerprint).UpdatePublicKeyRequest(updatePublicKeyRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SshApi.UpdatePublicKey``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdatePublicKey`: PublicKeyResponse
    fmt.Fprintf(os.Stdout, "Response from `SshApi.UpdatePublicKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fingerprint** | **string** | Public key&#39;s fingerprint. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdatePublicKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updatePublicKeyRequest** | [**UpdatePublicKeyRequest**](UpdatePublicKeyRequest.md) |  | 

### Return type

[**PublicKeyResponse**](PublicKeyResponse.md)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateTagsPublicKey

> UpdateTagsPublicKey(ctx, fingerprint).UpdateTagsPublicKeyRequest(updateTagsPublicKeyRequest).Execute()

Update tags public key



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
    fingerprint := "fingerprint_example" // string | Public key's fingerprint.
    updateTagsPublicKeyRequest := *openapiclient.NewUpdateTagsPublicKeyRequest() // UpdateTagsPublicKeyRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SshApi.UpdateTagsPublicKey(context.Background(), fingerprint).UpdateTagsPublicKeyRequest(updateTagsPublicKeyRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SshApi.UpdateTagsPublicKey``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fingerprint** | **string** | Public key&#39;s fingerprint. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateTagsPublicKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateTagsPublicKeyRequest** | [**UpdateTagsPublicKeyRequest**](UpdateTagsPublicKeyRequest.md) |  | 

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

