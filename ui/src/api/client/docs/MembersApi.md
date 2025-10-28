# MembersApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**acceptInvite**](#acceptinvite) | **PATCH** /api/namespaces/{tenant}/members/accept-invite | Accept a membership invite|
|[**addNamespaceMember**](#addnamespacemember) | **POST** /api/namespaces/{tenant}/members | Invite member|
|[**generateInvitationLink**](#generateinvitationlink) | **POST** /api/namespaces/{tenant}/members/invites | Generate an invitation link for a namespace member|
|[**leaveNamespace**](#leavenamespace) | **DELETE** /api/namespaces/{tenant}/members | Leave Namespace|
|[**lookupUserStatus**](#lookupuserstatus) | **GET** /api/namespaces/{tenant}/members/{id}/accept-invite | Lookup User\&#39;s Status|

# **acceptInvite**
> acceptInvite()

This route is intended to be accessed directly through the link sent in the invitation email. The user must be logged into the account that was invited. 

### Example

```typescript
import {
    MembersApi,
    Configuration,
    AcceptInviteRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new MembersApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let acceptInviteRequest: AcceptInviteRequest; // (optional)

const { status, data } = await apiInstance.acceptInvite(
    tenant,
    acceptInviteRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **acceptInviteRequest** | **AcceptInviteRequest**|  | |
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
    MembersApi,
    Configuration,
    AddNamespaceMemberRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new MembersApi(configuration);

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

# **generateInvitationLink**
> GenerateInvitationLink200Response generateInvitationLink()

Generates a unique invitation link to invite a member to a namespace using their email. Each invitation link is unique and tied to the provided email. Upon accepting the invitation, the user\'s status will automatically be set to `accepted`. If the user associated with the email does not exist, the invitation link will redirect them to the signup page.  The invitation remains valid for **7 days**. 

### Example

```typescript
import {
    MembersApi,
    Configuration,
    AddNamespaceMemberRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new MembersApi(configuration);

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

# **leaveNamespace**
> UserAuth leaveNamespace()

Allows the authenticated user to leave the specified namespace. Owners cannot leave a namespace; they must delete it instead. If the user attempts to leave their current authenticated namespace, the response will provide a new token that excludes this namespace. 

### Example

```typescript
import {
    MembersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new MembersApi(configuration);

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
    MembersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new MembersApi(configuration);

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

