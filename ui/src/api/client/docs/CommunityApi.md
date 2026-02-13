# CommunityApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**acceptDevice**](#acceptdevice) | **PATCH** /api/devices/{uid}/accept | Accept device|
|[**addNamespaceMember**](#addnamespacemember) | **POST** /api/namespaces/{tenant}/members | Invite member|
|[**apiKeyCreate**](#apikeycreate) | **POST** /api/namespaces/api-key | Creates an API key.|
|[**apiKeyDelete**](#apikeydelete) | **DELETE** /api/namespaces/api-key/{key} | Delete an API key|
|[**apiKeyList**](#apikeylist) | **GET** /api/namespaces/api-key | List API Keys|
|[**apiKeyUpdate**](#apikeyupdate) | **PATCH** /api/namespaces/api-key/{key} | Update an API key|
|[**authSSHPublicKey**](#authsshpublickey) | **POST** /api/auth/ssh | Auth SSH public key|
|[**authUser**](#authuser) | **POST** /api/auth/user | Auth a user|
|[**checkSessionRecord**](#checksessionrecord) | **GET** /api/users/security | Check session record status|
|[**createNamespace**](#createnamespace) | **POST** /api/namespaces | Create namespace|
|[**createPublicKey**](#createpublickey) | **POST** /api/sshkeys/public-keys | Create public key|
|[**createTag**](#createtag) | **POST** /api/tags | Create a new tag in the namespace|
|[**createTagDeprecated**](#createtagdeprecated) | **POST** /api/namespaces/{tenant}/tags | Create a new tag in the namespace|
|[**deleteContainer**](#deletecontainer) | **DELETE** /api/containers/{uid} | Delete container|
|[**deleteDevice**](#deletedevice) | **DELETE** /api/devices/{uid} | Delete device|
|[**deleteNamespace**](#deletenamespace) | **DELETE** /api/namespaces/{tenant} | Delete namespace|
|[**deletePublicKey**](#deletepublickey) | **DELETE** /api/sshkeys/public-keys/{fingerprint} | Delete public key|
|[**deleteTag**](#deletetag) | **DELETE** /api/tags/{name} | Delete a tag|
|[**deleteTagDeprecated**](#deletetagdeprecated) | **DELETE** /api/namespaces/{tenant}/tags/{name} | Delete a tag|
|[**editNamespace**](#editnamespace) | **PUT** /api/namespaces/{tenant} | Edit namespace|
|[**getAnnouncement**](#getannouncement) | **GET** /api/announcements/{uuid} | Get a announcement|
|[**getContainer**](#getcontainer) | **GET** /api/containers/{uid} | Get container|
|[**getContainers**](#getcontainers) | **GET** /api/containers | Get containers|
|[**getDevice**](#getdevice) | **GET** /api/devices/{uid} | Get device|
|[**getDevices**](#getdevices) | **GET** /api/devices | Get devices|
|[**getInfo**](#getinfo) | **GET** /info | Get info|
|[**getNamespace**](#getnamespace) | **GET** /api/namespaces/{tenant} | Get a namespace|
|[**getNamespaceToken**](#getnamespacetoken) | **GET** /api/auth/token/{tenant} | Get a new namespace\&#39;s token|
|[**getNamespaces**](#getnamespaces) | **GET** /api/namespaces | Get namespaces list|
|[**getPublicKeys**](#getpublickeys) | **GET** /api/sshkeys/public-keys | Get public keys|
|[**getSession**](#getsession) | **GET** /api/sessions/{uid} | Get session|
|[**getSessions**](#getsessions) | **GET** /api/sessions | Get sessions|
|[**getStatusDevices**](#getstatusdevices) | **GET** /api/stats | Get stats ShellHub instance|
|[**getTags**](#gettags) | **GET** /api/tags | Retrieve all tags associated with the namespace|
|[**getTagsDeprecated**](#gettagsdeprecated) | **GET** /api/namespaces/{tenant}/tags | Retrieve all tags associated with a namespace|
|[**getToken**](#gettoken) | **GET** /api/token/{tenant} | Get token|
|[**getUserInfo**](#getuserinfo) | **GET** /api/auth/user | Get user info|
|[**healthcheck**](#healthcheck) | **GET** /healthcheck | Health check endpoint|
|[**leaveNamespace**](#leavenamespace) | **DELETE** /api/namespaces/{tenant}/members | Leave Namespace|
|[**listAnnouncements**](#listannouncements) | **GET** /api/announcements | List announcements|
|[**login**](#login) | **POST** /api/login | Login|
|[**pullTagFromContainer**](#pulltagfromcontainer) | **DELETE** /api/namespaces/{tenant}/containers/{uid}/tags/{name} | Remove a tag from a container|
|[**pullTagFromDevice**](#pulltagfromdevice) | **DELETE** /api/devices/{uid}/tags/{name} | Remove a tag from a device|
|[**pullTagFromDeviceDeprecated**](#pulltagfromdevicedeprecated) | **DELETE** /api/namespaces/{tenant}/devices/{uid}/tags/{name} | Remove a tag from a device|
|[**pushTagToContainer**](#pushtagtocontainer) | **POST** /api/namespaces/{tenant}/containers/{uid}/tags/{name} | Associate a tag with a container|
|[**pushTagToDevice**](#pushtagtodevice) | **POST** /api/devices/{uid}/tags/{name} | Associate a tag with a device|
|[**pushTagToDeviceDeprecated**](#pushtagtodevicedeprecated) | **POST** /api/namespaces/{tenant}/devices/{uid}/tags/{name} | Associate a tag with a device|
|[**removeNamespaceMember**](#removenamespacemember) | **DELETE** /api/namespaces/{tenant}/members/{uid} | Remove a member from a namespace|
|[**resolveDevice**](#resolvedevice) | **GET** /api/devices/resolve | Resolve Device|
|[**setSessionAuthenticationStatus**](#setsessionauthenticationstatus) | **POST** /api/sessions/{uid} | Set session authentication status|
|[**setSessionRecord**](#setsessionrecord) | **PUT** /api/users/security/{tenant} | Set session record|
|[**setup**](#setup) | **POST** /api/setup | User setup|
|[**updateContainer**](#updatecontainer) | **PUT** /api/containers/{uid} | Update container|
|[**updateContainerStatus**](#updatecontainerstatus) | **PATCH** /api/containers/{uid}/{status} | Update container status|
|[**updateDevice**](#updatedevice) | **PUT** /api/devices/{uid} | Update device|
|[**updateDeviceStatus**](#updatedevicestatus) | **PATCH** /api/devices/{uid}/{status} | Update device status|
|[**updateDeviceStatusOffline**](#updatedevicestatusoffline) | **POST** /internal/devices/{uid}/offline | Update device status to offline|
|[**updateNamespaceMember**](#updatenamespacemember) | **PATCH** /api/namespaces/{tenant}/members/{uid} | Update a member from a namespace|
|[**updatePublicKey**](#updatepublickey) | **PUT** /api/sshkeys/public-keys/{fingerprint} | Update public key|
|[**updateTag**](#updatetag) | **PATCH** /api/tags/{name} | Update a tag|
|[**updateTagDeprecated**](#updatetagdeprecated) | **PATCH** /api/namespaces/{tenant}/tags/{name} | Update a tag|
|[**updateUser**](#updateuser) | **PATCH** /api/users | Update user|
|[**updateUserData**](#updateuserdata) | **PATCH** /api/users/{id}/data | Update user data|
|[**updateUserPassword**](#updateuserpassword) | **PATCH** /api/users/{id}/password | Update user password|

# **acceptDevice**
> acceptDevice()

Change device status to `accepted`.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

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

# **addNamespaceMember**
> Namespace addNamespaceMember()

Invites a member to a namespace.  In enterprise and community instances, the member will automatically accept the invite and will have an `accepted` status.  In cloud instances, the member will have a `pending` status until they accept the invite via an email sent to them. The invite is valid for **7 days**. If the member was previously invited and the invite is no longer valid, the same route will resend the invite. 

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    AddNamespaceMemberRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let addNamespaceMemberRequest: AddNamespaceMemberRequest; // (optional)

const { status, data } = await apiInstance.addNamespaceMember(
    tenant,
    addNamespaceMemberRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **addNamespaceMemberRequest** | **AddNamespaceMemberRequest**|  | |
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|


### Return type

**Namespace**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to add a member to a namespace. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiKeyCreate**
> ApiKeyWithID apiKeyCreate()

The `created_by`, `tenant_id`, and `role` (unless provided in the request body) values will be obtained from the JWT token. 

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    ApiKeyCreate
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let apiKeyCreate: ApiKeyCreate; // (optional)

const { status, data } = await apiInstance.apiKeyCreate(
    apiKeyCreate
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **apiKeyCreate** | **ApiKeyCreate**|  | |


### Return type

**ApiKeyWithID**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiKeyDelete**
> apiKeyDelete()


### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let key: string; // (default to undefined)

const { status, data } = await apiInstance.apiKeyDelete(
    key
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **key** | [**string**] |  | defaults to undefined|


### Return type

void (empty response body)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiKeyList**
> Array<ApiKey> apiKeyList()


### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)
let orderBy: ApiKeyListOrderByParameter; // (optional) (default to undefined)
let sortBy: string; //The property to sort of. (optional) (default to 'expires_in')

const { status, data } = await apiInstance.apiKeyList(
    page,
    perPage,
    orderBy,
    sortBy
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|
| **orderBy** | **ApiKeyListOrderByParameter** |  | (optional) defaults to undefined|
| **sortBy** | [**string**] | The property to sort of. | (optional) defaults to 'expires_in'|


### Return type

**Array<ApiKey>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiKeyUpdate**
> apiKeyUpdate()


### Example

```typescript
import {
    CommunityApi,
    Configuration,
    ApiKeyUpdate
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let key: string; // (default to undefined)
let apiKeyUpdate: ApiKeyUpdate; // (optional)

const { status, data } = await apiInstance.apiKeyUpdate(
    key,
    apiKeyUpdate
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **apiKeyUpdate** | **ApiKeyUpdate**|  | |
| **key** | [**string**] |  | defaults to undefined|


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
|**200** | Success |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **authSSHPublicKey**
> AuthSSHPublicKey200Response authSSHPublicKey()

Authenticate a SSH public key to ShellHub server.

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    AuthSSHPublicKeyRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let authSSHPublicKeyRequest: AuthSSHPublicKeyRequest; // (optional)

const { status, data } = await apiInstance.authSSHPublicKey(
    authSSHPublicKeyRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **authSSHPublicKeyRequest** | **AuthSSHPublicKeyRequest**|  | |


### Return type

**AuthSSHPublicKey200Response**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to auth a SSH public key. |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **authUser**
> UserAuth authUser()

Authenticate a user, returning the session\'s JWT token and data about the user.

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    LoginRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let loginRequest: LoginRequest; // (optional)

const { status, data } = await apiInstance.authUser(
    loginRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **loginRequest** | **LoginRequest**|  | |


### Return type

**UserAuth**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to auth the user. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **checkSessionRecord**
> boolean checkSessionRecord()

Check status from if `session record` feature is enable.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

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

# **createNamespace**
> Namespace createNamespace()

Create a namespace.

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    CreateNamespaceRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let createNamespaceRequest: CreateNamespaceRequest; // (optional)

const { status, data } = await apiInstance.createNamespace(
    createNamespaceRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createNamespaceRequest** | **CreateNamespaceRequest**|  | |


### Return type

**Namespace**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to create a namespace. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createPublicKey**
> CreatePublicKey200Response createPublicKey()

Create a new public key.

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    PublicKeyRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let publicKeyRequest: PublicKeyRequest; // (optional)

const { status, data } = await apiInstance.createPublicKey(
    publicKeyRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **publicKeyRequest** | **PublicKeyRequest**|  | |


### Return type

**CreatePublicKey200Response**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to create a public key. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**409** | Conflict |  -  |
|**422** | UnprocessableEntity |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createTag**
> createTag(createTagRequest)

Creates a tag in the authenticated namespace that can be later associated with content. Tag names must be unique within the namespace.

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    CreateTagRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let createTagRequest: CreateTagRequest; //

const { status, data } = await apiInstance.createTag(
    createTagRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createTagRequest** | **CreateTagRequest**|  | |


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
|**201** | Tag successfully created |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createTagDeprecated**
> createTagDeprecated(createTagRequest)

**Deprecated**: Use `POST /api/tags` instead. Creates a tag that can be later associated with content. Tag names must be unique within the namespace.

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    CreateTagRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let createTagRequest: CreateTagRequest; //

const { status, data } = await apiInstance.createTagDeprecated(
    tenant,
    createTagRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createTagRequest** | **CreateTagRequest**|  | |
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|


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
|**201** | Tag successfully created |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteContainer**
> deleteContainer()

Delete a container.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let uid: string; //Device\'s UID (default to undefined)

const { status, data } = await apiInstance.deleteContainer(
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
|**200** | Success to delete a container. |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteDevice**
> deleteDevice()

Delete a device.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

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

# **deleteNamespace**
> deleteNamespace()

Delete a namespace.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)

const { status, data } = await apiInstance.deleteNamespace(
    tenant
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|


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
|**200** | Success to delete a namespace. |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deletePublicKey**
> deletePublicKey()

Delete a public key.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let fingerprint: string; //Public key\'s fingerprint. (default to undefined)

const { status, data } = await apiInstance.deletePublicKey(
    fingerprint
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **fingerprint** | [**string**] | Public key\&#39;s fingerprint. | defaults to undefined|


### Return type

void (empty response body)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to delete a public key. |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteTag**
> deleteTag()

Removes a tag and all its associations from the authenticated namespace.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let name: string; //Tag name to delete (default to undefined)

const { status, data } = await apiInstance.deleteTag(
    name
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **name** | [**string**] | Tag name to delete | defaults to undefined|


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
|**204** | Tag successfully deleted |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteTagDeprecated**
> deleteTagDeprecated()

**Deprecated**: Use `DELETE /api/tags/{name}` instead. Removes a tag and all its associations.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let name: string; //Tag name to delete (default to undefined)

const { status, data } = await apiInstance.deleteTagDeprecated(
    tenant,
    name
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|
| **name** | [**string**] | Tag name to delete | defaults to undefined|


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
|**204** | Tag successfully deleted |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **editNamespace**
> Namespace editNamespace()

Edit a namespace.

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    EditNamespaceRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let editNamespaceRequest: EditNamespaceRequest; // (optional)

const { status, data } = await apiInstance.editNamespace(
    tenant,
    editNamespaceRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **editNamespaceRequest** | **EditNamespaceRequest**|  | |
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|


### Return type

**Namespace**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to edit a namespace. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAnnouncement**
> Announcement getAnnouncement()

Get a announcement.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let uuid: string; // (default to undefined)

const { status, data } = await apiInstance.getAnnouncement(
    uuid
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uuid** | [**string**] |  | defaults to undefined|


### Return type

**Announcement**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a announcement. |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getContainer**
> Device getContainer()

Get a container.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let uid: string; //Device\'s UID (default to undefined)

const { status, data } = await apiInstance.getContainer(
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
|**200** | Success get a container. |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getContainers**
> Array<Device> getContainers()

Get a list of containers.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let filter: string; //Filter field receives a JSON object enconded as base64 string for limit a search.  The JSON enconded must follow these interafaces: ```typescript interface ParamProperty {   name: string;   operator: \"contains\" | \"eq\" | \"bool\" | \"gt\" | \"lt\";   value: string; }  interface ParamOperator {   name: \"and\" | \"or\"; }  interface Filter {   type: \"property\" | \"operator\";   param: ParamOperator | ParamProperty; }  interface FilterList {   Filters: Array<Filter>; }  ```  ## Examples  This is a example to filter and get only the resource what property \"confirmed\" is \"true\" ```json [   {   \"type\": \"property\",   \"params\": {       \"name\": \"confirmed\",       \"operator\": \"bool\",       \"value\": \"true\"       }   } ] ```  This one, filter resource by the property \"id\" inside \"info\" structure when it is equal to \"manjaro\" and online property is set to \"true\" ```json [   {     \"type\": \"property\",     \"params\": {       \"name\": \"info.id\",       \"operator\": \"eq\",       \"value\": \"manjaro\"     }   },   {     \"type\": \"property\",     \"params\": {       \"name\": \"online\",       \"operator\": \"bool\",       \"value\": \"true\"     }   },   {     \"type\": \"operator\",     \"params\": {       \"name\": \"and\"     }   } ] ```  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)
let status: DeviceStatus; //Container\'s status (optional) (default to undefined)
let sortBy: string; //Container\'s property to sort of (optional) (default to 'last_seen')
let orderBy: GetContainersOrderByParameter; // (optional) (default to undefined)

const { status, data } = await apiInstance.getContainers(
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
| **status** | **DeviceStatus** | Container\&#39;s status | (optional) defaults to undefined|
| **sortBy** | [**string**] | Container\&#39;s property to sort of | (optional) defaults to 'last_seen'|
| **orderBy** | **GetContainersOrderByParameter** |  | (optional) defaults to undefined|


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
|**200** | Success to get a list of containers. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getDevice**
> Device getDevice()

Get a device.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

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

# **getDevices**
> Array<Device> getDevices()

Get a list of devices.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

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

# **getInfo**
> Info getInfo()

Get information about ShellHub instance like version, SSH and API addresses.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let agentVersion: string; //Agent\'s version. This parameter is used to filter instance information based on the requesting agent\'s version. (optional) (default to undefined)

const { status, data } = await apiInstance.getInfo(
    agentVersion
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **agentVersion** | [**string**] | Agent\&#39;s version. This parameter is used to filter instance information based on the requesting agent\&#39;s version. | (optional) defaults to undefined|


### Return type

**Info**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get ShellHub instance info. |  -  |
|**400** | Bad request |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getNamespace**
> Namespace getNamespace()

Get a namespace.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)

const { status, data } = await apiInstance.getNamespace(
    tenant
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|


### Return type

**Namespace**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a namespace. |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getNamespaceToken**
> UserAuth getNamespaceToken()

This route works like a login\'s one; returns a JWT token and extra information about namespace.  You can use this route to swap between namespaces. 

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)

const { status, data } = await apiInstance.getNamespaceToken(
    tenant
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|


### Return type

**UserAuth**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get namespace\&#39;s token |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getNamespaces**
> Array<Namespace> getNamespaces()

Returns a list of namespaces.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let filter: string; //Namespaces\'s filter.   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called `type`, it will filter by a `property` called `name` where the value should `contains` `examplespace`.  If you want get only Namespaces name as `examplespace`, the JSON object will looks like this   ```json [   {     \"type\":\"property\",     \"params\":{       \"name\":\"name\",       \"operator\":\"contains\",       \"value\":\"examplespace\"     }   } ] ```  So, the output encoded string will result on: `W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lIiwib3BlcmF0b3IiOiJjb250YWlucyIsInZhbHVlIjoiZXhhbXBsZXNwYWNlIn19XQ==`  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.getNamespaces(
    filter,
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **filter** | [**string**] | Namespaces\&#39;s filter.   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called &#x60;type&#x60;, it will filter by a &#x60;property&#x60; called &#x60;name&#x60; where the value should &#x60;contains&#x60; &#x60;examplespace&#x60;.  If you want get only Namespaces name as &#x60;examplespace&#x60;, the JSON object will looks like this   &#x60;&#x60;&#x60;json [   {     \&quot;type\&quot;:\&quot;property\&quot;,     \&quot;params\&quot;:{       \&quot;name\&quot;:\&quot;name\&quot;,       \&quot;operator\&quot;:\&quot;contains\&quot;,       \&quot;value\&quot;:\&quot;examplespace\&quot;     }   } ] &#x60;&#x60;&#x60;  So, the output encoded string will result on: &#x60;W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lIiwib3BlcmF0b3IiOiJjb250YWlucyIsInZhbHVlIjoiZXhhbXBsZXNwYWNlIn19XQ&#x3D;&#x3D;&#x60;  | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|


### Return type

**Array<Namespace>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a namespace list. |  * X-Total-Count - Namespaces\&#39; total number. <br>  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getPublicKeys**
> Array<PublicKeyResponse> getPublicKeys()

Get a list from all public keys.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let filter: string; //Filter field receives a JSON object enconded as base64 string for limit a search.  The JSON enconded must follow these interafaces: ```typescript interface ParamProperty {   name: string;   operator: \"contains\" | \"eq\" | \"bool\" | \"gt\" | \"lt\";   value: string; }  interface ParamOperator {   name: \"and\" | \"or\"; }  interface Filter {   type: \"property\" | \"operator\";   param: ParamOperator | ParamProperty; }  interface FilterList {   Filters: Array<Filter>; }  ```  ## Examples  This is a example to filter and get only the resource what property \"confirmed\" is \"true\" ```json [   {   \"type\": \"property\",   \"params\": {       \"name\": \"confirmed\",       \"operator\": \"bool\",       \"value\": \"true\"       }   } ] ```  This one, filter resource by the property \"id\" inside \"info\" structure when it is equal to \"manjaro\" and online property is set to \"true\" ```json [   {     \"type\": \"property\",     \"params\": {       \"name\": \"info.id\",       \"operator\": \"eq\",       \"value\": \"manjaro\"     }   },   {     \"type\": \"property\",     \"params\": {       \"name\": \"online\",       \"operator\": \"bool\",       \"value\": \"true\"     }   },   {     \"type\": \"operator\",     \"params\": {       \"name\": \"and\"     }   } ] ```  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.getPublicKeys(
    filter,
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **filter** | [**string**] | Filter field receives a JSON object enconded as base64 string for limit a search.  The JSON enconded must follow these interafaces: &#x60;&#x60;&#x60;typescript interface ParamProperty {   name: string;   operator: \&quot;contains\&quot; | \&quot;eq\&quot; | \&quot;bool\&quot; | \&quot;gt\&quot; | \&quot;lt\&quot;;   value: string; }  interface ParamOperator {   name: \&quot;and\&quot; | \&quot;or\&quot;; }  interface Filter {   type: \&quot;property\&quot; | \&quot;operator\&quot;;   param: ParamOperator | ParamProperty; }  interface FilterList {   Filters: Array&lt;Filter&gt;; }  &#x60;&#x60;&#x60;  ## Examples  This is a example to filter and get only the resource what property \&quot;confirmed\&quot; is \&quot;true\&quot; &#x60;&#x60;&#x60;json [   {   \&quot;type\&quot;: \&quot;property\&quot;,   \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;confirmed\&quot;,       \&quot;operator\&quot;: \&quot;bool\&quot;,       \&quot;value\&quot;: \&quot;true\&quot;       }   } ] &#x60;&#x60;&#x60;  This one, filter resource by the property \&quot;id\&quot; inside \&quot;info\&quot; structure when it is equal to \&quot;manjaro\&quot; and online property is set to \&quot;true\&quot; &#x60;&#x60;&#x60;json [   {     \&quot;type\&quot;: \&quot;property\&quot;,     \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;info.id\&quot;,       \&quot;operator\&quot;: \&quot;eq\&quot;,       \&quot;value\&quot;: \&quot;manjaro\&quot;     }   },   {     \&quot;type\&quot;: \&quot;property\&quot;,     \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;online\&quot;,       \&quot;operator\&quot;: \&quot;bool\&quot;,       \&quot;value\&quot;: \&quot;true\&quot;     }   },   {     \&quot;type\&quot;: \&quot;operator\&quot;,     \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;and\&quot;     }   } ] &#x60;&#x60;&#x60;  | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|


### Return type

**Array<PublicKeyResponse>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a list of public keys. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getSession**
> Session getSession()

Get a session.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

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

# **getSessions**
> Array<Session> getSessions()

Get a list sessions.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

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

# **getStatusDevices**
> GetStatusDevices200Response getStatusDevices()

Get stats ShellHub instance.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

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

# **getTags**
> Array<Tag> getTags()

Retrieves all tags for the authenticated namespace.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let filter: string; //Filter field receives a JSON object enconded as base64 string for limit a search.  The JSON enconded must follow these interafaces: ```typescript interface ParamProperty {   name: string;   operator: \"contains\" | \"eq\" | \"bool\" | \"gt\" | \"lt\";   value: string; }  interface ParamOperator {   name: \"and\" | \"or\"; }  interface Filter {   type: \"property\" | \"operator\";   param: ParamOperator | ParamProperty; }  interface FilterList {   Filters: Array<Filter>; }  ```  ## Examples  This is a example to filter and get only the resource what property \"confirmed\" is \"true\" ```json [   {   \"type\": \"property\",   \"params\": {       \"name\": \"confirmed\",       \"operator\": \"bool\",       \"value\": \"true\"       }   } ] ```  This one, filter resource by the property \"id\" inside \"info\" structure when it is equal to \"manjaro\" and online property is set to \"true\" ```json [   {     \"type\": \"property\",     \"params\": {       \"name\": \"info.id\",       \"operator\": \"eq\",       \"value\": \"manjaro\"     }   },   {     \"type\": \"property\",     \"params\": {       \"name\": \"online\",       \"operator\": \"bool\",       \"value\": \"true\"     }   },   {     \"type\": \"operator\",     \"params\": {       \"name\": \"and\"     }   } ] ```  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.getTags(
    filter,
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **filter** | [**string**] | Filter field receives a JSON object enconded as base64 string for limit a search.  The JSON enconded must follow these interafaces: &#x60;&#x60;&#x60;typescript interface ParamProperty {   name: string;   operator: \&quot;contains\&quot; | \&quot;eq\&quot; | \&quot;bool\&quot; | \&quot;gt\&quot; | \&quot;lt\&quot;;   value: string; }  interface ParamOperator {   name: \&quot;and\&quot; | \&quot;or\&quot;; }  interface Filter {   type: \&quot;property\&quot; | \&quot;operator\&quot;;   param: ParamOperator | ParamProperty; }  interface FilterList {   Filters: Array&lt;Filter&gt;; }  &#x60;&#x60;&#x60;  ## Examples  This is a example to filter and get only the resource what property \&quot;confirmed\&quot; is \&quot;true\&quot; &#x60;&#x60;&#x60;json [   {   \&quot;type\&quot;: \&quot;property\&quot;,   \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;confirmed\&quot;,       \&quot;operator\&quot;: \&quot;bool\&quot;,       \&quot;value\&quot;: \&quot;true\&quot;       }   } ] &#x60;&#x60;&#x60;  This one, filter resource by the property \&quot;id\&quot; inside \&quot;info\&quot; structure when it is equal to \&quot;manjaro\&quot; and online property is set to \&quot;true\&quot; &#x60;&#x60;&#x60;json [   {     \&quot;type\&quot;: \&quot;property\&quot;,     \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;info.id\&quot;,       \&quot;operator\&quot;: \&quot;eq\&quot;,       \&quot;value\&quot;: \&quot;manjaro\&quot;     }   },   {     \&quot;type\&quot;: \&quot;property\&quot;,     \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;online\&quot;,       \&quot;operator\&quot;: \&quot;bool\&quot;,       \&quot;value\&quot;: \&quot;true\&quot;     }   },   {     \&quot;type\&quot;: \&quot;operator\&quot;,     \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;and\&quot;     }   } ] &#x60;&#x60;&#x60;  | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|


### Return type

**Array<Tag>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get tag list. |  * X-Total-Count - Namespaces\&#39; total number. <br>  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getTagsDeprecated**
> Array<Tag> getTagsDeprecated()

**Deprecated**: Use `GET /api/tags` instead.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let filter: string; //Filter field receives a JSON object enconded as base64 string for limit a search.  The JSON enconded must follow these interafaces: ```typescript interface ParamProperty {   name: string;   operator: \"contains\" | \"eq\" | \"bool\" | \"gt\" | \"lt\";   value: string; }  interface ParamOperator {   name: \"and\" | \"or\"; }  interface Filter {   type: \"property\" | \"operator\";   param: ParamOperator | ParamProperty; }  interface FilterList {   Filters: Array<Filter>; }  ```  ## Examples  This is a example to filter and get only the resource what property \"confirmed\" is \"true\" ```json [   {   \"type\": \"property\",   \"params\": {       \"name\": \"confirmed\",       \"operator\": \"bool\",       \"value\": \"true\"       }   } ] ```  This one, filter resource by the property \"id\" inside \"info\" structure when it is equal to \"manjaro\" and online property is set to \"true\" ```json [   {     \"type\": \"property\",     \"params\": {       \"name\": \"info.id\",       \"operator\": \"eq\",       \"value\": \"manjaro\"     }   },   {     \"type\": \"property\",     \"params\": {       \"name\": \"online\",       \"operator\": \"bool\",       \"value\": \"true\"     }   },   {     \"type\": \"operator\",     \"params\": {       \"name\": \"and\"     }   } ] ```  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.getTagsDeprecated(
    tenant,
    filter,
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|
| **filter** | [**string**] | Filter field receives a JSON object enconded as base64 string for limit a search.  The JSON enconded must follow these interafaces: &#x60;&#x60;&#x60;typescript interface ParamProperty {   name: string;   operator: \&quot;contains\&quot; | \&quot;eq\&quot; | \&quot;bool\&quot; | \&quot;gt\&quot; | \&quot;lt\&quot;;   value: string; }  interface ParamOperator {   name: \&quot;and\&quot; | \&quot;or\&quot;; }  interface Filter {   type: \&quot;property\&quot; | \&quot;operator\&quot;;   param: ParamOperator | ParamProperty; }  interface FilterList {   Filters: Array&lt;Filter&gt;; }  &#x60;&#x60;&#x60;  ## Examples  This is a example to filter and get only the resource what property \&quot;confirmed\&quot; is \&quot;true\&quot; &#x60;&#x60;&#x60;json [   {   \&quot;type\&quot;: \&quot;property\&quot;,   \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;confirmed\&quot;,       \&quot;operator\&quot;: \&quot;bool\&quot;,       \&quot;value\&quot;: \&quot;true\&quot;       }   } ] &#x60;&#x60;&#x60;  This one, filter resource by the property \&quot;id\&quot; inside \&quot;info\&quot; structure when it is equal to \&quot;manjaro\&quot; and online property is set to \&quot;true\&quot; &#x60;&#x60;&#x60;json [   {     \&quot;type\&quot;: \&quot;property\&quot;,     \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;info.id\&quot;,       \&quot;operator\&quot;: \&quot;eq\&quot;,       \&quot;value\&quot;: \&quot;manjaro\&quot;     }   },   {     \&quot;type\&quot;: \&quot;property\&quot;,     \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;online\&quot;,       \&quot;operator\&quot;: \&quot;bool\&quot;,       \&quot;value\&quot;: \&quot;true\&quot;     }   },   {     \&quot;type\&quot;: \&quot;operator\&quot;,     \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;and\&quot;     }   } ] &#x60;&#x60;&#x60;  | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|


### Return type

**Array<Tag>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get tag list. |  * X-Total-Count - Namespaces\&#39; total number. <br>  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getToken**
> UserAuth getToken()

Get a token from its tenant.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let tenant: string; //Tenant (default to undefined)

const { status, data } = await apiInstance.getToken(
    tenant
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tenant** | [**string**] | Tenant | defaults to undefined|


### Return type

**UserAuth**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success get token |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getUserInfo**
> UserAuth getUserInfo()


### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

const { status, data } = await apiInstance.getUserInfo();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**UserAuth**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get the user info. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **healthcheck**
> healthcheck()

Lightweight endpoint to verify API availability. Returns HTTP 200 with empty body if the service is healthy. Used by load balancers, monitoring systems, and the UI for service availability detection. 

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

const { status, data } = await apiInstance.healthcheck();
```

### Parameters
This endpoint does not have any parameters.


### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Service is healthy and responding |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **leaveNamespace**
> UserAuth leaveNamespace()

Allows the authenticated user to leave the specified namespace. Owners cannot leave a namespace; they must delete it instead. If the user attempts to leave their current authenticated namespace, the response will provide a new token that excludes this namespace. 

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)

const { status, data } = await apiInstance.leaveNamespace(
    tenant
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|


### Return type

**UserAuth**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Successfully left the namespace. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listAnnouncements**
> Array<AnnouncementShort> listAnnouncements()

List the announcements posted by ShellHub Cloud.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)
let orderBy: ListAnnouncementsOrderByParameter; // (optional) (default to undefined)

const { status, data } = await apiInstance.listAnnouncements(
    page,
    perPage,
    orderBy
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|
| **orderBy** | **ListAnnouncementsOrderByParameter** |  | (optional) defaults to undefined|


### Return type

**Array<AnnouncementShort>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get the announcements. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**400** | Bad request |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **login**
> UserAuth login()

Authenticate a \"local\" user by returning the session\'s JWT token and user data. Local users are those registered via the ShellHub form without relying on external Identity Providers (IdPs).  Authentication may result in an account lockout after N consecutive incorrect login attempts. The lockout applies specifically to a particular source and user combination. Check for the presence of the `X-Account-Lockout` header to determine the account lockout status. When it\'s 0, there are no active lockouts.  Users with MFA enabled cannot authenticate via this route. In such cases, the API will respond with a status `401` and an `X-MFA-Token` header with a UUID. Authentication must be med to `/api/mfa/auth` with this token in these instances. 

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    LoginRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let loginRequest: LoginRequest; // (optional)

const { status, data } = await apiInstance.login(
    loginRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **loginRequest** | **LoginRequest**|  | |


### Return type

**UserAuth**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to login. |  * X-Account-Lockout -  <br>  * X-MFA-Token -  <br>  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  * X-MFA-Token -  <br>  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**429** | Too Many Requests |  * X-Account-Lockout -  <br>  * X-MFA-Token -  <br>  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **pullTagFromContainer**
> pullTagFromContainer()


### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let uid: string; //Device\'s UID (default to undefined)
let name: string; //Tag name to remove (default to undefined)

const { status, data } = await apiInstance.pullTagFromContainer(
    tenant,
    uid,
    name
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|
| **name** | [**string**] | Tag name to remove | defaults to undefined|


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
|**204** | Tag successfully removed from container |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **pullTagFromDevice**
> pullTagFromDevice()

Removes a tag from a device in the authenticated namespace.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let uid: string; //Device\'s UID (default to undefined)
let name: string; //Tag name to remove (default to undefined)

const { status, data } = await apiInstance.pullTagFromDevice(
    uid,
    name
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|
| **name** | [**string**] | Tag name to remove | defaults to undefined|


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
|**204** | Tag successfully removed from device |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **pullTagFromDeviceDeprecated**
> pullTagFromDeviceDeprecated()

**Deprecated**: Use `DELETE /api/devices/{uid}/tags/{name}` instead.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let uid: string; //Device\'s UID (default to undefined)
let name: string; //Tag name to remove (default to undefined)

const { status, data } = await apiInstance.pullTagFromDeviceDeprecated(
    tenant,
    uid,
    name
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|
| **name** | [**string**] | Tag name to remove | defaults to undefined|


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
|**204** | Tag successfully removed from device |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **pushTagToContainer**
> pushTagToContainer()


### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let uid: string; //Device\'s UID (default to undefined)
let name: string; //Tag name to associate (default to undefined)

const { status, data } = await apiInstance.pushTagToContainer(
    tenant,
    uid,
    name
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|
| **name** | [**string**] | Tag name to associate | defaults to undefined|


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
|**204** | Tag successfully associated with container |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **pushTagToDevice**
> pushTagToDevice()

Associates a tag with a device in the authenticated namespace.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let uid: string; //Device\'s UID (default to undefined)
let name: string; //Tag name to associate (default to undefined)

const { status, data } = await apiInstance.pushTagToDevice(
    uid,
    name
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|
| **name** | [**string**] | Tag name to associate | defaults to undefined|


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
|**204** | Tag successfully associated with device |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **pushTagToDeviceDeprecated**
> pushTagToDeviceDeprecated()

**Deprecated**: Use `POST /api/devices/{uid}/tags/{name}` instead.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let uid: string; //Device\'s UID (default to undefined)
let name: string; //Tag name to associate (default to undefined)

const { status, data } = await apiInstance.pushTagToDeviceDeprecated(
    tenant,
    uid,
    name
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|
| **name** | [**string**] | Tag name to associate | defaults to undefined|


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
|**204** | Tag successfully associated with device |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **removeNamespaceMember**
> Namespace removeNamespaceMember()

Remove a member from a namespace.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let uid: string; //Member\'s ID (default to undefined)

const { status, data } = await apiInstance.removeNamespaceMember(
    tenant,
    uid
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|
| **uid** | [**string**] | Member\&#39;s ID | defaults to undefined|


### Return type

**Namespace**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to delete a member from namespace. |  -  |
|**401** | Unauthorized |  -  |
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
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

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

# **setSessionAuthenticationStatus**
> setSessionAuthenticationStatus()

Set session authentication status.

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    SetSessionAuthenticationStatusRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

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

# **setSessionRecord**
> setSessionRecord()

Define if sessions will be recorded.

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    SetSessionRecordRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

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

# **setup**
> setup()

Register an user and create namespace with the same name as username

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    SetupRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let sign: string; //Signature used to validate request origin generated by running `./bin/setup` script (default to undefined)
let setupRequest: SetupRequest; // (optional)

const { status, data } = await apiInstance.setup(
    sign,
    setupRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **setupRequest** | **SetupRequest**|  | |
| **sign** | [**string**] | Signature used to validate request origin generated by running &#x60;./bin/setup&#x60; script | defaults to undefined|


### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to register user on setup. |  -  |
|**400** | Bad request |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateContainer**
> updateContainer()

Update container\'s data.

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    UpdateDeviceRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let uid: string; //Device\'s UID (default to undefined)
let updateDeviceRequest: UpdateDeviceRequest; // (optional)

const { status, data } = await apiInstance.updateContainer(
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
|**200** | Success to update container\&#39;s data. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateContainerStatus**
> updateContainerStatus()

Update container\'s status.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let uid: string; //Device\'s UID (default to undefined)
let status: UpdateDeviceStatusStatusParameter; //Container\'s status (default to undefined)

const { status, data } = await apiInstance.updateContainerStatus(
    uid,
    status
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|
| **status** | **UpdateDeviceStatusStatusParameter** | Container\&#39;s status | defaults to undefined|


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
|**200** | Success to update container status. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**402** | Payment required |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateDevice**
> updateDevice()

Update device\'s data.

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    UpdateDeviceRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

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

# **updateDeviceStatus**
> updateDeviceStatus()

Update device\'s status.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

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

# **updateDeviceStatusOffline**
> updateDeviceStatusOffline()

Update device\'s status to offiline.

### Example

```typescript
import {
    CommunityApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

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

# **updateNamespaceMember**
> updateNamespaceMember()

Update a member role from a namespace.

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    UpdateNamespaceMemberRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let uid: string; //Member\'s ID (default to undefined)
let updateNamespaceMemberRequest: UpdateNamespaceMemberRequest; // (optional)

const { status, data } = await apiInstance.updateNamespaceMember(
    tenant,
    uid,
    updateNamespaceMemberRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateNamespaceMemberRequest** | **UpdateNamespaceMemberRequest**|  | |
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|
| **uid** | [**string**] | Member\&#39;s ID | defaults to undefined|


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
|**200** | Success to update member role from a namespace. |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updatePublicKey**
> PublicKeyResponse updatePublicKey()

Update a public key.

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    UpdatePublicKeyRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let fingerprint: string; //Public key\'s fingerprint. (default to undefined)
let updatePublicKeyRequest: UpdatePublicKeyRequest; // (optional)

const { status, data } = await apiInstance.updatePublicKey(
    fingerprint,
    updatePublicKeyRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updatePublicKeyRequest** | **UpdatePublicKeyRequest**|  | |
| **fingerprint** | [**string**] | Public key\&#39;s fingerprint. | defaults to undefined|


### Return type

**PublicKeyResponse**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to update a public key. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateTag**
> Tag updateTag(updateTagRequest)

Updates a tag in the authenticated namespace.

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    UpdateTagRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let name: string; //Current tag name (default to undefined)
let updateTagRequest: UpdateTagRequest; //

const { status, data } = await apiInstance.updateTag(
    name,
    updateTagRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateTagRequest** | **UpdateTagRequest**|  | |
| **name** | [**string**] | Current tag name | defaults to undefined|


### Return type

**Tag**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Tag successfully updated |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateTagDeprecated**
> Tag updateTagDeprecated(updateTagRequest)

**Deprecated**: Use `PATCH /api/tags/{name}` instead.

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    UpdateTagRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let name: string; //Current tag name (default to undefined)
let updateTagRequest: UpdateTagRequest; //

const { status, data } = await apiInstance.updateTagDeprecated(
    tenant,
    name,
    updateTagRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateTagRequest** | **UpdateTagRequest**|  | |
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|
| **name** | [**string**] | Current tag name | defaults to undefined|


### Return type

**Tag**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Tag successfully updated |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateUser**
> updateUser()


### Example

```typescript
import {
    CommunityApi,
    Configuration,
    UpdateUserRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let updateUserRequest: UpdateUserRequest; // (optional)

const { status, data } = await apiInstance.updateUser(
    updateUserRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateUserRequest** | **UpdateUserRequest**|  | |


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
|**200** | Success |  -  |
|**400** | Invalid Fields |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**409** | Conflict Fields |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateUserData**
> updateUserData()

Update user\'s data.

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    UpdateUserDataRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let id: string; //User\'s ID. (default to undefined)
let updateUserDataRequest: UpdateUserDataRequest; // (optional)

const { status, data } = await apiInstance.updateUserData(
    id,
    updateUserDataRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateUserDataRequest** | **UpdateUserDataRequest**|  | |
| **id** | [**string**] | User\&#39;s ID. | defaults to undefined|


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
|**200** | Success |  -  |
|**400** | Invalid Fields |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**409** | Conflict Fields |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateUserPassword**
> updateUserPassword()

Update only the user password.

### Example

```typescript
import {
    CommunityApi,
    Configuration,
    UpdateUserPasswordRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CommunityApi(configuration);

let id: string; //User ID (default to undefined)
let updateUserPasswordRequest: UpdateUserPasswordRequest; // (optional)

const { status, data } = await apiInstance.updateUserPassword(
    id,
    updateUserPasswordRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateUserPasswordRequest** | **UpdateUserPasswordRequest**|  | |
| **id** | [**string**] | User ID | defaults to undefined|


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
|**200** | Success |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

