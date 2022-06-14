# Namespace

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | Namespace&#39;s name | [optional] 
**Owner** | Pointer to **string** | User&#39;s ID. | [optional] 
**TenantId** | Pointer to **string** | Namespace&#39;s tenant ID | [optional] 
**Members** | Pointer to [**[]NamespaceMembersInner**](NamespaceMembersInner.md) | Namespace&#39;s members | [optional] 
**Settings** | Pointer to [**NamespaceSettings**](NamespaceSettings.md) |  | [optional] 
**MaxDevices** | Pointer to **int32** | Namespace&#39;s max device numbers | [optional] [default to 3]
**DeviceCount** | Pointer to **int32** | Namespace&#39;s total devices | [optional] 
**CreatedAt** | Pointer to **time.Time** | Namespace&#39;s creation date | [optional] 
**Billing** | Pointer to **map[string]interface{}** | Namespace&#39;s billing | [optional] 

## Methods

### NewNamespace

`func NewNamespace() *Namespace`

NewNamespace instantiates a new Namespace object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNamespaceWithDefaults

`func NewNamespaceWithDefaults() *Namespace`

NewNamespaceWithDefaults instantiates a new Namespace object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Namespace) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Namespace) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Namespace) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *Namespace) HasName() bool`

HasName returns a boolean if a field has been set.

### GetOwner

`func (o *Namespace) GetOwner() string`

GetOwner returns the Owner field if non-nil, zero value otherwise.

### GetOwnerOk

`func (o *Namespace) GetOwnerOk() (*string, bool)`

GetOwnerOk returns a tuple with the Owner field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOwner

`func (o *Namespace) SetOwner(v string)`

SetOwner sets Owner field to given value.

### HasOwner

`func (o *Namespace) HasOwner() bool`

HasOwner returns a boolean if a field has been set.

### GetTenantId

`func (o *Namespace) GetTenantId() string`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *Namespace) GetTenantIdOk() (*string, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *Namespace) SetTenantId(v string)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *Namespace) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetMembers

`func (o *Namespace) GetMembers() []NamespaceMembersInner`

GetMembers returns the Members field if non-nil, zero value otherwise.

### GetMembersOk

`func (o *Namespace) GetMembersOk() (*[]NamespaceMembersInner, bool)`

GetMembersOk returns a tuple with the Members field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMembers

`func (o *Namespace) SetMembers(v []NamespaceMembersInner)`

SetMembers sets Members field to given value.

### HasMembers

`func (o *Namespace) HasMembers() bool`

HasMembers returns a boolean if a field has been set.

### GetSettings

`func (o *Namespace) GetSettings() NamespaceSettings`

GetSettings returns the Settings field if non-nil, zero value otherwise.

### GetSettingsOk

`func (o *Namespace) GetSettingsOk() (*NamespaceSettings, bool)`

GetSettingsOk returns a tuple with the Settings field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSettings

`func (o *Namespace) SetSettings(v NamespaceSettings)`

SetSettings sets Settings field to given value.

### HasSettings

`func (o *Namespace) HasSettings() bool`

HasSettings returns a boolean if a field has been set.

### GetMaxDevices

`func (o *Namespace) GetMaxDevices() int32`

GetMaxDevices returns the MaxDevices field if non-nil, zero value otherwise.

### GetMaxDevicesOk

`func (o *Namespace) GetMaxDevicesOk() (*int32, bool)`

GetMaxDevicesOk returns a tuple with the MaxDevices field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxDevices

`func (o *Namespace) SetMaxDevices(v int32)`

SetMaxDevices sets MaxDevices field to given value.

### HasMaxDevices

`func (o *Namespace) HasMaxDevices() bool`

HasMaxDevices returns a boolean if a field has been set.

### GetDeviceCount

`func (o *Namespace) GetDeviceCount() int32`

GetDeviceCount returns the DeviceCount field if non-nil, zero value otherwise.

### GetDeviceCountOk

`func (o *Namespace) GetDeviceCountOk() (*int32, bool)`

GetDeviceCountOk returns a tuple with the DeviceCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceCount

`func (o *Namespace) SetDeviceCount(v int32)`

SetDeviceCount sets DeviceCount field to given value.

### HasDeviceCount

`func (o *Namespace) HasDeviceCount() bool`

HasDeviceCount returns a boolean if a field has been set.

### GetCreatedAt

`func (o *Namespace) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Namespace) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Namespace) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *Namespace) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetBilling

`func (o *Namespace) GetBilling() map[string]interface{}`

GetBilling returns the Billing field if non-nil, zero value otherwise.

### GetBillingOk

`func (o *Namespace) GetBillingOk() (*map[string]interface{}, bool)`

GetBillingOk returns a tuple with the Billing field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBilling

`func (o *Namespace) SetBilling(v map[string]interface{})`

SetBilling sets Billing field to given value.

### HasBilling

`func (o *Namespace) HasBilling() bool`

HasBilling returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


