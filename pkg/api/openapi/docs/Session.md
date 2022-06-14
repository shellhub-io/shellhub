# Session

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Uid** | Pointer to **string** | Session&#39;s UID | [optional] 
**DeviceUid** | Pointer to **string** | Device&#39;s UID | [optional] 
**Device** | Pointer to [**Device**](Device.md) |  | [optional] 
**TenantId** | Pointer to **string** | Namespace&#39;s tenant ID | [optional] 
**Username** | Pointer to **string** | Session&#39;s username | [optional] 
**IpAddress** | Pointer to **string** | Session&#39;s IP address | [optional] 
**StartedAt** | Pointer to **string** | Session&#39;s started date | [optional] 
**LastSeen** | Pointer to **string** | Session&#39;s last seen date | [optional] 
**Active** | Pointer to **bool** | Session&#39;s active status | [optional] 
**Authenticated** | Pointer to **bool** | Session&#39;s authenticated status | [optional] 
**Recorded** | Pointer to **bool** | Session&#39;s recorded status | [optional] 
**Type** | Pointer to **string** | Session&#39;s type | [optional] 
**Term** | Pointer to **string** | Session&#39;s terminal | [optional] 

## Methods

### NewSession

`func NewSession() *Session`

NewSession instantiates a new Session object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSessionWithDefaults

`func NewSessionWithDefaults() *Session`

NewSessionWithDefaults instantiates a new Session object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUid

`func (o *Session) GetUid() string`

GetUid returns the Uid field if non-nil, zero value otherwise.

### GetUidOk

`func (o *Session) GetUidOk() (*string, bool)`

GetUidOk returns a tuple with the Uid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUid

`func (o *Session) SetUid(v string)`

SetUid sets Uid field to given value.

### HasUid

`func (o *Session) HasUid() bool`

HasUid returns a boolean if a field has been set.

### GetDeviceUid

`func (o *Session) GetDeviceUid() string`

GetDeviceUid returns the DeviceUid field if non-nil, zero value otherwise.

### GetDeviceUidOk

`func (o *Session) GetDeviceUidOk() (*string, bool)`

GetDeviceUidOk returns a tuple with the DeviceUid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceUid

`func (o *Session) SetDeviceUid(v string)`

SetDeviceUid sets DeviceUid field to given value.

### HasDeviceUid

`func (o *Session) HasDeviceUid() bool`

HasDeviceUid returns a boolean if a field has been set.

### GetDevice

`func (o *Session) GetDevice() Device`

GetDevice returns the Device field if non-nil, zero value otherwise.

### GetDeviceOk

`func (o *Session) GetDeviceOk() (*Device, bool)`

GetDeviceOk returns a tuple with the Device field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDevice

`func (o *Session) SetDevice(v Device)`

SetDevice sets Device field to given value.

### HasDevice

`func (o *Session) HasDevice() bool`

HasDevice returns a boolean if a field has been set.

### GetTenantId

`func (o *Session) GetTenantId() string`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *Session) GetTenantIdOk() (*string, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *Session) SetTenantId(v string)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *Session) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetUsername

`func (o *Session) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *Session) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *Session) SetUsername(v string)`

SetUsername sets Username field to given value.

### HasUsername

`func (o *Session) HasUsername() bool`

HasUsername returns a boolean if a field has been set.

### GetIpAddress

`func (o *Session) GetIpAddress() string`

GetIpAddress returns the IpAddress field if non-nil, zero value otherwise.

### GetIpAddressOk

`func (o *Session) GetIpAddressOk() (*string, bool)`

GetIpAddressOk returns a tuple with the IpAddress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIpAddress

`func (o *Session) SetIpAddress(v string)`

SetIpAddress sets IpAddress field to given value.

### HasIpAddress

`func (o *Session) HasIpAddress() bool`

HasIpAddress returns a boolean if a field has been set.

### GetStartedAt

`func (o *Session) GetStartedAt() string`

GetStartedAt returns the StartedAt field if non-nil, zero value otherwise.

### GetStartedAtOk

`func (o *Session) GetStartedAtOk() (*string, bool)`

GetStartedAtOk returns a tuple with the StartedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStartedAt

`func (o *Session) SetStartedAt(v string)`

SetStartedAt sets StartedAt field to given value.

### HasStartedAt

`func (o *Session) HasStartedAt() bool`

HasStartedAt returns a boolean if a field has been set.

### GetLastSeen

`func (o *Session) GetLastSeen() string`

GetLastSeen returns the LastSeen field if non-nil, zero value otherwise.

### GetLastSeenOk

`func (o *Session) GetLastSeenOk() (*string, bool)`

GetLastSeenOk returns a tuple with the LastSeen field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastSeen

`func (o *Session) SetLastSeen(v string)`

SetLastSeen sets LastSeen field to given value.

### HasLastSeen

`func (o *Session) HasLastSeen() bool`

HasLastSeen returns a boolean if a field has been set.

### GetActive

`func (o *Session) GetActive() bool`

GetActive returns the Active field if non-nil, zero value otherwise.

### GetActiveOk

`func (o *Session) GetActiveOk() (*bool, bool)`

GetActiveOk returns a tuple with the Active field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActive

`func (o *Session) SetActive(v bool)`

SetActive sets Active field to given value.

### HasActive

`func (o *Session) HasActive() bool`

HasActive returns a boolean if a field has been set.

### GetAuthenticated

`func (o *Session) GetAuthenticated() bool`

GetAuthenticated returns the Authenticated field if non-nil, zero value otherwise.

### GetAuthenticatedOk

`func (o *Session) GetAuthenticatedOk() (*bool, bool)`

GetAuthenticatedOk returns a tuple with the Authenticated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthenticated

`func (o *Session) SetAuthenticated(v bool)`

SetAuthenticated sets Authenticated field to given value.

### HasAuthenticated

`func (o *Session) HasAuthenticated() bool`

HasAuthenticated returns a boolean if a field has been set.

### GetRecorded

`func (o *Session) GetRecorded() bool`

GetRecorded returns the Recorded field if non-nil, zero value otherwise.

### GetRecordedOk

`func (o *Session) GetRecordedOk() (*bool, bool)`

GetRecordedOk returns a tuple with the Recorded field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecorded

`func (o *Session) SetRecorded(v bool)`

SetRecorded sets Recorded field to given value.

### HasRecorded

`func (o *Session) HasRecorded() bool`

HasRecorded returns a boolean if a field has been set.

### GetType

`func (o *Session) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *Session) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *Session) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *Session) HasType() bool`

HasType returns a boolean if a field has been set.

### GetTerm

`func (o *Session) GetTerm() string`

GetTerm returns the Term field if non-nil, zero value otherwise.

### GetTermOk

`func (o *Session) GetTermOk() (*string, bool)`

GetTermOk returns a tuple with the Term field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerm

`func (o *Session) SetTerm(v string)`

SetTerm sets Term field to given value.

### HasTerm

`func (o *Session) HasTerm() bool`

HasTerm returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


