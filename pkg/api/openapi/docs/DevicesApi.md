# \DevicesApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateDeviceTag**](DevicesApi.md#CreateDeviceTag) | **Post** /api/devices/{uid}/tags | Create a tag
[**DeleteDevice**](DevicesApi.md#DeleteDevice) | **Delete** /api/devices/{uid} | Delete device
[**DeleteDeviceTag**](DevicesApi.md#DeleteDeviceTag) | **Delete** /api/devices/{uid}/tags/{tag} | Delete a tag from device
[**GetDevice**](DevicesApi.md#GetDevice) | **Get** /api/devices/{uid} | Get device
[**GetDevices**](DevicesApi.md#GetDevices) | **Get** /api/devices | Get devices
[**GetStatusDevices**](DevicesApi.md#GetStatusDevices) | **Get** /api/stats | Get stats ShellHub instance
[**PostAuthDevice**](DevicesApi.md#PostAuthDevice) | **Post** /api/devices/auth | Auth device
[**PostAuthDevice_0**](DevicesApi.md#PostAuthDevice_0) | **Post** /api/auth/device | Auth device
[**UpdateDeviceName**](DevicesApi.md#UpdateDeviceName) | **Patch** /api/devices/{uid} | Update device name
[**UpdateDeviceStatus**](DevicesApi.md#UpdateDeviceStatus) | **Patch** /api/devices/{uid}/{status} | Update device status
[**UpdateTagsDevice**](DevicesApi.md#UpdateTagsDevice) | **Put** /api/devices/{uid}/tags | Update tags to device



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
    resp, r, err := apiClient.DevicesApi.CreateDeviceTag(context.Background(), uid).CreateDeviceTagRequest(createDeviceTagRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.CreateDeviceTag``: %v\n", err)
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


## DeleteDevice

> DeleteDevice(ctx, uid).Execute()

Delete device



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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.DeleteDevice(context.Background(), uid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.DeleteDevice``: %v\n", err)
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

Other parameters are passed through a pointer to a apiDeleteDeviceRequest struct via the builder pattern


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
    resp, r, err := apiClient.DevicesApi.DeleteDeviceTag(context.Background(), uid, tag).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.DeleteDeviceTag``: %v\n", err)
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


## GetDevice

> Device GetDevice(ctx, uid).Execute()

Get device



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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.GetDevice(context.Background(), uid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.GetDevice``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDevice`: Device
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.GetDevice`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uid** | **string** | Device&#39;s UID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDeviceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Device**](Device.md)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDevices

> []Device GetDevices(ctx).Filter(filter).Page(page).PerPage(perPage).Status(status).SortBy(sortBy).OrderBy(orderBy).Execute()

Get devices



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
    filter := "W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lIiwib3BlcmF0b3IiOiJjb250YWlucyIsInZhbHVlIjoiZDAifX1d
" // string | Device's filter   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called `type`, it will filter by a `property` called `name` where the value should `contains` `linux`.  If you want get only Devices name as `Linux`, the JSON object will looks like this   ```json   [     {       \"type\":\"property\",       \"params\":         {           \"name\":\"name\",           \"operator\":\"contains\",           \"value\":\"linux\"         }     }   ] ```  So, the output encoded string will result on: `W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lIiwib3BlcmF0b3IiOiJjb250YWlucyIsInZhbHVlIjoiZDAifX1d`  (optional)
    page := int32(56) // int32 | Pagination page number (optional) (default to 1)
    perPage := int32(56) // int32 | Pagination items per page (optional) (default to 10)
    status := openapiclient.deviceStatus("accepted") // DeviceStatus | Device's status (optional)
    sortBy := "name" // string | Device's property to sort of (optional)
    orderBy := "asc" // string |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.GetDevices(context.Background()).Filter(filter).Page(page).PerPage(perPage).Status(status).SortBy(sortBy).OrderBy(orderBy).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.GetDevices``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDevices`: []Device
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.GetDevices`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetDevicesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filter** | **string** | Device&#39;s filter   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called &#x60;type&#x60;, it will filter by a &#x60;property&#x60; called &#x60;name&#x60; where the value should &#x60;contains&#x60; &#x60;linux&#x60;.  If you want get only Devices name as &#x60;Linux&#x60;, the JSON object will looks like this   &#x60;&#x60;&#x60;json   [     {       \&quot;type\&quot;:\&quot;property\&quot;,       \&quot;params\&quot;:         {           \&quot;name\&quot;:\&quot;name\&quot;,           \&quot;operator\&quot;:\&quot;contains\&quot;,           \&quot;value\&quot;:\&quot;linux\&quot;         }     }   ] &#x60;&#x60;&#x60;  So, the output encoded string will result on: &#x60;W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lIiwib3BlcmF0b3IiOiJjb250YWlucyIsInZhbHVlIjoiZDAifX1d&#x60;  | 
 **page** | **int32** | Pagination page number | [default to 1]
 **perPage** | **int32** | Pagination items per page | [default to 10]
 **status** | [**DeviceStatus**](DeviceStatus.md) | Device&#39;s status | 
 **sortBy** | **string** | Device&#39;s property to sort of | 
 **orderBy** | **string** |  | 

### Return type

[**[]Device**](Device.md)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetStatusDevices

> GetStatusDevices200Response GetStatusDevices(ctx).Execute()

Get stats ShellHub instance



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
    resp, r, err := apiClient.DevicesApi.GetStatusDevices(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.GetStatusDevices``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetStatusDevices`: GetStatusDevices200Response
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.GetStatusDevices`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetStatusDevicesRequest struct via the builder pattern


### Return type

[**GetStatusDevices200Response**](GetStatusDevices200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PostAuthDevice

> PostAuthDevice200Response PostAuthDevice(ctx).XRealIP(xRealIP).PostAuthDeviceRequest(postAuthDeviceRequest).Execute()

Auth device



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
    xRealIP := "127.0.0.1" // string |  (optional)
    postAuthDeviceRequest := *openapiclient.NewPostAuthDeviceRequest(*openapiclient.NewDeviceInfo(), "Hostname_example", "PublicKey_example", "3dd0d1f8-8246-4519-b11a-a3dd33717f65") // PostAuthDeviceRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.PostAuthDevice(context.Background()).XRealIP(xRealIP).PostAuthDeviceRequest(postAuthDeviceRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.PostAuthDevice``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `PostAuthDevice`: PostAuthDevice200Response
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.PostAuthDevice`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiPostAuthDeviceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xRealIP** | **string** |  | 
 **postAuthDeviceRequest** | [**PostAuthDeviceRequest**](PostAuthDeviceRequest.md) |  | 

### Return type

[**PostAuthDevice200Response**](PostAuthDevice200Response.md)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PostAuthDevice_0

> PostAuthDevice200Response PostAuthDevice_0(ctx).XRealIP(xRealIP).PostAuthDeviceRequest(postAuthDeviceRequest).Execute()

Auth device



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
    xRealIP := "127.0.0.1" // string |  (optional)
    postAuthDeviceRequest := *openapiclient.NewPostAuthDeviceRequest(*openapiclient.NewDeviceInfo(), "Hostname_example", "PublicKey_example", "3dd0d1f8-8246-4519-b11a-a3dd33717f65") // PostAuthDeviceRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.PostAuthDevice_0(context.Background()).XRealIP(xRealIP).PostAuthDeviceRequest(postAuthDeviceRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.PostAuthDevice_0``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `PostAuthDevice_0`: PostAuthDevice200Response
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.PostAuthDevice_0`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiPostAuthDevice_1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xRealIP** | **string** |  | 
 **postAuthDeviceRequest** | [**PostAuthDeviceRequest**](PostAuthDeviceRequest.md) |  | 

### Return type

[**PostAuthDevice200Response**](PostAuthDevice200Response.md)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateDeviceName

> UpdateDeviceName(ctx, uid).UpdateDeviceNameRequest(updateDeviceNameRequest).Execute()

Update device name



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
    updateDeviceNameRequest := *openapiclient.NewUpdateDeviceNameRequest("example") // UpdateDeviceNameRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.UpdateDeviceName(context.Background(), uid).UpdateDeviceNameRequest(updateDeviceNameRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.UpdateDeviceName``: %v\n", err)
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

Other parameters are passed through a pointer to a apiUpdateDeviceNameRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateDeviceNameRequest** | [**UpdateDeviceNameRequest**](UpdateDeviceNameRequest.md) |  | 

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


## UpdateDeviceStatus

> UpdateDeviceStatus(ctx, uid, status).Execute()

Update device status



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
    status := "accept" // string | Device's status

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.UpdateDeviceStatus(context.Background(), uid, status).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.UpdateDeviceStatus``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uid** | **string** | Device&#39;s UID | 
**status** | **string** | Device&#39;s status | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDeviceStatusRequest struct via the builder pattern


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
    resp, r, err := apiClient.DevicesApi.UpdateTagsDevice(context.Background(), uid).UpdateTagsDeviceRequest(updateTagsDeviceRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.UpdateTagsDevice``: %v\n", err)
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

