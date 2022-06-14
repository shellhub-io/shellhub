# DeviceInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | Device&#39;s OS name | [optional] 
**PrettyName** | Pointer to **string** | Device&#39;s OS pretty name | [optional] 
**Version** | Pointer to **string** | Device&#39;s OS version | [optional] 
**Arch** | Pointer to **string** | Device&#39;s OS arch | [optional] 
**Platform** | Pointer to **string** | Device&#39;s OS platform | [optional] 

## Methods

### NewDeviceInfo

`func NewDeviceInfo() *DeviceInfo`

NewDeviceInfo instantiates a new DeviceInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceInfoWithDefaults

`func NewDeviceInfoWithDefaults() *DeviceInfo`

NewDeviceInfoWithDefaults instantiates a new DeviceInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DeviceInfo) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DeviceInfo) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DeviceInfo) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *DeviceInfo) HasId() bool`

HasId returns a boolean if a field has been set.

### GetPrettyName

`func (o *DeviceInfo) GetPrettyName() string`

GetPrettyName returns the PrettyName field if non-nil, zero value otherwise.

### GetPrettyNameOk

`func (o *DeviceInfo) GetPrettyNameOk() (*string, bool)`

GetPrettyNameOk returns a tuple with the PrettyName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrettyName

`func (o *DeviceInfo) SetPrettyName(v string)`

SetPrettyName sets PrettyName field to given value.

### HasPrettyName

`func (o *DeviceInfo) HasPrettyName() bool`

HasPrettyName returns a boolean if a field has been set.

### GetVersion

`func (o *DeviceInfo) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *DeviceInfo) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *DeviceInfo) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *DeviceInfo) HasVersion() bool`

HasVersion returns a boolean if a field has been set.

### GetArch

`func (o *DeviceInfo) GetArch() string`

GetArch returns the Arch field if non-nil, zero value otherwise.

### GetArchOk

`func (o *DeviceInfo) GetArchOk() (*string, bool)`

GetArchOk returns a tuple with the Arch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArch

`func (o *DeviceInfo) SetArch(v string)`

SetArch sets Arch field to given value.

### HasArch

`func (o *DeviceInfo) HasArch() bool`

HasArch returns a boolean if a field has been set.

### GetPlatform

`func (o *DeviceInfo) GetPlatform() string`

GetPlatform returns the Platform field if non-nil, zero value otherwise.

### GetPlatformOk

`func (o *DeviceInfo) GetPlatformOk() (*string, bool)`

GetPlatformOk returns a tuple with the Platform field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlatform

`func (o *DeviceInfo) SetPlatform(v string)`

SetPlatform sets Platform field to given value.

### HasPlatform

`func (o *DeviceInfo) HasPlatform() bool`

HasPlatform returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


