# \TagsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddTagPublicKey**](TagsApi.md#AddTagPublicKey) | **Post** /api/sshkeys/public-keys/{fingerprint}/tags | Add tag public key
[**CreateDeviceTag**](TagsApi.md#CreateDeviceTag) | **Post** /api/devices/{uid}/tags | Create a tag
[**DeleteDeviceTag**](TagsApi.md#DeleteDeviceTag) | **Delete** /api/devices/{uid}/tags/{tag} | Delete a tag from device
[**DeleteTag**](TagsApi.md#DeleteTag) | **Delete** /api/tags/{tag} | Delete a tag name.
[**GetTags**](TagsApi.md#GetTags) | **Get** /api/tags | Get tags
[**RemoveTagPublicKey**](TagsApi.md#RemoveTagPublicKey) | **Delete** /api/sshkeys/public-keys/{fingerprint}/tags/{tag} | Remove tag public key
[**RenameTag**](TagsApi.md#RenameTag) | **Put** /api/tags/{tag} | Rename a tag name.
[**UpdateTagsDevice**](TagsApi.md#UpdateTagsDevice) | **Put** /api/devices/{uid}/tags | Update tags to device
[**UpdateTagsPublicKey**](TagsApi.md#UpdateTagsPublicKey) | **Put** /api/sshkeys/public-keys/{fingerprint}/tags | Update tags public key



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
    resp, r, err := apiClient.TagsApi.AddTagPublicKey(context.Background(), fingerprint).AddTagPublicKeyRequest(addTagPublicKeyRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TagsApi.AddTagPublicKey``: %v\n", err)
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


## CreateDeviceTag

> CreateDeviceTag(ctx, uid).CreateDeviceTagRequest(createDeviceTagRequest).Execute()

Create a tag



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
    uid := "uid_example" // string | Device's UID
    createDeviceTagRequest := *openapiclient.NewCreateDeviceTagRequest("tag") // CreateDeviceTagRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TagsApi.CreateDeviceTag(context.Background(), uid).CreateDeviceTagRequest(createDeviceTagRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TagsApi.CreateDeviceTag``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uid** | **string** | Device&#39;s UID | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateDeviceTagRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **createDeviceTagRequest** | [**CreateDeviceTagRequest**](CreateDeviceTagRequest.md) |  | 

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


## DeleteDeviceTag

> DeleteDeviceTag(ctx, uid, tag).Execute()

Delete a tag from device



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
    uid := "uid_example" // string | Device's UID
    tag := "tag_example" // string | Device's tag name

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TagsApi.DeleteDeviceTag(context.Background(), uid, tag).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TagsApi.DeleteDeviceTag``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uid** | **string** | Device&#39;s UID | 
**tag** | **string** | Device&#39;s tag name | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDeviceTagRequest struct via the builder pattern


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


## DeleteTag

> DeleteTag(ctx, tag).Execute()

Delete a tag name.

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
    tag := "tag_example" // string | Tag's name.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TagsApi.DeleteTag(context.Background(), tag).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TagsApi.DeleteTag``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tag** | **string** | Tag&#39;s name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteTagRequest struct via the builder pattern


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


## GetTags

> []string GetTags(ctx).Execute()

Get tags

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
    resp, r, err := apiClient.TagsApi.GetTags(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TagsApi.GetTags``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetTags`: []string
    fmt.Fprintf(os.Stdout, "Response from `TagsApi.GetTags`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetTagsRequest struct via the builder pattern


### Return type

**[]string**

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
    resp, r, err := apiClient.TagsApi.RemoveTagPublicKey(context.Background(), fingerprint, tag).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TagsApi.RemoveTagPublicKey``: %v\n", err)
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


## RenameTag

> RenameTag(ctx, tag).RenameTagRequest(renameTagRequest).Execute()

Rename a tag name.

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
    tag := "tag_example" // string | Tag's name.
    renameTagRequest := *openapiclient.NewRenameTagRequest() // RenameTagRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TagsApi.RenameTag(context.Background(), tag).RenameTagRequest(renameTagRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TagsApi.RenameTag``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tag** | **string** | Tag&#39;s name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiRenameTagRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **renameTagRequest** | [**RenameTagRequest**](RenameTagRequest.md) |  | 

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


## UpdateTagsDevice

> UpdateTagsDevice(ctx, uid).UpdateTagsDeviceRequest(updateTagsDeviceRequest).Execute()

Update tags to device



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
    uid := "uid_example" // string | Device's UID
    updateTagsDeviceRequest := *openapiclient.NewUpdateTagsDeviceRequest([]string{"tag"}) // UpdateTagsDeviceRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TagsApi.UpdateTagsDevice(context.Background(), uid).UpdateTagsDeviceRequest(updateTagsDeviceRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TagsApi.UpdateTagsDevice``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uid** | **string** | Device&#39;s UID | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateTagsDeviceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateTagsDeviceRequest** | [**UpdateTagsDeviceRequest**](UpdateTagsDeviceRequest.md) |  | 

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
    resp, r, err := apiClient.TagsApi.UpdateTagsPublicKey(context.Background(), fingerprint).UpdateTagsPublicKeyRequest(updateTagsPublicKeyRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TagsApi.UpdateTagsPublicKey``: %v\n", err)
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

