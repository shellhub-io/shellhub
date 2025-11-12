# UsersApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**acceptInvite**](#acceptinvite) | **PATCH** /api/namespaces/{tenant}/members/accept-invite | Accept a membership invite|
|[**adminDeleteUser**](#admindeleteuser) | **DELETE** /admin/api/users/{id} | Delete user|
|[**adminResetUserPassword**](#adminresetuserpassword) | **PATCH** /admin/api/users/{id}/password/reset | Reset user password|
|[**adminUpdateUser**](#adminupdateuser) | **PUT** /admin/api/users/{id} | Update user|
|[**authMFA**](#authmfa) | **POST** /api/user/mfa/auth | Auth MFA|
|[**authUser**](#authuser) | **POST** /api/auth/user | Auth a user|
|[**checkSessionRecord**](#checksessionrecord) | **GET** /api/users/security | Check session record status|
|[**createUserAdmin**](#createuseradmin) | **POST** /admin/api/users | Create a User admin|
|[**deleteUser**](#deleteuser) | **DELETE** /api/user | Delete user|
|[**disableMFA**](#disablemfa) | **PUT** /api/user/mfa/disable | Disable MFA|
|[**enableMFA**](#enablemfa) | **PUT** /api/user/mfa/enable | Enable MFA|
|[**exportUsers**](#exportusers) | **GET** /admin/api/export/users | export users|
|[**getSamlAuthUrl**](#getsamlauthurl) | **GET** /api/user/saml/auth | Get SAML authentication URL|
|[**getToken**](#gettoken) | **GET** /api/token/{tenant} | Get token|
|[**getUser**](#getuser) | **GET** /admin/api/users/{id} | Get user|
|[**getUserInfo**](#getuserinfo) | **GET** /api/auth/user | Get user info|
|[**getUserTokenAdmin**](#getusertokenadmin) | **GET** /admin/api/auth/token/{id} | Get user token|
|[**getUsers**](#getusers) | **GET** /admin/api/users | Get users|
|[**getValidateAccount**](#getvalidateaccount) | **GET** /api/user/validation_account | Validate activation link|
|[**login**](#login) | **POST** /api/login | Login|
|[**mfaRecover**](#mfarecover) | **POST** /api/user/mfa/recover | Recover MFA|
|[**recoverPassword**](#recoverpassword) | **POST** /api/user/recover_password | Recover password|
|[**registerUser**](#registeruser) | **POST** /api/register | Register a new user|
|[**requestResetMFA**](#requestresetmfa) | **POST** /api/user/mfa/reset | Request Reset MFA|
|[**resendEmail**](#resendemail) | **POST** /api/user/resend_email | Resend confirmation|
|[**resetMFA**](#resetmfa) | **PUT** /api/user/mfa/reset/{user-id} | Reset MFA|
|[**setSessionRecord**](#setsessionrecord) | **PUT** /api/users/security/{tenant} | Set session record|
|[**updateRecoverPassword**](#updaterecoverpassword) | **POST** /api/user/{uid}/update_password | Update user password|
|[**updateUser**](#updateuser) | **PATCH** /api/users | Update user|
|[**updateUserData**](#updateuserdata) | **PATCH** /api/users/{id}/data | Update user data|
|[**updateUserPassword**](#updateuserpassword) | **PATCH** /api/users/{id}/password | Update user password|

# **acceptInvite**
> acceptInvite()

This route is intended to be accessed directly through the link sent in the invitation email. The user must be logged into the account that was invited. 

### Example

```typescript
import {
    UsersApi,
    Configuration,
    AcceptInviteRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

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

# **adminDeleteUser**
> adminDeleteUser()

Delete a user.

### Example

```typescript
import {
    UsersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.adminDeleteUser(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] |  | defaults to undefined|


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
|**200** | Success |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **adminResetUserPassword**
> AdminResetUserPassword200Response adminResetUserPassword()

Resets the password for a specified user. It\'s particularly useful for users who initially authenticated via SAML and therefore may not have a password set up. This enables them to maintain access even if their original authentication method becomes unavailable (e.g., if SAML authentication is disabled).  The endpoint generates a secure 16-character random password that includes: - Uppercase letters - Lowercase letters - Numbers - Special characters  Users are strongly encouraged to change this temporary password after their first successful authentication. If the user already has a password, a `400 Bad Request` status code will be returned. 

### Example

```typescript
import {
    UsersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let id: string; //The ID of the user whose password needs to be reset (default to undefined)

const { status, data } = await apiInstance.adminResetUserPassword(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] | The ID of the user whose password needs to be reset | defaults to undefined|


### Return type

**AdminResetUserPassword200Response**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Password successfully reset |  -  |
|**400** | Invalid Fields |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **adminUpdateUser**
> adminUpdateUser()

Update a user.

### Example

```typescript
import {
    UsersApi,
    Configuration,
    UserAdminRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let id: string; // (default to undefined)
let userAdminRequest: UserAdminRequest; // (optional)

const { status, data } = await apiInstance.adminUpdateUser(
    id,
    userAdminRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **userAdminRequest** | **UserAdminRequest**|  | |
| **id** | [**string**] |  | defaults to undefined|


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
|**400** | Invalid Fields |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**409** | Conflict Fields |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **authMFA**
> UserAuth authMFA()

Authenticate a user who has MFA enabled. This endpoint should be called after the default authUser endpoint, which generates an `X-MFA-Token` indicating that the user has already authenticated with a password. 

### Example

```typescript
import {
    UsersApi,
    Configuration,
    MfaAuth
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let mfaAuth: MfaAuth; // (optional)

const { status, data } = await apiInstance.authMFA(
    mfaAuth
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **mfaAuth** | **MfaAuth**|  | |


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
|**200** | Success to authenticate. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **authUser**
> UserAuth authUser()

Authenticate a user, returning the session\'s JWT token and data about the user.

### Example

```typescript
import {
    UsersApi,
    Configuration,
    LoginRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

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
    UsersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

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

# **createUserAdmin**
> createUserAdmin()

Create a User.

### Example

```typescript
import {
    UsersApi,
    Configuration,
    UserAdminRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let userAdminRequest: UserAdminRequest; // (optional)

const { status, data } = await apiInstance.createUserAdmin(
    userAdminRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **userAdminRequest** | **UserAdminRequest**|  | |


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
|**400** | Invalid Fields |  -  |
|**401** | Unauthorized |  -  |
|**409** | Conflict Fields |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteUser**
> deleteUser()

Deletes the authenticated user. The user will be removed from any namespaces they are a member of. Users who are owners of namespaces cannot be deleted. In such cases, the user must delete the namespace(s) first.  > NOTE: This route is available only for **cloud** instances. Enterprise users must use the admin console, and community users must use the CLI. 

### Example

```typescript
import {
    UsersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

const { status, data } = await apiInstance.deleteUser();
```

### Parameters
This endpoint does not have any parameters.


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
|**204** | Success to delete the user |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **disableMFA**
> disableMFA()

Disable MFA for a user. To disable MFA, the user must provide either a recovery code or the current MFA code. If a recovery code is used, it will be invalidated for future use.  The recovery code used to regain access to the account can be used within a 10-minute window on this endpoint. 

### Example

```typescript
import {
    UsersApi,
    Configuration,
    MfaDisable
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let mfaDisable: MfaDisable; // (optional)

const { status, data } = await apiInstance.disableMFA(
    mfaDisable
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **mfaDisable** | **MfaDisable**|  | |


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
|**200** | Success to disable a MFA status. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **enableMFA**
> enableMFA()

Enable MFA for a user. The secret and recovery codes must be created by the generateMFA endpoint. Users with MFA already enabled cannot override their MFA credentials; in these cases, a user must disable MFA before proceeding. The recovery e-mail must be a valid value in order to enable the MFA. 

### Example

```typescript
import {
    UsersApi,
    Configuration,
    MfaEnable
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let mfaEnable: MfaEnable; // (optional)

const { status, data } = await apiInstance.enableMFA(
    mfaEnable
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **mfaEnable** | **MfaEnable**|  | |


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
|**200** | Success to enable a MFA. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **exportUsers**
> File exportUsers()

Export users to csv file.

### Example

```typescript
import {
    UsersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let filter: string; //User\'s filter   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called `type`, it will filter by a `property` called `namespaces` where the value should be `eq` to `0`.  An example of JSON object will looks like this:  ```json   [     {       \"type\":\"property\",       \"params\":         {           \"name\":\"namespace\",           \"operator\":\"eq\",           \"value\":\"0\"         }     }   ] ```  So, the output encoded string will result on: `W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lc3BhY2VzIiwib3BlcmF0b3IiOiJndCIsInZhbHVlIjoiMCJ9fV0=`  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.exportUsers(
    filter,
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **filter** | [**string**] | User\&#39;s filter   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called &#x60;type&#x60;, it will filter by a &#x60;property&#x60; called &#x60;namespaces&#x60; where the value should be &#x60;eq&#x60; to &#x60;0&#x60;.  An example of JSON object will looks like this:  &#x60;&#x60;&#x60;json   [     {       \&quot;type\&quot;:\&quot;property\&quot;,       \&quot;params\&quot;:         {           \&quot;name\&quot;:\&quot;namespace\&quot;,           \&quot;operator\&quot;:\&quot;eq\&quot;,           \&quot;value\&quot;:\&quot;0\&quot;         }     }   ] &#x60;&#x60;&#x60;  So, the output encoded string will result on: &#x60;W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lc3BhY2VzIiwib3BlcmF0b3IiOiJndCIsInZhbHVlIjoiMCJ9fV0&#x3D;&#x60;  | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|


### Return type

**File**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/octet-stream, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to export users. |  -  |
|**204** | No content. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getSamlAuthUrl**
> GetSamlAuthUrl200Response getSamlAuthUrl()

Retrieves the Identity Provider (IdP) URL for authentication within ShellHub. After successful authentication, users are automatically redirected to the ShellHub dashboard.  To access this endpoint, SAML authentication must be enabled and the instance must be running the Enterprise edition. If not, the endpoint returns a `501 Not Implemented` status code. 

### Example

```typescript
import {
    UsersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

const { status, data } = await apiInstance.getSamlAuthUrl();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**GetSamlAuthUrl200Response**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Successfully retrieved the SAML authentication URL |  -  |
|**400** | Bad request |  -  |
|**500** | Internal error |  -  |
|**501** | SAML authentication is not enabled or instance is not running Enterprise edition |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getToken**
> UserAuth getToken()

Get a token from its tenant.

### Example

```typescript
import {
    UsersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

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

# **getUser**
> GetUser200Response getUser()

Get a user.

### Example

```typescript
import {
    UsersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.getUser(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] |  | defaults to undefined|


### Return type

**GetUser200Response**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a user. |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getUserInfo**
> UserAuth getUserInfo()


### Example

```typescript
import {
    UsersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

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

# **getUserTokenAdmin**
> GetUserTokenAdmin200Response getUserTokenAdmin()

Get user JWT token to login.

### Example

```typescript
import {
    UsersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let id: string; //User\'s ID (default to undefined)

const { status, data } = await apiInstance.getUserTokenAdmin(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] | User\&#39;s ID | defaults to undefined|


### Return type

**GetUserTokenAdmin200Response**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get the JWT token to login. |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not Found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getUsers**
> Array<UserAdminResponse> getUsers()

Get a list of users.

### Example

```typescript
import {
    UsersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let filter: string; //Filter field receives a JSON object enconded as base64 string for limit a search.  The JSON enconded must follow these interafaces: ```typescript interface ParamProperty {   name: string;   operator: \"contains\" | \"eq\" | \"bool\" | \"gt\" | \"lt\";   value: string; }  interface ParamOperator {   name: \"and\" | \"or\"; }  interface Filter {   type: \"property\" | \"operator\";   param: ParamOperator | ParamProperty; }  interface FilterList {   Filters: Array<Filter>; }  ```  ## Examples  This is a example to filter and get only the resource what property \"confirmed\" is \"true\" ```json [   {   \"type\": \"property\",   \"params\": {       \"name\": \"confirmed\",       \"operator\": \"bool\",       \"value\": \"true\"       }   } ] ```  This one, filter resource by the property \"id\" inside \"info\" structure when it is equal to \"manjaro\" and online property is set to \"true\" ```json [   {     \"type\": \"property\",     \"params\": {       \"name\": \"info.id\",       \"operator\": \"eq\",       \"value\": \"manjaro\"     }   },   {     \"type\": \"property\",     \"params\": {       \"name\": \"online\",       \"operator\": \"bool\",       \"value\": \"true\"     }   },   {     \"type\": \"operator\",     \"params\": {       \"name\": \"and\"     }   } ] ```  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.getUsers(
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

**Array<UserAdminResponse>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a list of users. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getValidateAccount**
> getValidateAccount()

Validate the activation link for user.

### Example

```typescript
import {
    UsersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let email: string; //User\'s email. (default to undefined)
let token: string; //User\'s validation token.   It is a token received from the email used to validate the user. (default to undefined)

const { status, data } = await apiInstance.getValidateAccount(
    email,
    token
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **email** | [**string**] | User\&#39;s email. | defaults to undefined|
| **token** | [**string**] | User\&#39;s validation token.   It is a token received from the email used to validate the user. | defaults to undefined|


### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to validate user. |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **login**
> UserAuth login()

Authenticate a \"local\" user by returning the session\'s JWT token and user data. Local users are those registered via the ShellHub form without relying on external Identity Providers (IdPs).  Authentication may result in an account lockout after N consecutive incorrect login attempts. The lockout applies specifically to a particular source and user combination. Check for the presence of the `X-Account-Lockout` header to determine the account lockout status. When it\'s 0, there are no active lockouts.  Users with MFA enabled cannot authenticate via this route. In such cases, the API will respond with a status `401` and an `X-MFA-Token` header with a UUID. Authentication must be med to `/api/mfa/auth` with this token in these instances. 

### Example

```typescript
import {
    UsersApi,
    Configuration,
    LoginRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

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

# **mfaRecover**
> UserAuth mfaRecover()

Recover account access by providing one of the user\'s recovery codes. It will be invalidated for future uses.  The recovery code will be cached for 10 minutes. During this period, the user can use the same recovery code to disable their MFA without needing to provide two separate codes. The `X-Expires-At` header specifies the epoch value marking the end of the cache period. 

### Example

```typescript
import {
    UsersApi,
    Configuration,
    MfaRecover
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let mfaRecover: MfaRecover; // (optional)

const { status, data } = await apiInstance.mfaRecover(
    mfaRecover
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **mfaRecover** | **MfaRecover**|  | |


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
|**200** | Success recover account\&#39;s access. |  * X-Expires-At - The epoch time at which the recovery code will become invalid. <br>  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **recoverPassword**
> recoverPassword()

Send a recovery email to the user.

### Example

```typescript
import {
    UsersApi,
    Configuration,
    RecoverPasswordRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let recoverPasswordRequest: RecoverPasswordRequest; // (optional)

const { status, data } = await apiInstance.recoverPassword(
    recoverPasswordRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **recoverPasswordRequest** | **RecoverPasswordRequest**|  | |


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
|**200** | Success to send email to recover user password. |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **registerUser**
> registerUser()


### Example

```typescript
import {
    UsersApi,
    Configuration,
    RegisterUserRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let registerUserRequest: RegisterUserRequest; // (optional)

const { status, data } = await apiInstance.registerUser(
    registerUserRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **registerUserRequest** | **RegisterUserRequest**|  | |


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
|**200** | User registered successfully |  -  |
|**400** | Invalid Fields |  -  |
|**401** | Unauthorized |  -  |
|**409** | Conflict Fields |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **requestResetMFA**
> requestResetMFA()

Sends an email to both the user\'s main and recovery addresses. Each email contains a unique code, which remains valid for at most 1 day. The user must provide both codes to reset their MFA. 

### Example

```typescript
import {
    UsersApi,
    Configuration,
    RequestResetMFARequest
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let requestResetMFARequest: RequestResetMFARequest; // (optional)

const { status, data } = await apiInstance.requestResetMFA(
    requestResetMFARequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **requestResetMFARequest** | **RequestResetMFARequest**|  | |


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
|**200** | Success to send the email. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **resendEmail**
> resendEmail()

Resend confirmation to user.

### Example

```typescript
import {
    UsersApi,
    Configuration,
    ResendEmailRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let resendEmailRequest: ResendEmailRequest; // (optional)

const { status, data } = await apiInstance.resendEmail(
    resendEmailRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **resendEmailRequest** | **ResendEmailRequest**|  | |


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
|**200** | Success to resend confirmation to user. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **resetMFA**
> UserAuth resetMFA()

Similar to the `disableMFA` operation, this endpoint uses the two codes sent by `requestResetMFA` instead of a TOTP or recovery code. The user ID must be the same as the one used for `requestResetMFA`. 

### Example

```typescript
import {
    UsersApi,
    Configuration,
    MfaReset
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let userId: string; // (default to undefined)
let mfaReset: MfaReset; // (optional)

const { status, data } = await apiInstance.resetMFA(
    userId,
    mfaReset
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **mfaReset** | **MfaReset**|  | |
| **userId** | [**string**] |  | defaults to undefined|


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
|**200** | Success to reset and authenticate. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **setSessionRecord**
> setSessionRecord()

Define if sessions will be recorded.

### Example

```typescript
import {
    UsersApi,
    Configuration,
    SetSessionRecordRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

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

# **updateRecoverPassword**
> updateRecoverPassword()

Update user password from a recovery token got from email.

### Example

```typescript
import {
    UsersApi,
    Configuration,
    UpdateRecoverPasswordRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let uid: string; //User\'s UID. (default to undefined)
let updateRecoverPasswordRequest: UpdateRecoverPasswordRequest; // (optional)

const { status, data } = await apiInstance.updateRecoverPassword(
    uid,
    updateRecoverPasswordRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateRecoverPasswordRequest** | **UpdateRecoverPasswordRequest**|  | |
| **uid** | [**string**] | User\&#39;s UID. | defaults to undefined|


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
|**200** | Success to update user password. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateUser**
> updateUser()


### Example

```typescript
import {
    UsersApi,
    Configuration,
    UpdateUserRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

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
    UsersApi,
    Configuration,
    UpdateUserDataRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

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
    UsersApi,
    Configuration,
    UpdateUserPasswordRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

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

