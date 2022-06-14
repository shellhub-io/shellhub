# Device

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Uid** | Pointer to **string** | Device&#39;s UID | [optional] 
**Name** | Pointer to **string** | Device&#39;s name   By default, the name is the device&#39;s MAC address when it just added.  | [optional] 
**Identity** | Pointer to [**DeviceIdentity**](DeviceIdentity.md) |  | [optional] 
**Info** | Pointer to [**DeviceInfo**](DeviceInfo.md) |  | [optional] 
**PublicKey** | Pointer to **string** | Device&#39;s public key. | [optional] 
**TenantId** | Pointer to **string** | Namespace&#39;s tenant ID | [optional] 
**LastSeen** | Pointer to **time.Time** | Device&#39;s last seen date | [optional] 
**Online** | Pointer to **bool** | Device&#39;s availability status | [optional] 
**Namespace** | Pointer to **string** | Namespace&#39;s name | [optional] 
**Status** | Pointer to [**DeviceStatus**](DeviceStatus.md) |  | [optional] 
**CreatedAt** | Pointer to **time.Time** | Device&#39;s creation date | [optional] 
**RemoteAddr** | Pointer to **string** | Device&#39;s remote address | [optional] 
**Position** | Pointer to [**DevicePosition**](DevicePosition.md) |  | [optional] 
**Tags** | Pointer to **[]string** | Device&#39;s Tags list | [optional] 

## Methods

### NewDevice

`func NewDevice() *Device`

NewDevice instantiates a new Device object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceWithDefaults

`func NewDeviceWithDefaults() *Device`

NewDeviceWithDefaults instantiates a new Device object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUid

`func (o *Device) GetUid() string`

GetUid returns the Uid field if non-nil, zero value otherwise.

### GetUidOk

`func (o *Device) GetUidOk() (*string, bool)`

GetUidOk returns a tuple with the Uid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUid

`func (o *Device) SetUid(v string)`

SetUid sets Uid field to given value.

### HasUid

`func (o *Device) HasUid() bool`

HasUid returns a boolean if a field has been set.

### GetName

`func (o *Device) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Device) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Device) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *Device) HasName() bool`

HasName returns a boolean if a field has been set.

### GetIdentity

`func (o *Device) GetIdentity() DeviceIdentity`

GetIdentity returns the Identity field if non-nil, zero value otherwise.

### GetIdentityOk

`func (o *Device) GetIdentityOk() (*DeviceIdentity, bool)`

GetIdentityOk returns a tuple with the Identity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdentity

`func (o *Device) SetIdentity(v DeviceIdentity)`

SetIdentity sets Identity field to given value.

### HasIdentity

`func (o *Device) HasIdentity() bool`

HasIdentity returns a boolean if a field has been set.

### GetInfo

`func (o *Device) GetInfo() DeviceInfo`

GetInfo returns the Info field if non-nil, zero value otherwise.

### GetInfoOk

`func (o *Device) GetInfoOk() (*DeviceInfo, bool)`

GetInfoOk returns a tuple with the Info field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInfo

`func (o *Device) SetInfo(v DeviceInfo)`

SetInfo sets Info field to given value.

### HasInfo

`func (o *Device) HasInfo() bool`

HasInfo returns a boolean if a field has been set.

### GetPublicKey

`func (o *Device) GetPublicKey() string`

GetPublicKey returns the PublicKey field if non-nil, zero value otherwise.

### GetPublicKeyOk

`func (o *Device) GetPublicKeyOk() (*string, bool)`

GetPublicKeyOk returns a tuple with the PublicKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicKey

`func (o *Device) SetPublicKey(v string)`

SetPublicKey sets PublicKey field to given value.

### HasPublicKey

`func (o *Device) HasPublicKey() bool`

HasPublicKey returns a boolean if a field has been set.

### GetTenantId

`func (o *Device) GetTenantId() string`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *Device) GetTenantIdOk() (*string, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *Device) SetTenantId(v string)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *Device) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetLastSeen

`func (o *Device) GetLastSeen() time.Time`

GetLastSeen returns the LastSeen field if non-nil, zero value otherwise.

### GetLastSeenOk

`func (o *Device) GetLastSeenOk() (*time.Time, bool)`

GetLastSeenOk returns a tuple with the LastSeen field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastSeen

`func (o *Device) SetLastSeen(v time.Time)`

SetLastSeen sets LastSeen field to given value.

### HasLastSeen

`func (o *Device) HasLastSeen() bool`

HasLastSeen returns a boolean if a field has been set.

### GetOnline

`func (o *Device) GetOnline() bool`

GetOnline returns the Online field if non-nil, zero value otherwise.

### GetOnlineOk

`func (o *Device) GetOnlineOk() (*bool, bool)`

GetOnlineOk returns a tuple with the Online field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOnline

`func (o *Device) SetOnline(v bool)`

SetOnline sets Online field to given value.

### HasOnline

`func (o *Device) HasOnline() bool`

HasOnline returns a boolean if a field has been set.

### GetNamespace

`func (o *Device) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *Device) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *Device) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *Device) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.

### GetStatus

`func (o *Device) GetStatus() DeviceStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *Device) GetStatusOk() (*DeviceStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *Device) SetStatus(v DeviceStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *Device) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetCreatedAt

`func (o *Device) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Device) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Device) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *Device) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetRemoteAddr

`func (o *Device) GetRemoteAddr() string`

GetRemoteAddr returns the RemoteAddr field if non-nil, zero value otherwise.

### GetRemoteAddrOk

`func (o *Device) GetRemoteAddrOk() (*string, bool)`

GetRemoteAddrOk returns a tuple with the RemoteAddr field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRemoteAddr

`func (o *Device) SetRemoteAddr(v string)`

SetRemoteAddr sets RemoteAddr field to given value.

### HasRemoteAddr

`func (o *Device) HasRemoteAddr() bool`

HasRemoteAddr returns a boolean if a field has been set.

### GetPosition

`func (o *Device) GetPosition() DevicePosition`

GetPosition returns the Position field if non-nil, zero value otherwise.

### GetPositionOk

`func (o *Device) GetPositionOk() (*DevicePosition, bool)`

GetPositionOk returns a tuple with the Position field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPosition

`func (o *Device) SetPosition(v DevicePosition)`

SetPosition sets Position field to given value.

### HasPosition

`func (o *Device) HasPosition() bool`

HasPosition returns a boolean if a field has been set.

### GetTags

`func (o *Device) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *Device) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *Device) SetTags(v []string)`

SetTags sets Tags field to given value.

### HasTags

`func (o *Device) HasTags() bool`

HasTags returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


