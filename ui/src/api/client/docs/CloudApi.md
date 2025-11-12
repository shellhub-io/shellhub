# CloudApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**acceptInvite**](#acceptinvite) | **PATCH** /api/namespaces/{tenant}/members/accept-invite | Accept a membership invite|
|[**attachPaymentMethod**](#attachpaymentmethod) | **POST** /api/billing/paymentmethod/attach | Attach payment method|
|[**authMFA**](#authmfa) | **POST** /api/user/mfa/auth | Auth MFA|
|[**choiceDevices**](#choicedevices) | **POST** /api/billing/device-choice | Choice devices|
|[**clsoeSession**](#clsoesession) | **POST** /api/sessions/{uid}/close | Close session|
|[**connectorCreate**](#connectorcreate) | **POST** /api/connector | Connector\&#39;s create|
|[**connectorDelete**](#connectordelete) | **DELETE** /api/connector/{uid} | Connector\&#39;s delete|
|[**connectorGet**](#connectorget) | **GET** /api/connector/{uid} | Connector\&#39;s get|
|[**connectorInfo**](#connectorinfo) | **GET** /api/connector/{uid}/info | Connector\&#39;s get Docker info|
|[**connectorList**](#connectorlist) | **GET** /api/connector | Connector\&#39;s list|
|[**connectorUpdate**](#connectorupdate) | **PATCH** /api/connector/{uid} | Connector\&#39;s setting update|
|[**createAnnouncement**](#createannouncement) | **POST** /admin/api/announcements | Create an announcement|
|[**createCustomer**](#createcustomer) | **POST** /api/billing/customer | Create customer|
|[**createFirewallRule**](#createfirewallrule) | **POST** /api/firewall/rules | Create firewall rule|
|[**createSubscription**](#createsubscription) | **POST** /api/billing/subscription | Create subscription|
|[**createTunnel**](#createtunnel) | **POST** /api/devices/{uid}/tunnels | Create a tunnel|
|[**createWebEndpoint**](#createwebendpoint) | **POST** /api/web-endpoints | Create a web-endpoint|
|[**deleteAnnouncement**](#deleteannouncement) | **DELETE** /admin/api/announcements/{uuid} | Delete an announcement|
|[**deleteFirewallRule**](#deletefirewallrule) | **DELETE** /api/firewall/rules/{id} | Delete firewall rule|
|[**deleteSessionRecord**](#deletesessionrecord) | **DELETE** /api/sessions/{uid}/records/{seat} | Delete session record|
|[**deleteTunnel**](#deletetunnel) | **DELETE** /api/devices/{uid}/tunnels/{address} | Delete a tunnel|
|[**deleteUser**](#deleteuser) | **DELETE** /api/user | Delete user|
|[**deleteWebEndpoint**](#deletewebendpoint) | **DELETE** /api/web-endpoints/{address} | Delete a web-endpoint|
|[**detachPaymentMethod**](#detachpaymentmethod) | **POST** /api/billing/paymentmethod/detach | Detach payment method|
|[**disableMFA**](#disablemfa) | **PUT** /api/user/mfa/disable | Disable MFA|
|[**enableMFA**](#enablemfa) | **PUT** /api/user/mfa/enable | Enable MFA|
|[**evaluate**](#evaluate) | **POST** /api/billing/evaluate | Evaluate|
|[**generateInvitationLink**](#generateinvitationlink) | **POST** /api/namespaces/{tenant}/members/invites | Generate an invitation link for a namespace member|
|[**generateMFA**](#generatemfa) | **GET** /api/user/mfa/generate | Generate MFA Credentials|
|[**getAnnouncementAdmin**](#getannouncementadmin) | **GET** /admin/api/announcements/{uuid} | Get a announcement|
|[**getCustomer**](#getcustomer) | **GET** /api/billing/customer | Get Customer|
|[**getDevicesMostUsed**](#getdevicesmostused) | **GET** /api/billing/devices-most-used | Get devices most used|
|[**getFirewallRule**](#getfirewallrule) | **GET** /api/firewall/rules/{id} | Get firewall rule|
|[**getFirewallRules**](#getfirewallrules) | **GET** /api/firewall/rules | Get firewall rules|
|[**getNamespaceSupport**](#getnamespacesupport) | **GET** /api/namespaces/{tenant}/support | Get a namespace support identifier.|
|[**getSamlAuthUrl**](#getsamlauthurl) | **GET** /api/user/saml/auth | Get SAML authentication URL|
|[**getSessionRecord**](#getsessionrecord) | **GET** /api/sessions/{uid}/records/{seat} | Get session record|
|[**getSubscription**](#getsubscription) | **GET** /api/billing/subscription | Get subscription|
|[**getValidateAccount**](#getvalidateaccount) | **GET** /api/user/validation_account | Validate activation link|
|[**listAnnouncementsAdmin**](#listannouncementsadmin) | **GET** /admin/api/announcements | List announcements|
|[**listTunnels**](#listtunnels) | **GET** /api/devices/{uid}/tunnels | List tunnels|
|[**listWebEndpoints**](#listwebendpoints) | **GET** /api/web-endpoints | List web-endpoints|
|[**mfaRecover**](#mfarecover) | **POST** /api/user/mfa/recover | Recover MFA|
|[**recordSession**](#recordsession) | **POST** /api/sessions/{uid}/records/{seat} | Record session|
|[**recoverPassword**](#recoverpassword) | **POST** /api/user/recover_password | Recover password|
|[**registerUser**](#registeruser) | **POST** /api/register | Register a new user|
|[**report**](#report) | **POST** /api/billing/report | Report|
|[**requestResetMFA**](#requestresetmfa) | **POST** /api/user/mfa/reset | Request Reset MFA|
|[**resendEmail**](#resendemail) | **POST** /api/user/resend_email | Resend confirmation|
|[**resetMFA**](#resetmfa) | **PUT** /api/user/mfa/reset/{user-id} | Reset MFA|
|[**setDefaultPaymentMethod**](#setdefaultpaymentmethod) | **POST** /api/billing/paymentmethod/default | Set default payment method|
|[**updateAnnouncement**](#updateannouncement) | **PUT** /admin/api/announcements/{uuid} | Update an announcement|
|[**updateFirewallRule**](#updatefirewallrule) | **PUT** /api/firewall/rules/{id} | Update firewall rule|
|[**updateRecoverPassword**](#updaterecoverpassword) | **POST** /api/user/{uid}/update_password | Update user password|

# **acceptInvite**
> acceptInvite()

This route is intended to be accessed directly through the link sent in the invitation email. The user must be logged into the account that was invited. 

### Example

```typescript
import {
    CloudApi,
    Configuration,
    AcceptInviteRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

# **attachPaymentMethod**
> attachPaymentMethod()

Attachs a payment method to a customer.

### Example

```typescript
import {
    CloudApi,
    Configuration,
    AttachPaymentMethodRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let attachPaymentMethodRequest: AttachPaymentMethodRequest; // (optional)

const { status, data } = await apiInstance.attachPaymentMethod(
    attachPaymentMethodRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **attachPaymentMethodRequest** | **AttachPaymentMethodRequest**|  | |


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
|**200** | Success to attach the payment method to a customer. |  -  |
|**400** | Error |  -  |
|**401** | Error |  -  |
|**403** | Error |  -  |
|**404** | Error |  -  |
|**424** | Error |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **authMFA**
> UserAuth authMFA()

Authenticate a user who has MFA enabled. This endpoint should be called after the default authUser endpoint, which generates an `X-MFA-Token` indicating that the user has already authenticated with a password. 

### Example

```typescript
import {
    CloudApi,
    Configuration,
    MfaAuth
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

# **choiceDevices**
> choiceDevices()

Choice devices when device\'s limit is rechead.

### Example

```typescript
import {
    CloudApi,
    Configuration,
    ChoiceDevicesRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let choiceDevicesRequest: ChoiceDevicesRequest; // (optional)

const { status, data } = await apiInstance.choiceDevices(
    choiceDevicesRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **choiceDevicesRequest** | **ChoiceDevicesRequest**|  | |


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
|**200** | Success to choice devices. |  -  |
|**400** | Bad request |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **clsoeSession**
> clsoeSession()

Close a session.

### Example

```typescript
import {
    CloudApi,
    Configuration,
    ClsoeSessionRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

# **connectorCreate**
> connectorCreate(connectorData)

Create a new connector.

### Example

```typescript
import {
    CloudApi,
    Configuration,
    ConnectorData
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let connectorData: ConnectorData; //

const { status, data } = await apiInstance.connectorCreate(
    connectorData
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **connectorData** | **ConnectorData**|  | |


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
|**200** | Create a Connector and connected successfully |  -  |
|**201** | Create a Connector successfully, but failed to connect |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **connectorDelete**
> connectorDelete()

Deletes a connector.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let uid: string; //Connector UID (default to undefined)

const { status, data } = await apiInstance.connectorDelete(
    uid
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Connector UID | defaults to undefined|


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
|**200** | Delete the Connector successfully |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **connectorGet**
> Connector connectorGet()

Gets a connector.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let uid: string; //Connector UID (default to undefined)

const { status, data } = await apiInstance.connectorGet(
    uid
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Connector UID | defaults to undefined|


### Return type

**Connector**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Get the Connector successfully |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **connectorInfo**
> ConnectorInfo200Response connectorInfo()

Gets the connector\'s connection docker info.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let uid: string; //Connector UID (default to undefined)

const { status, data } = await apiInstance.connectorInfo(
    uid
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Connector UID | defaults to undefined|


### Return type

**ConnectorInfo200Response**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Get Connector\&#39;s connection Docker info successfully |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **connectorList**
> Array<Connector> connectorList()

List connectors.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let enable: boolean; //Enable status. (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.connectorList(
    enable,
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **enable** | [**boolean**] | Enable status. | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|


### Return type

**Array<Connector>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | List Connectors for a namespace successfully |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **connectorUpdate**
> connectorUpdate(connectorData)

Updates a connector settings.

### Example

```typescript
import {
    CloudApi,
    Configuration,
    ConnectorData
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let uid: string; //Connector UID (default to undefined)
let connectorData: ConnectorData; //

const { status, data } = await apiInstance.connectorUpdate(
    uid,
    connectorData
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **connectorData** | **ConnectorData**|  | |
| **uid** | [**string**] | Connector UID | defaults to undefined|


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
|**200** | Update the Connector settings successfully |  -  |
|**201** | Create a Connector successfully, but failed to connect |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createAnnouncement**
> Announcement createAnnouncement()

Create an announcement.

### Example

```typescript
import {
    CloudApi,
    Configuration,
    CreateAnnouncementRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let createAnnouncementRequest: CreateAnnouncementRequest; // (optional)

const { status, data } = await apiInstance.createAnnouncement(
    createAnnouncementRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createAnnouncementRequest** | **CreateAnnouncementRequest**|  | |


### Return type

**Announcement**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to create an announcement. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createCustomer**
> createCustomer()

creates a new customer defining, optionaly, the default payment method.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

const { status, data } = await apiInstance.createCustomer();
```

### Parameters
This endpoint does not have any parameters.


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
|**200** | Success to create a new customer. |  -  |
|**400** | Error |  -  |
|**401** | Error |  -  |
|**403** | Error |  -  |
|**404** | Error |  -  |
|**409** | Error |  -  |
|**424** | Error |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createFirewallRule**
> FirewallRulesResponse createFirewallRule()

Create a firewall rule.

### Example

```typescript
import {
    CloudApi,
    Configuration,
    FirewallRulesRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let firewallRulesRequest: FirewallRulesRequest; // (optional)

const { status, data } = await apiInstance.createFirewallRule(
    firewallRulesRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **firewallRulesRequest** | **FirewallRulesRequest**|  | |


### Return type

**FirewallRulesResponse**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to create firewall rule. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createSubscription**
> createSubscription()

Create a subscription.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

const { status, data } = await apiInstance.createSubscription();
```

### Parameters
This endpoint does not have any parameters.


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
|**200** | Success to create a new subscription. |  -  |
|**400** | Error |  -  |
|**402** | Error |  -  |
|**403** | Error |  -  |
|**404** | Error |  -  |
|**409** | Error |  -  |
|**424** | Error |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createTunnel**
> Tunnel createTunnel(createTunnelRequest)

Creates a new tunnel for a device.

### Example

```typescript
import {
    CloudApi,
    Configuration,
    CreateTunnelRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

# **createWebEndpoint**
> Webendpoint createWebEndpoint(createWebEndpointRequest)

Creates a new web-endpoint for a device.

### Example

```typescript
import {
    CloudApi,
    Configuration,
    CreateWebEndpointRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let createWebEndpointRequest: CreateWebEndpointRequest; //

const { status, data } = await apiInstance.createWebEndpoint(
    createWebEndpointRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createWebEndpointRequest** | **CreateWebEndpointRequest**|  | |


### Return type

**Webendpoint**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Web-endpoint created successfully. |  -  |
|**400** | Bad request |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteAnnouncement**
> Announcement deleteAnnouncement()

Delete an announcement.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let uuid: string; // (default to undefined)

const { status, data } = await apiInstance.deleteAnnouncement(
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

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to delete an announcement. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteFirewallRule**
> deleteFirewallRule()

Delete a firewall rule.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.deleteFirewallRule(
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
|**200** | Success to delete a firewall rule. |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteSessionRecord**
> deleteSessionRecord()

Deletes a session record based on its seat.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

# **deleteTunnel**
> deleteTunnel()

Deletes a tunnel for a specific device and port.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

# **deleteUser**
> deleteUser()

Deletes the authenticated user. The user will be removed from any namespaces they are a member of. Users who are owners of namespaces cannot be deleted. In such cases, the user must delete the namespace(s) first.  > NOTE: This route is available only for **cloud** instances. Enterprise users must use the admin console, and community users must use the CLI. 

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

# **deleteWebEndpoint**
> deleteWebEndpoint()

Deletes a web-endpoint by address.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let address: string; //Tunnel\'s address (default to undefined)

const { status, data } = await apiInstance.deleteWebEndpoint(
    address
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
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
|**200** | Web-endpoint deleted successfully. |  -  |
|**400** | Bad request |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **detachPaymentMethod**
> detachPaymentMethod()

Detachs a payment method from a customer.

### Example

```typescript
import {
    CloudApi,
    Configuration,
    AttachPaymentMethodRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let attachPaymentMethodRequest: AttachPaymentMethodRequest; // (optional)

const { status, data } = await apiInstance.detachPaymentMethod(
    attachPaymentMethodRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **attachPaymentMethodRequest** | **AttachPaymentMethodRequest**|  | |


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
|**200** | Success to detach the payment method from a customer. |  -  |
|**400** | Error |  -  |
|**401** | Error |  -  |
|**403** | Error |  -  |
|**404** | Error |  -  |
|**424** | Error |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **disableMFA**
> disableMFA()

Disable MFA for a user. To disable MFA, the user must provide either a recovery code or the current MFA code. If a recovery code is used, it will be invalidated for future use.  The recovery code used to regain access to the account can be used within a 10-minute window on this endpoint. 

### Example

```typescript
import {
    CloudApi,
    Configuration,
    MfaDisable
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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
    CloudApi,
    Configuration,
    MfaEnable
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

# **evaluate**
> Evaluate200Response evaluate()

evaluate the namespace capabilities.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

const { status, data } = await apiInstance.evaluate();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**Evaluate200Response**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to evaluate the namespace. |  -  |
|**400** | Bad Request. |  -  |
|**401** | Unauthorized. |  -  |
|**403** | Forbidden. |  -  |
|**404** | Not found. |  -  |
|**424** | Failed dependency. |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **generateInvitationLink**
> GenerateInvitationLink200Response generateInvitationLink()

Generates a unique invitation link to invite a member to a namespace using their email. Each invitation link is unique and tied to the provided email. Upon accepting the invitation, the user\'s status will automatically be set to `accepted`. If the user associated with the email does not exist, the invitation link will redirect them to the signup page.  The invitation remains valid for **7 days**. 

### Example

```typescript
import {
    CloudApi,
    Configuration,
    AddNamespaceMemberRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

# **generateMFA**
> MfaGenerate generateMFA()

Generate the credentials to enable a user\'s MFA. The user must save the recovery codes a secure manner. 

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

const { status, data } = await apiInstance.generateMFA();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**MfaGenerate**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Successfully generated MFA credentials. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAnnouncementAdmin**
> Announcement getAnnouncementAdmin()

Get a announcement.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let uuid: string; // (default to undefined)

const { status, data } = await apiInstance.getAnnouncementAdmin(
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

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

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

# **getCustomer**
> GetCustomer200Response getCustomer()

Get the customer.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

const { status, data } = await apiInstance.getCustomer();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**GetCustomer200Response**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a customer. |  -  |
|**400** | Error |  -  |
|**403** | Error |  -  |
|**404** | Error |  -  |
|**424** | Error |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getDevicesMostUsed**
> Array<Device> getDevicesMostUsed()

Get the most used devices.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

const { status, data } = await apiInstance.getDevicesMostUsed();
```

### Parameters
This endpoint does not have any parameters.


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
|**200** | Success to get the most used devices. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getFirewallRule**
> FirewallRulesResponse getFirewallRule()

Get a firewall rule.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.getFirewallRule(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] |  | defaults to undefined|


### Return type

**FirewallRulesResponse**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get firewall rule. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getFirewallRules**
> Array<FirewallRulesResponse> getFirewallRules()

Get a list of firewall rules.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.getFirewallRules(
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

**Array<FirewallRulesResponse>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get firewall rules. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getNamespaceSupport**
> Support getNamespaceSupport()

Get a namespace support identifier.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

# **getSamlAuthUrl**
> GetSamlAuthUrl200Response getSamlAuthUrl()

Retrieves the Identity Provider (IdP) URL for authentication within ShellHub. After successful authentication, users are automatically redirected to the ShellHub dashboard.  To access this endpoint, SAML authentication must be enabled and the instance must be running the Enterprise edition. If not, the endpoint returns a `501 Not Implemented` status code. 

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

# **getSessionRecord**
> Array<RecordedSessionResponseInner> getSessionRecord()

Get a session record based on its seat.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

# **getSubscription**
> GetSubscription200Response getSubscription()

Get the subscription.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

const { status, data } = await apiInstance.getSubscription();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**GetSubscription200Response**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a subscription. |  -  |
|**402** | Error |  -  |
|**403** | Error |  -  |
|**404** | Error |  -  |
|**424** | Error |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getValidateAccount**
> getValidateAccount()

Validate the activation link for user.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

# **listAnnouncementsAdmin**
> Array<AnnouncementShort> listAnnouncementsAdmin()

List the announcements posted by ShellHub Cloud.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)
let orderBy: ListAnnouncementsOrderByParameter; // (optional) (default to undefined)

const { status, data } = await apiInstance.listAnnouncementsAdmin(
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

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

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

# **listTunnels**
> Array<Tunnel> listTunnels()

List the tunnels per devices.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

# **listWebEndpoints**
> Array<Webendpoint> listWebEndpoints()

List all web-endpoints in the namespace.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let filter: string; //Web endpoint\'s filter  Filter field receives a base64 enconded JSON object for limit a search.  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)
let sortBy: ListWebEndpointsSortByParameter; //Field to sort by (optional) (default to undefined)
let orderBy: ListWebEndpointsOrderByParameter; //Sort order (asc or desc) (optional) (default to undefined)

const { status, data } = await apiInstance.listWebEndpoints(
    filter,
    page,
    perPage,
    sortBy,
    orderBy
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **filter** | [**string**] | Web endpoint\&#39;s filter  Filter field receives a base64 enconded JSON object for limit a search.  | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|
| **sortBy** | **ListWebEndpointsSortByParameter** | Field to sort by | (optional) defaults to undefined|
| **orderBy** | **ListWebEndpointsOrderByParameter** | Sort order (asc or desc) | (optional) defaults to undefined|


### Return type

**Array<Webendpoint>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get the web-endpoints. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**400** | Bad request |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **mfaRecover**
> UserAuth mfaRecover()

Recover account access by providing one of the user\'s recovery codes. It will be invalidated for future uses.  The recovery code will be cached for 10 minutes. During this period, the user can use the same recovery code to disable their MFA without needing to provide two separate codes. The `X-Expires-At` header specifies the epoch value marking the end of the cache period. 

### Example

```typescript
import {
    CloudApi,
    Configuration,
    MfaRecover
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

# **recordSession**
> recordSession()

Record data about session session.

### Example

```typescript
import {
    CloudApi,
    Configuration,
    RecordSessionRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

# **recoverPassword**
> recoverPassword()

Send a recovery email to the user.

### Example

```typescript
import {
    CloudApi,
    Configuration,
    RecoverPasswordRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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
    CloudApi,
    Configuration,
    RegisterUserRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

# **report**
> report()

Report an action.

### Example

```typescript
import {
    CloudApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let action: ReportActionParameter; // (default to undefined)

const { status, data } = await apiInstance.report(
    action
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **action** | **ReportActionParameter** |  | defaults to undefined|


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
|**200** | Success to report a action. |  -  |
|**400** | Bad Request. |  -  |
|**401** | Unauthorized. |  -  |
|**402** | Payment Required. |  -  |
|**403** | Forbidden. |  -  |
|**404** | Not found. |  -  |
|**424** | Failed dependency. |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **requestResetMFA**
> requestResetMFA()

Sends an email to both the user\'s main and recovery addresses. Each email contains a unique code, which remains valid for at most 1 day. The user must provide both codes to reset their MFA. 

### Example

```typescript
import {
    CloudApi,
    Configuration,
    RequestResetMFARequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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
    CloudApi,
    Configuration,
    ResendEmailRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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
    CloudApi,
    Configuration,
    MfaReset
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

# **setDefaultPaymentMethod**
> setDefaultPaymentMethod()

Set default payment method to the customer.

### Example

```typescript
import {
    CloudApi,
    Configuration,
    AttachPaymentMethodRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let attachPaymentMethodRequest: AttachPaymentMethodRequest; // (optional)

const { status, data } = await apiInstance.setDefaultPaymentMethod(
    attachPaymentMethodRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **attachPaymentMethodRequest** | **AttachPaymentMethodRequest**|  | |


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
|**200** | Success to set default payment method. |  -  |
|**400** | Error |  -  |
|**401** | Error |  -  |
|**403** | Error |  -  |
|**404** | Error |  -  |
|**424** | Error |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateAnnouncement**
> Announcement updateAnnouncement()

Update an announcement.

### Example

```typescript
import {
    CloudApi,
    Configuration,
    CreateAnnouncementRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let uuid: string; // (default to undefined)
let createAnnouncementRequest: CreateAnnouncementRequest; // (optional)

const { status, data } = await apiInstance.updateAnnouncement(
    uuid,
    createAnnouncementRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createAnnouncementRequest** | **CreateAnnouncementRequest**|  | |
| **uuid** | [**string**] |  | defaults to undefined|


### Return type

**Announcement**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to update an announcement. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateFirewallRule**
> FirewallRulesResponse updateFirewallRule()

Update a firewall rule.

### Example

```typescript
import {
    CloudApi,
    Configuration,
    FirewallRulesRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

let id: string; // (default to undefined)
let firewallRulesRequest: FirewallRulesRequest; // (optional)

const { status, data } = await apiInstance.updateFirewallRule(
    id,
    firewallRulesRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **firewallRulesRequest** | **FirewallRulesRequest**|  | |
| **id** | [**string**] |  | defaults to undefined|


### Return type

**FirewallRulesResponse**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to update firewall rule. |  -  |
|**400** | Bad request |  -  |
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
    CloudApi,
    Configuration,
    UpdateRecoverPasswordRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new CloudApi(configuration);

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

