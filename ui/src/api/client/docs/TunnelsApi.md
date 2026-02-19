# TunnelsApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createTunnel**](#createtunnel) | **POST** /api/devices/{uid}/tunnels | Create a tunnel|
|[**deleteTunnel**](#deletetunnel) | **DELETE** /api/devices/{uid}/tunnels/{address} | Delete a tunnel|
|[**listTunnels**](#listtunnels) | **GET** /api/devices/{uid}/tunnels | List tunnels|

# **createTunnel**
> Tunnel createTunnel(createTunnelRequest)

Creates a new tunnel for a device.

### Example

```typescript
import {
    TunnelsApi,
    Configuration,
    CreateTunnelRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new TunnelsApi(configuration);

let uid: string; //Device\'s UID (default to undefined)
let createTunnelRequest: CreateTunnelRequest; //

const { status, data } = await apiInstance.createTunnel(
    uid,
    createTunnelRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createTunnelRequest** | **CreateTunnelRequest**|  | |
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|


### Return type

**Tunnel**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Tunnel created successfully. |  -  |
|**400** | Bad request |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteTunnel**
> deleteTunnel()

Deletes a tunnel for a specific device and port.

### Example

```typescript
import {
    TunnelsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new TunnelsApi(configuration);

let uid: string; //Device\'s UID (default to undefined)
let address: string; //Tunnel\'s address (default to undefined)

const { status, data } = await apiInstance.deleteTunnel(
    uid,
    address
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|
| **address** | [**string**] | Tunnel\&#39;s address | defaults to undefined|


### Return type

void (empty response body)

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Tunnel deleted successfully. |  -  |
|**400** | Bad request |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listTunnels**
> Array<Tunnel> listTunnels()

List the tunnels per devices.

### Example

```typescript
import {
    TunnelsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new TunnelsApi(configuration);

let uid: string; //Device\'s UID (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.listTunnels(
    uid,
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|


### Return type

**Array<Tunnel>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get the tunnels. |  * X-Total-Count -  <br>  |
|**400** | Bad request |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

