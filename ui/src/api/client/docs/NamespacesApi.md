# NamespacesApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**acceptInvite**](#acceptinvite) | **PATCH** /api/namespaces/{tenant}/invitations/accept | Accept a membership invite|
|[**addNamespaceMember**](#addnamespacemember) | **POST** /api/namespaces/{tenant}/members | Invite member|
|[**apiKeyCreate**](#apikeycreate) | **POST** /api/namespaces/api-key | Creates an API key.|
|[**apiKeyDelete**](#apikeydelete) | **DELETE** /api/namespaces/api-key/{key} | Delete an API key|
|[**apiKeyList**](#apikeylist) | **GET** /api/namespaces/api-key | List API Keys|
|[**apiKeyUpdate**](#apikeyupdate) | **PATCH** /api/namespaces/api-key/{key} | Update an API key|
|[**cancelMembershipInvitation**](#cancelmembershipinvitation) | **DELETE** /api/namespaces/{tenant}/invitations/{user-id} | Cancel a pending membership invitation|
|[**createNamespace**](#createnamespace) | **POST** /api/namespaces | Create namespace|
|[**createNamespaceAdmin**](#createnamespaceadmin) | **POST** /admin/api/namespaces/{tenant} | Create namespace admin|
|[**declineInvite**](#declineinvite) | **PATCH** /api/namespaces/{tenant}/invitations/decline | Decline a membership invite|
|[**deleteNamespace**](#deletenamespace) | **DELETE** /api/namespaces/{tenant} | Delete namespace|
|[**deleteNamespaceAdmin**](#deletenamespaceadmin) | **DELETE** /admin/api/namespaces/{tenant} | Delete namespace admin|
|[**editNamespace**](#editnamespace) | **PUT** /api/namespaces/{tenant} | Edit namespace|
|[**editNamespaceAdmin**](#editnamespaceadmin) | **PUT** /admin/api/namespaces-update/{tenantID} | Edit namespace admin|
|[**exportNamespaces**](#exportnamespaces) | **GET** /admin/api/export/namespaces | export namespace|
|[**generateInvitationLink**](#generateinvitationlink) | **POST** /api/namespaces/{tenant}/invitations/links | Generate an invitation link for a namespace member|
|[**getMembershipInvitationList**](#getmembershipinvitationlist) | **GET** /api/users/invitations | Get membership invitations for the authenticated user|
|[**getNamespace**](#getnamespace) | **GET** /api/namespaces/{tenant} | Get a namespace|
|[**getNamespaceAdmin**](#getnamespaceadmin) | **GET** /admin/api/namespaces/{tenant} | Get namespace admin|
|[**getNamespaceMembershipInvitationList**](#getnamespacemembershipinvitationlist) | **GET** /api/namespaces/{tenant}/invitations | Get membership invitations for a namespace|
|[**getNamespaceSupport**](#getnamespacesupport) | **GET** /api/namespaces/{tenant}/support | Get a namespace support identifier.|
|[**getNamespaceToken**](#getnamespacetoken) | **GET** /api/auth/token/{tenant} | Get a new namespace\&#39;s token|
|[**getNamespaces**](#getnamespaces) | **GET** /api/namespaces | Get namespaces list|
|[**getNamespacesAdmin**](#getnamespacesadmin) | **GET** /admin/api/namespaces | Get namespaces admin|
|[**leaveNamespace**](#leavenamespace) | **DELETE** /api/namespaces/{tenant}/members | Leave Namespace|
|[**lookupUserStatus**](#lookupuserstatus) | **GET** /api/namespaces/{tenant}/members/{id}/accept-invite | Lookup User\&#39;s Status|
|[**lookupUserStatus_0**](#lookupuserstatus_0) | **GET** /api/namespaces/{tenant}/members/{id}/accept-invite | Lookup User\&#39;s Status|
|[**removeNamespaceMember**](#removenamespacemember) | **DELETE** /api/namespaces/{tenant}/members/{uid} | Remove a member from a namespace|
|[**updateMembershipInvitation**](#updatemembershipinvitation) | **PATCH** /api/namespaces/{tenant}/invitations/{user-id} | Update a pending membership invitation|
|[**updateNamespaceMember**](#updatenamespacemember) | **PATCH** /api/namespaces/{tenant}/members/{uid} | Update a member from a namespace|

# **acceptInvite**
> acceptInvite()

Accepts a pending membership invitation for the authenticated user. The user must be logged into the account that was invited. 

### Example

```typescript
import {
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)

const { status, data } = await apiInstance.acceptInvite(
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

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Invitation successfully accepted |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **addNamespaceMember**
> Namespace addNamespaceMember()

Invites a member to a namespace.  In enterprise and community instances, the member will automatically accept the invite and will have an `accepted` status.  In cloud instances, the member will have a `pending` status until they accept the invite via an email sent to them. The invite is valid for **7 days**. If the member was previously invited and the invite is no longer valid, the same route will resend the invite. 

### Example

```typescript
import {
    NamespacesApi,
    Configuration,
    AddNamespaceMemberRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

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
    NamespacesApi,
    Configuration,
    ApiKeyCreate
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

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
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

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
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

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
    NamespacesApi,
    Configuration,
    ApiKeyUpdate
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

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

# **cancelMembershipInvitation**
> cancelMembershipInvitation()

Allows namespace administrators to cancel a pending membership invitation. The invitation status will be updated to \"cancelled\". The active user must have authority over the role of the invitation being cancelled. 

### Example

```typescript
import {
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let userId: string; //The ID of the invited user (default to undefined)

const { status, data } = await apiInstance.cancelMembershipInvitation(
    tenant,
    userId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|
| **userId** | [**string**] | The ID of the invited user | defaults to undefined|


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
|**200** | Invitation successfully cancelled |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createNamespace**
> Namespace createNamespace()

Create a namespace.

### Example

```typescript
import {
    NamespacesApi,
    Configuration,
    CreateNamespaceRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

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

# **createNamespaceAdmin**
> Namespace createNamespaceAdmin()

Create a namespace.

### Example

```typescript
import {
    NamespacesApi,
    Configuration,
    CreateNamespaceRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let createNamespaceRequest: CreateNamespaceRequest; // (optional)

const { status, data } = await apiInstance.createNamespaceAdmin(
    tenant,
    createNamespaceRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createNamespaceRequest** | **CreateNamespaceRequest**|  | |
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
|**200** | Success to create a namespace. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **declineInvite**
> declineInvite()

Declines a pending membership invitation for the authenticated user. The user must be logged into the account that was invited. The invitation status will be updated to \"rejected\". 

### Example

```typescript
import {
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)

const { status, data } = await apiInstance.declineInvite(
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

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Invitation successfully declined |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteNamespace**
> deleteNamespace()

Delete a namespace.

### Example

```typescript
import {
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

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

# **deleteNamespaceAdmin**
> deleteNamespaceAdmin()

Delete a namespace.

### Example

```typescript
import {
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)

const { status, data } = await apiInstance.deleteNamespaceAdmin(
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

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to delete a namespace. |  -  |
|**401** | Unauthorized |  -  |
|**402** | Payment required |  -  |
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
    NamespacesApi,
    Configuration,
    EditNamespaceRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

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

# **editNamespaceAdmin**
> editNamespaceAdmin()

Edit a namespace.

### Example

```typescript
import {
    NamespacesApi,
    Configuration,
    Namespace
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

let tenantID: string; //Namespace\'s tenant ID (default to undefined)
let namespace: Namespace; // (optional)

const { status, data } = await apiInstance.editNamespaceAdmin(
    tenantID,
    namespace
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **namespace** | **Namespace**|  | |
| **tenantID** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|


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
|**200** | Success to edit a namespace. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **exportNamespaces**
> File exportNamespaces()

Export namespaces to csv file. This endpoint has been deprecated and will be removed in v1.0.0. 

### Example

```typescript
import {
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

let filter: string; //Namespace\'s filter   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called `type`, it will filter by a `property` called `devices` where the value should be \'gt\' `0`.  An example of JSON object will looks like this:  ```json   [     {       \"type\":\"property\",       \"params\":         {           \"name\":\"devices\",           \"operator\":\"eq\",           \"value\":\"0\"         }     }   ] ```  So, the output encoded string will result on: `W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJkZXZpY2VzIiwib3BlcmF0b3IiOiJndCIsInZhbHVlIjoiMCJ9fV0=`  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.exportNamespaces(
    filter,
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **filter** | [**string**] | Namespace\&#39;s filter   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called &#x60;type&#x60;, it will filter by a &#x60;property&#x60; called &#x60;devices&#x60; where the value should be \&#39;gt\&#39; &#x60;0&#x60;.  An example of JSON object will looks like this:  &#x60;&#x60;&#x60;json   [     {       \&quot;type\&quot;:\&quot;property\&quot;,       \&quot;params\&quot;:         {           \&quot;name\&quot;:\&quot;devices\&quot;,           \&quot;operator\&quot;:\&quot;eq\&quot;,           \&quot;value\&quot;:\&quot;0\&quot;         }     }   ] &#x60;&#x60;&#x60;  So, the output encoded string will result on: &#x60;W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJkZXZpY2VzIiwib3BlcmF0b3IiOiJndCIsInZhbHVlIjoiMCJ9fV0&#x3D;&#x60;  | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|


### Return type

**File**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/octet-stream, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to export namespaces. |  -  |
|**204** | No content. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **generateInvitationLink**
> GenerateInvitationLink200Response generateInvitationLink()

Generates a unique invitation link to invite a member to a namespace using their email. Each invitation link is unique and tied to the provided email. Upon accepting the invitation, the user\'s status will automatically be set to `accepted`. If the user associated with the email does not exist, the invitation link will redirect them to the signup page.  The invitation remains valid for **7 days**. 

### Example

```typescript
import {
    NamespacesApi,
    Configuration,
    AddNamespaceMemberRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let addNamespaceMemberRequest: AddNamespaceMemberRequest; // (optional)

const { status, data } = await apiInstance.generateInvitationLink(
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

**GenerateInvitationLink200Response**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to generate a invitation link. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getMembershipInvitationList**
> Array<MembershipInvitation> getMembershipInvitationList()

Returns a paginated list of membership invitations for the authenticated user. This endpoint allows users to view all namespace invitations they have received. 

### Example

```typescript
import {
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

let filter: string; //Membership invitations filter.  Filter field receives a base64 encoded JSON object to limit the search.  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.getMembershipInvitationList(
    filter,
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **filter** | [**string**] | Membership invitations filter.  Filter field receives a base64 encoded JSON object to limit the search.  | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|


### Return type

**Array<MembershipInvitation>**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Successfully retrieved membership invitations list. |  * X-Total-Count - Total number of membership invitations. <br>  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getNamespace**
> Namespace getNamespace()

Get a namespace.

### Example

```typescript
import {
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

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

# **getNamespaceAdmin**
> Namespace getNamespaceAdmin()

Get a namespace.

### Example

```typescript
import {
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)

const { status, data } = await apiInstance.getNamespaceAdmin(
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

# **getNamespaceMembershipInvitationList**
> Array<MembershipInvitation> getNamespaceMembershipInvitationList()

Returns a paginated list of membership invitations for the specified namespace. This endpoint allows namespace administrators to view all pending invitations. 

### Example

```typescript
import {
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let filter: string; //Membership invitations filter.  Filter field receives a base64 encoded JSON object to limit the search.  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.getNamespaceMembershipInvitationList(
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
| **filter** | [**string**] | Membership invitations filter.  Filter field receives a base64 encoded JSON object to limit the search.  | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|


### Return type

**Array<MembershipInvitation>**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Successfully retrieved namespace membership invitations list. |  * X-Total-Count - Total number of membership invitations. <br>  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getNamespaceSupport**
> Support getNamespaceSupport()

Get a namespace support identifier.

### Example

```typescript
import {
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)

const { status, data } = await apiInstance.getNamespaceSupport(
    tenant
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|


### Return type

**Support**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a namespace support identifier. |  -  |
|**401** | Unauthorized |  -  |
|**402** | Payment required |  -  |
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
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

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
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

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

# **getNamespacesAdmin**
> Array<Namespace> getNamespacesAdmin()

Returns a list of namespaces.

### Example

```typescript
import {
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

let filter: string; //Namespaces\'s filter.   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called `type`, it will filter by a `property` called `name` where the value should `contains` `examplespace`.  If you want get only Namespaces name as `examplespace`, the JSON object will looks like this   ```json [   {     \"type\":\"property\",     \"params\":{       \"name\":\"name\",       \"operator\":\"contains\",       \"value\":\"examplespace\"     }   } ] ```  So, the output encoded string will result on: `W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lIiwib3BlcmF0b3IiOiJjb250YWlucyIsInZhbHVlIjoiZXhhbXBsZXNwYWNlIn19XQ==`  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.getNamespacesAdmin(
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

# **leaveNamespace**
> UserAuth leaveNamespace()

Allows the authenticated user to leave the specified namespace. Owners cannot leave a namespace; they must delete it instead. If the user attempts to leave their current authenticated namespace, the response will provide a new token that excludes this namespace. 

### Example

```typescript
import {
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

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

# **lookupUserStatus**
> LookupUserStatus200Response lookupUserStatus()

Clients may need to check a user\'s status before deciding whether to redirect to the accept-invite workflow or to the signup process. It is intended for use exclusively by clients in the `invite-member` pipeline. 

### Example

```typescript
import {
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

let tenant: string; //The tenant ID of the namespace. (default to undefined)
let id: string; //The user\'s ID. (default to undefined)
let sig: string; //The signature included in the email. This is used instead of the user\'s token to authenticate the request. (default to undefined)

const { status, data } = await apiInstance.lookupUserStatus(
    tenant,
    id,
    sig
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tenant** | [**string**] | The tenant ID of the namespace. | defaults to undefined|
| **id** | [**string**] | The user\&#39;s ID. | defaults to undefined|
| **sig** | [**string**] | The signature included in the email. This is used instead of the user\&#39;s token to authenticate the request. | defaults to undefined|


### Return type

**LookupUserStatus200Response**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **lookupUserStatus_0**
> LookupUserStatus200Response lookupUserStatus_0()

Clients may need to check a user\'s status before deciding whether to redirect to the accept-invite workflow or to the signup process. It is intended for use exclusively by clients in the `invite-member` pipeline. 

### Example

```typescript
import {
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

let tenant: string; //The tenant ID of the namespace. (default to undefined)
let id: string; //The user\'s ID. (default to undefined)
let sig: string; //The signature included in the email. This is used instead of the user\'s token to authenticate the request. (default to undefined)

const { status, data } = await apiInstance.lookupUserStatus_0(
    tenant,
    id,
    sig
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tenant** | [**string**] | The tenant ID of the namespace. | defaults to undefined|
| **id** | [**string**] | The user\&#39;s ID. | defaults to undefined|
| **sig** | [**string**] | The signature included in the email. This is used instead of the user\&#39;s token to authenticate the request. | defaults to undefined|


### Return type

**LookupUserStatus200Response**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success |  -  |
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
    NamespacesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

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

# **updateMembershipInvitation**
> updateMembershipInvitation()

Allows namespace administrators to update a pending membership invitation. Currently supports updating the role assigned to the invitation. The active user must have authority over the role being assigned. 

### Example

```typescript
import {
    NamespacesApi,
    Configuration,
    UpdateNamespaceMemberRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let userId: string; //The ID of the invited user (default to undefined)
let updateNamespaceMemberRequest: UpdateNamespaceMemberRequest; // (optional)

const { status, data } = await apiInstance.updateMembershipInvitation(
    tenant,
    userId,
    updateNamespaceMemberRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateNamespaceMemberRequest** | **UpdateNamespaceMemberRequest**|  | |
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|
| **userId** | [**string**] | The ID of the invited user | defaults to undefined|


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
|**200** | Invitation successfully updated |  -  |
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
    NamespacesApi,
    Configuration,
    UpdateNamespaceMemberRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new NamespacesApi(configuration);

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

