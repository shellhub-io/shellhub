# DevicesApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**acceptDevice**](#acceptdevice) | **PATCH** /api/devices/{uid}/accept | Accept device|
|[**createTunnel**](#createtunnel) | **POST** /api/devices/{uid}/tunnels | Create a tunnel|
|[**deleteDevice**](#deletedevice) | **DELETE** /api/devices/{uid} | Delete device|
|[**deleteDeviceAdmin**](#deletedeviceadmin) | **DELETE** /admin/api/devices/{uid} | Delete device admin|
|[**deleteTunnel**](#deletetunnel) | **DELETE** /api/devices/{uid}/tunnels/{address} | Delete a tunnel|
|[**getDevice**](#getdevice) | **GET** /api/devices/{uid} | Get device|
|[**getDeviceAdmin**](#getdeviceadmin) | **GET** /admin/api/devices/{uid} | Get device admin|
|[**getDevices**](#getdevices) | **GET** /api/devices | Get devices|
|[**getDevicesAdmin**](#getdevicesadmin) | **GET** /admin/api/devices | Get devices admin|
|[**getStatusDevices**](#getstatusdevices) | **GET** /api/stats | Get stats ShellHub instance|
|[**listTunnels**](#listtunnels) | **GET** /api/devices/{uid}/tunnels | List tunnels|
|[**resolveDevice**](#resolvedevice) | **GET** /api/devices/resolve | Resolve Device|
|[**updateDevice**](#updatedevice) | **PUT** /api/devices/{uid} | Update device|
|[**updateDeviceNameAdmin**](#updatedevicenameadmin) | **PATCH** /admin/api/devices/{uid} | Update device name Admin|
|[**updateDeviceStatus**](#updatedevicestatus) | **PATCH** /api/devices/{uid}/{status} | Update device status|
|[**updateDeviceStatusAdmin**](#updatedevicestatusadmin) | **PATCH** /admin/api/devices/{uid}/{status} | Update status Admin|
|[**updateDeviceStatusOffline**](#updatedevicestatusoffline) | **POST** /internal/devices/{uid}/offline | Update device status to offline|

# **acceptDevice**
> acceptDevice()

Change device status to `accepted`.

### Example

```typescript
import {
    DevicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DevicesApi(configuration);

let uid: string; //Device\'s UID (default to undefined)

const { status, data } = await apiInstance.acceptDevice(
    uid
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|


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
|**200** | Success to accept the device |  -  |
|**401** | Unauthorized |  -  |
|**402** | Payment required |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createTunnel**
> Tunnel createTunnel(createTunnelRequest)

Creates a new tunnel for a device.

### Example

```typescript
import {
    DevicesApi,
    Configuration,
    CreateTunnelRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DevicesApi(configuration);

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

# **deleteDevice**
> deleteDevice()

Delete a device.

### Example

```typescript
import {
    DevicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DevicesApi(configuration);

let uid: string; //Device\'s UID (default to undefined)

const { status, data } = await apiInstance.deleteDevice(
    uid
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|


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
|**200** | Success to delete a device. |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteDeviceAdmin**
> deleteDeviceAdmin()

Delete a device.

### Example

```typescript
import {
    DevicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DevicesApi(configuration);

let uid: string; //Device\'s UID (default to undefined)

const { status, data } = await apiInstance.deleteDeviceAdmin(
    uid
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|


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
|**200** | Success to delete a device. |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteTunnel**
> deleteTunnel()

Deletes a tunnel for a specific device and port.

### Example

```typescript
import {
    DevicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DevicesApi(configuration);

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

# **getDevice**
> Device getDevice()

Get a device.

### Example

```typescript
import {
    DevicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DevicesApi(configuration);

let uid: string; //Device\'s UID (default to undefined)

const { status, data } = await apiInstance.getDevice(
    uid
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|


### Return type

**Device**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success get a device. |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getDeviceAdmin**
> Device getDeviceAdmin()

Get a device.

### Example

```typescript
import {
    DevicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DevicesApi(configuration);

let uid: string; //Device\'s UID (default to undefined)

const { status, data } = await apiInstance.getDeviceAdmin(
    uid
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|


### Return type

**Device**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a device. |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getDevices**
> Array<Device> getDevices()

Get a list of devices.

### Example

```typescript
import {
    DevicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DevicesApi(configuration);

let filter: string; //Filter field receives a JSON object enconded as base64 string for limit a search.  The JSON enconded must follow these interafaces: ```typescript interface ParamProperty {   name: string;   operator: \"contains\" | \"eq\" | \"bool\" | \"gt\" | \"lt\";   value: string; }  interface ParamOperator {   name: \"and\" | \"or\"; }  interface Filter {   type: \"property\" | \"operator\";   param: ParamOperator | ParamProperty; }  interface FilterList {   Filters: Array<Filter>; }  ```  ## Examples  This is a example to filter and get only the resource what property \"confirmed\" is \"true\" ```json [   {   \"type\": \"property\",   \"params\": {       \"name\": \"confirmed\",       \"operator\": \"bool\",       \"value\": \"true\"       }   } ] ```  This one, filter resource by the property \"id\" inside \"info\" structure when it is equal to \"manjaro\" and online property is set to \"true\" ```json [   {     \"type\": \"property\",     \"params\": {       \"name\": \"info.id\",       \"operator\": \"eq\",       \"value\": \"manjaro\"     }   },   {     \"type\": \"property\",     \"params\": {       \"name\": \"online\",       \"operator\": \"bool\",       \"value\": \"true\"     }   },   {     \"type\": \"operator\",     \"params\": {       \"name\": \"and\"     }   } ] ```  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)
let status: DeviceStatus; //Device\'s status (optional) (default to undefined)
let sortBy: string; //Device\'s property to sort of (optional) (default to 'last_seen')
let orderBy: GetDevicesOrderByParameter; // (optional) (default to undefined)

const { status, data } = await apiInstance.getDevices(
    filter,
    page,
    perPage,
    status,
    sortBy,
    orderBy
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **filter** | [**string**] | Filter field receives a JSON object enconded as base64 string for limit a search.  The JSON enconded must follow these interafaces: &#x60;&#x60;&#x60;typescript interface ParamProperty {   name: string;   operator: \&quot;contains\&quot; | \&quot;eq\&quot; | \&quot;bool\&quot; | \&quot;gt\&quot; | \&quot;lt\&quot;;   value: string; }  interface ParamOperator {   name: \&quot;and\&quot; | \&quot;or\&quot;; }  interface Filter {   type: \&quot;property\&quot; | \&quot;operator\&quot;;   param: ParamOperator | ParamProperty; }  interface FilterList {   Filters: Array&lt;Filter&gt;; }  &#x60;&#x60;&#x60;  ## Examples  This is a example to filter and get only the resource what property \&quot;confirmed\&quot; is \&quot;true\&quot; &#x60;&#x60;&#x60;json [   {   \&quot;type\&quot;: \&quot;property\&quot;,   \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;confirmed\&quot;,       \&quot;operator\&quot;: \&quot;bool\&quot;,       \&quot;value\&quot;: \&quot;true\&quot;       }   } ] &#x60;&#x60;&#x60;  This one, filter resource by the property \&quot;id\&quot; inside \&quot;info\&quot; structure when it is equal to \&quot;manjaro\&quot; and online property is set to \&quot;true\&quot; &#x60;&#x60;&#x60;json [   {     \&quot;type\&quot;: \&quot;property\&quot;,     \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;info.id\&quot;,       \&quot;operator\&quot;: \&quot;eq\&quot;,       \&quot;value\&quot;: \&quot;manjaro\&quot;     }   },   {     \&quot;type\&quot;: \&quot;property\&quot;,     \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;online\&quot;,       \&quot;operator\&quot;: \&quot;bool\&quot;,       \&quot;value\&quot;: \&quot;true\&quot;     }   },   {     \&quot;type\&quot;: \&quot;operator\&quot;,     \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;and\&quot;     }   } ] &#x60;&#x60;&#x60;  | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|
| **status** | **DeviceStatus** | Device\&#39;s status | (optional) defaults to undefined|
| **sortBy** | [**string**] | Device\&#39;s property to sort of | (optional) defaults to 'last_seen'|
| **orderBy** | **GetDevicesOrderByParameter** |  | (optional) defaults to undefined|


### Return type

**Array<Device>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a list of devices. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getDevicesAdmin**
> Array<Device> getDevicesAdmin()

Get a list of devices.

### Example

```typescript
import {
    DevicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DevicesApi(configuration);

let filter: string; //Device\'s filter   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called `type`, it will filter by a `property` called `name` where the value should `contains` `linux`.  If you want get only Devices name as `Linux`, the JSON object will looks like this   ```json   [     {       \"type\":\"property\",       \"params\":         {           \"name\":\"name\",           \"operator\":\"contains\",           \"value\":\"linux\"         }     }   ] ```  So, the output encoded string will result on: `W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lIiwib3BlcmF0b3IiOiJjb250YWlucyIsInZhbHVlIjoiZDAifX1d`  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)
let status: DeviceStatus; //Device\'s status (optional) (default to undefined)
let sortBy: string; //Device\'s property to sort of (optional) (default to undefined)
let orderBy: GetDevicesAdminOrderByParameter; // (optional) (default to undefined)

const { status, data } = await apiInstance.getDevicesAdmin(
    filter,
    page,
    perPage,
    status,
    sortBy,
    orderBy
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **filter** | [**string**] | Device\&#39;s filter   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called &#x60;type&#x60;, it will filter by a &#x60;property&#x60; called &#x60;name&#x60; where the value should &#x60;contains&#x60; &#x60;linux&#x60;.  If you want get only Devices name as &#x60;Linux&#x60;, the JSON object will looks like this   &#x60;&#x60;&#x60;json   [     {       \&quot;type\&quot;:\&quot;property\&quot;,       \&quot;params\&quot;:         {           \&quot;name\&quot;:\&quot;name\&quot;,           \&quot;operator\&quot;:\&quot;contains\&quot;,           \&quot;value\&quot;:\&quot;linux\&quot;         }     }   ] &#x60;&#x60;&#x60;  So, the output encoded string will result on: &#x60;W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lIiwib3BlcmF0b3IiOiJjb250YWlucyIsInZhbHVlIjoiZDAifX1d&#x60;  | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|
| **status** | **DeviceStatus** | Device\&#39;s status | (optional) defaults to undefined|
| **sortBy** | [**string**] | Device\&#39;s property to sort of | (optional) defaults to undefined|
| **orderBy** | **GetDevicesAdminOrderByParameter** |  | (optional) defaults to undefined|


### Return type

**Array<Device>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a list of devices. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getStatusDevices**
> GetStatusDevices200Response getStatusDevices()

Get stats ShellHub instance.

### Example

```typescript
import {
    DevicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DevicesApi(configuration);

const { status, data } = await apiInstance.getStatusDevices();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**GetStatusDevices200Response**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success get stats from ShellHub instance. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listTunnels**
> Array<Tunnel> listTunnels()

List the tunnels per devices.

### Example

```typescript
import {
    DevicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DevicesApi(configuration);

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
|**200** | Success to get the tunnels. |  -  |
|**400** | Bad request |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **resolveDevice**
> Device resolveDevice()

Retrieve a device using flexible resolution methods. The device can be identified by either its unique identifier (UID) or hostname. The endpoint automatically scopes results to the authenticated tenant\'s namespace for security isolation.  When both UID and hostname are provided, UID takes precedence over hostname. 

### Example

```typescript
import {
    DevicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DevicesApi(configuration);

let hostname: string; //The UID of the device (optional) (default to undefined)
let uid: string; //The hostname of the device (optional) (default to undefined)

const { status, data } = await apiInstance.resolveDevice(
    hostname,
    uid
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **hostname** | [**string**] | The UID of the device | (optional) defaults to undefined|
| **uid** | [**string**] | The hostname of the device | (optional) defaults to undefined|


### Return type

**Device**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Device successfully retrieved |  -  |
|**401** | Unauthorized |  -  |
|**404** | Device not found with the specified resolver |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateDevice**
> updateDevice()

Update device\'s data.

### Example

```typescript
import {
    DevicesApi,
    Configuration,
    UpdateDeviceRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DevicesApi(configuration);

let uid: string; //Device\'s UID (default to undefined)
let updateDeviceRequest: UpdateDeviceRequest; // (optional)

const { status, data } = await apiInstance.updateDevice(
    uid,
    updateDeviceRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateDeviceRequest** | **UpdateDeviceRequest**|  | |
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|


### Return type

void (empty response body)

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to update device\&#39;s data. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateDeviceNameAdmin**
> updateDeviceNameAdmin()

Update device\'s name.

### Example

```typescript
import {
    DevicesApi,
    Configuration,
    UpdateDeviceNameAdminRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DevicesApi(configuration);

let uid: string; //Device\'s UID (default to undefined)
let updateDeviceNameAdminRequest: UpdateDeviceNameAdminRequest; // (optional)

const { status, data } = await apiInstance.updateDeviceNameAdmin(
    uid,
    updateDeviceNameAdminRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateDeviceNameAdminRequest** | **UpdateDeviceNameAdminRequest**|  | |
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|


### Return type

void (empty response body)

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to update device\&#39;s name. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateDeviceStatus**
> updateDeviceStatus()

Update device\'s status.

### Example

```typescript
import {
    DevicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DevicesApi(configuration);

let uid: string; //Device\'s UID (default to undefined)
let status: UpdateDeviceStatusStatusParameter; //Device\'s status (default to undefined)

const { status, data } = await apiInstance.updateDeviceStatus(
    uid,
    status
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|
| **status** | **UpdateDeviceStatusStatusParameter** | Device\&#39;s status | defaults to undefined|


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
|**200** | Success to update device status. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**402** | Payment required |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateDeviceStatusAdmin**
> updateDeviceStatusAdmin()

Update device\'s status.

### Example

```typescript
import {
    DevicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DevicesApi(configuration);

let uid: string; //Device\'s UID (default to undefined)
let status: DeviceStatus; //Device\'s status (default to undefined)

const { status, data } = await apiInstance.updateDeviceStatusAdmin(
    uid,
    status
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|
| **status** | **DeviceStatus** | Device\&#39;s status | defaults to undefined|


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
|**200** | Success to update device status. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**402** | Payment required. |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateDeviceStatusOffline**
> updateDeviceStatusOffline()

Update device\'s status to offiline.

### Example

```typescript
import {
    DevicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DevicesApi(configuration);

let uid: string; //Device\'s UID (default to undefined)

const { status, data } = await apiInstance.updateDeviceStatusOffline(
    uid
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|


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
|**200** | Success to update device status to offline. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

