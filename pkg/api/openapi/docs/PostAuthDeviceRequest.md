# PostAuthDeviceRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Info** | [**DeviceInfo**](DeviceInfo.md) |  | 
**Sessions** | Pointer to **[]string** |  | [optional] 
**Hostname** | **string** |  | 
**Identity** | Pointer to [**DeviceIdentity**](DeviceIdentity.md) |  | [optional] 
**PublicKey** | **string** | Device&#39;s public key. | 
**TenantId** | **string** | Namespace&#39;s tenant ID | 

## Methods

### NewPostAuthDeviceRequest

`func NewPostAuthDeviceRequest(info DeviceInfo, hostname string, publicKey string, tenantId string, ) *PostAuthDeviceRequest`

NewPostAuthDeviceRequest instantiates a new PostAuthDeviceRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostAuthDeviceRequestWithDefaults

`func NewPostAuthDeviceRequestWithDefaults() *PostAuthDeviceRequest`

NewPostAuthDeviceRequestWithDefaults instantiates a new PostAuthDeviceRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInfo

`func (o *PostAuthDeviceRequest) GetInfo() DeviceInfo`

GetInfo returns the Info field if non-nil, zero value otherwise.

### GetInfoOk

`func (o *PostAuthDeviceRequest) GetInfoOk() (*DeviceInfo, bool)`

GetInfoOk returns a tuple with the Info field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInfo

`func (o *PostAuthDeviceRequest) SetInfo(v DeviceInfo)`

SetInfo sets Info field to given value.


### GetSessions

`func (o *PostAuthDeviceRequest) GetSessions() []string`

GetSessions returns the Sessions field if non-nil, zero value otherwise.

### GetSessionsOk

`func (o *PostAuthDeviceRequest) GetSessionsOk() (*[]string, bool)`

GetSessionsOk returns a tuple with the Sessions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSessions

`func (o *PostAuthDeviceRequest) SetSessions(v []string)`

SetSessions sets Sessions field to given value.

### HasSessions

`func (o *PostAuthDeviceRequest) HasSessions() bool`

HasSessions returns a boolean if a field has been set.

### GetHostname

`func (o *PostAuthDeviceRequest) GetHostname() string`

GetHostname returns the Hostname field if non-nil, zero value otherwise.

### GetHostnameOk

`func (o *PostAuthDeviceRequest) GetHostnameOk() (*string, bool)`

GetHostnameOk returns a tuple with the Hostname field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostname

`func (o *PostAuthDeviceRequest) SetHostname(v string)`

SetHostname sets Hostname field to given value.


### GetIdentity

`func (o *PostAuthDeviceRequest) GetIdentity() DeviceIdentity`

GetIdentity returns the Identity field if non-nil, zero value otherwise.

### GetIdentityOk

`func (o *PostAuthDeviceRequest) GetIdentityOk() (*DeviceIdentity, bool)`

GetIdentityOk returns a tuple with the Identity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdentity

`func (o *PostAuthDeviceRequest) SetIdentity(v DeviceIdentity)`

SetIdentity sets Identity field to given value.

### HasIdentity

`func (o *PostAuthDeviceRequest) HasIdentity() bool`

HasIdentity returns a boolean if a field has been set.

### GetPublicKey

`func (o *PostAuthDeviceRequest) GetPublicKey() string`

GetPublicKey returns the PublicKey field if non-nil, zero value otherwise.

### GetPublicKeyOk

`func (o *PostAuthDeviceRequest) GetPublicKeyOk() (*string, bool)`

GetPublicKeyOk returns a tuple with the PublicKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicKey

`func (o *PostAuthDeviceRequest) SetPublicKey(v string)`

SetPublicKey sets PublicKey field to given value.


### GetTenantId

`func (o *PostAuthDeviceRequest) GetTenantId() string`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *PostAuthDeviceRequest) GetTenantIdOk() (*string, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *PostAuthDeviceRequest) SetTenantId(v string)`

SetTenantId sets TenantId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


