# MembersApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**acceptInvite**](#acceptinvite) | **PATCH** /api/namespaces/{tenant}/invitations/accept | Accept a membership invite|
|[**addNamespaceMember**](#addnamespacemember) | **POST** /api/namespaces/{tenant}/members | Invite member|
|[**cancelMembershipInvitation**](#cancelmembershipinvitation) | **DELETE** /api/namespaces/{tenant}/invitations/{user-id} | Cancel a pending membership invitation|
|[**declineInvite**](#declineinvite) | **PATCH** /api/namespaces/{tenant}/invitations/decline | Decline a membership invite|
|[**generateInvitationLink**](#generateinvitationlink) | **POST** /api/namespaces/{tenant}/invitations/links | Generate an invitation link for a namespace member|
|[**getMembershipInvitationList**](#getmembershipinvitationlist) | **GET** /api/users/invitations | Get membership invitations for the authenticated user|
|[**getNamespaceMembershipInvitationList**](#getnamespacemembershipinvitationlist) | **GET** /api/namespaces/{tenant}/invitations | Get membership invitations for a namespace|
|[**leaveNamespace**](#leavenamespace) | **DELETE** /api/namespaces/{tenant}/members | Leave Namespace|
|[**lookupUserStatus**](#lookupuserstatus) | **GET** /api/namespaces/{tenant}/members/{id}/accept-invite | Lookup User\&#39;s Status|
|[**updateMembershipInvitation**](#updatemembershipinvitation) | **PATCH** /api/namespaces/{tenant}/invitations/{user-id} | Update a pending membership invitation|

# **acceptInvite**
> acceptInvite()

Accepts a pending membership invitation for the authenticated user. The user must be logged into the account that was invited. 

### Example

```typescript
import {
    MembersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new MembersApi(configuration);

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

# **cancelMembershipInvitation**
> cancelMembershipInvitation()

Allows namespace administrators to cancel a pending membership invitation. The invitation status will be updated to \"cancelled\". The active user must have authority over the role of the invitation being cancelled. 

### Example

```typescript
import {
    MembersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new MembersApi(configuration);

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

# **declineInvite**
> declineInvite()

Declines a pending membership invitation for the authenticated user. The user must be logged into the account that was invited. The invitation status will be updated to \"rejected\". 

### Example

```typescript
import {
    MembersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new MembersApi(configuration);

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

# **getMembershipInvitationList**
> Array<MembershipInvitation> getMembershipInvitationList()

Returns a paginated list of membership invitations for the authenticated user. This endpoint allows users to view all namespace invitations they have received. 

### Example

```typescript
import {
    MembersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new MembersApi(configuration);

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
|**200** | Successfully retrieved membership invitations list. |  * X-Total-Count -  <br>  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getNamespaceMembershipInvitationList**
> Array<MembershipInvitation> getNamespaceMembershipInvitationList()

Returns a paginated list of membership invitations for the specified namespace. This endpoint allows namespace administrators to view all pending invitations. 

### Example

```typescript
import {
    MembersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new MembersApi(configuration);

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
|**200** | Successfully retrieved namespace membership invitations list. |  * X-Total-Count -  <br>  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
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

# **updateMembershipInvitation**
> updateMembershipInvitation()

Allows namespace administrators to update a pending membership invitation. Currently supports updating the role assigned to the invitation. The active user must have authority over the role being assigned. 

### Example

```typescript
import {
    MembersApi,
    Configuration,
    UpdateNamespaceMemberRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new MembersApi(configuration);

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

