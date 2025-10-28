# SessionsApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**checkSessionRecord**](#checksessionrecord) | **GET** /api/users/security | Check session record status|
|[**clsoeSession**](#clsoesession) | **POST** /api/sessions/{uid}/close | Close session|
|[**deleteSessionRecord**](#deletesessionrecord) | **DELETE** /api/sessions/{uid}/records/{seat} | Delete session record|
|[**getSession**](#getsession) | **GET** /api/sessions/{uid} | Get session|
|[**getSessionAdmin**](#getsessionadmin) | **GET** /admin/api/sessions/{uid} | Get session admin|
|[**getSessionRecord**](#getsessionrecord) | **GET** /api/sessions/{uid}/records/{seat} | Get session record|
|[**getSessions**](#getsessions) | **GET** /api/sessions | Get sessions|
|[**getSessionsAdmin**](#getsessionsadmin) | **GET** /admin/api/sessions | Get sessions admin|
|[**recordSession**](#recordsession) | **POST** /api/sessions/{uid}/records/{seat} | Record session|
|[**setSessionAuthenticationStatus**](#setsessionauthenticationstatus) | **POST** /api/sessions/{uid} | Set session authentication status|
|[**setSessionAuthenticationStatusAdmin**](#setsessionauthenticationstatusadmin) | **POST** /admin/api/sessions/{uid} | Set session authentication status admin|
|[**setSessionRecord**](#setsessionrecord) | **PUT** /api/users/security/{tenant} | Set session record|

# **checkSessionRecord**
> boolean checkSessionRecord()

Check status from if `session record` feature is enable.

### Example

```typescript
import {
    SessionsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new SessionsApi(configuration);

const { status, data } = await apiInstance.checkSessionRecord();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**boolean**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success get session record status |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **clsoeSession**
> clsoeSession()

Close a session.

### Example

```typescript
import {
    SessionsApi,
    Configuration,
    ClsoeSessionRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new SessionsApi(configuration);

let uid: string; // (default to undefined)
let clsoeSessionRequest: ClsoeSessionRequest; // (optional)

const { status, data } = await apiInstance.clsoeSession(
    uid,
    clsoeSessionRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **clsoeSessionRequest** | **ClsoeSessionRequest**|  | |
| **uid** | [**string**] |  | defaults to undefined|


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
|**200** | Success to close session. |  -  |
|**400** | Bad request |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteSessionRecord**
> deleteSessionRecord()

Deletes a session record based on its seat.

### Example

```typescript
import {
    SessionsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new SessionsApi(configuration);

let uid: string; // (default to undefined)
let seat: number; // (default to undefined)

const { status, data } = await apiInstance.deleteSessionRecord(
    uid,
    seat
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] |  | defaults to undefined|
| **seat** | [**number**] |  | defaults to undefined|


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
|**200** | Success to delete a session\&#39;s record. |  -  |
|**400** | Bad request |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getSession**
> Session getSession()

Get a session.

### Example

```typescript
import {
    SessionsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new SessionsApi(configuration);

let uid: string; // (default to undefined)

const { status, data } = await apiInstance.getSession(
    uid
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] |  | defaults to undefined|


### Return type

**Session**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a session. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getSessionAdmin**
> Session getSessionAdmin()

Get a session.

### Example

```typescript
import {
    SessionsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new SessionsApi(configuration);

let uid: string; // (default to undefined)

const { status, data } = await apiInstance.getSessionAdmin(
    uid
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] |  | defaults to undefined|


### Return type

**Session**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a session. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getSessionRecord**
> Array<RecordedSessionResponseInner> getSessionRecord()

Get a session record based on its seat.

### Example

```typescript
import {
    SessionsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new SessionsApi(configuration);

let uid: string; // (default to undefined)
let seat: number; // (default to undefined)

const { status, data } = await apiInstance.getSessionRecord(
    uid,
    seat
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] |  | defaults to undefined|
| **seat** | [**number**] |  | defaults to undefined|


### Return type

**Array<RecordedSessionResponseInner>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to play a session\&#39;s record. |  -  |
|**400** | Bad request |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getSessions**
> Array<Session> getSessions()

Get a list sessions.

### Example

```typescript
import {
    SessionsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new SessionsApi(configuration);

let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.getSessions(
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|


### Return type

**Array<Session>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get list of sessions. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getSessionsAdmin**
> Array<Session> getSessionsAdmin()

Get a list sessions.

### Example

```typescript
import {
    SessionsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new SessionsApi(configuration);

let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.getSessionsAdmin(
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|


### Return type

**Array<Session>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get list of sessions. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **recordSession**
> recordSession()

Record data about session session.

### Example

```typescript
import {
    SessionsApi,
    Configuration,
    RecordSessionRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new SessionsApi(configuration);

let uid: string; // (default to undefined)
let seat: number; // (default to undefined)
let recordSessionRequest: RecordSessionRequest; // (optional)

const { status, data } = await apiInstance.recordSession(
    uid,
    seat,
    recordSessionRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **recordSessionRequest** | **RecordSessionRequest**|  | |
| **uid** | [**string**] |  | defaults to undefined|
| **seat** | [**number**] |  | defaults to undefined|


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
|**200** | Success to record session. |  -  |
|**400** | Bad request |  -  |
|**422** | Unprocessable Entity |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **setSessionAuthenticationStatus**
> setSessionAuthenticationStatus()

Set session authentication status.

### Example

```typescript
import {
    SessionsApi,
    Configuration,
    SetSessionAuthenticationStatusRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new SessionsApi(configuration);

let uid: string; // (default to undefined)
let setSessionAuthenticationStatusRequest: SetSessionAuthenticationStatusRequest; // (optional)

const { status, data } = await apiInstance.setSessionAuthenticationStatus(
    uid,
    setSessionAuthenticationStatusRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **setSessionAuthenticationStatusRequest** | **SetSessionAuthenticationStatusRequest**|  | |
| **uid** | [**string**] |  | defaults to undefined|


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
|**200** | Success |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **setSessionAuthenticationStatusAdmin**
> setSessionAuthenticationStatusAdmin()

Set session authentication status.

### Example

```typescript
import {
    SessionsApi,
    Configuration,
    SetSessionAuthenticationStatusRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new SessionsApi(configuration);

let uid: string; // (default to undefined)
let setSessionAuthenticationStatusRequest: SetSessionAuthenticationStatusRequest; // (optional)

const { status, data } = await apiInstance.setSessionAuthenticationStatusAdmin(
    uid,
    setSessionAuthenticationStatusRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **setSessionAuthenticationStatusRequest** | **SetSessionAuthenticationStatusRequest**|  | |
| **uid** | [**string**] |  | defaults to undefined|


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
|**200** | Success |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **setSessionRecord**
> setSessionRecord()

Define if sessions will be recorded.

### Example

```typescript
import {
    SessionsApi,
    Configuration,
    SetSessionRecordRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new SessionsApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let setSessionRecordRequest: SetSessionRecordRequest; // (optional)

const { status, data } = await apiInstance.setSessionRecord(
    tenant,
    setSessionRecordRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **setSessionRecordRequest** | **SetSessionRecordRequest**|  | |
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|


### Return type

void (empty response body)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to set session record status. |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

