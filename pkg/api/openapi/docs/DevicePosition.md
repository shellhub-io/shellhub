# DevicePosition

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Latitude** | Pointer to **float32** | Device&#39;s latitude position | [optional] 
**Longitude** | Pointer to **float32** | Device&#39;s longitude position | [optional] 

## Methods

### NewDevicePosition

`func NewDevicePosition() *DevicePosition`

NewDevicePosition instantiates a new DevicePosition object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDevicePositionWithDefaults

`func NewDevicePositionWithDefaults() *DevicePosition`

NewDevicePositionWithDefaults instantiates a new DevicePosition object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLatitude

`func (o *DevicePosition) GetLatitude() float32`

GetLatitude returns the Latitude field if non-nil, zero value otherwise.

### GetLatitudeOk

`func (o *DevicePosition) GetLatitudeOk() (*float32, bool)`

GetLatitudeOk returns a tuple with the Latitude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLatitude

`func (o *DevicePosition) SetLatitude(v float32)`

SetLatitude sets Latitude field to given value.

### HasLatitude

`func (o *DevicePosition) HasLatitude() bool`

HasLatitude returns a boolean if a field has been set.

### GetLongitude

`func (o *DevicePosition) GetLongitude() float32`

GetLongitude returns the Longitude field if non-nil, zero value otherwise.

### GetLongitudeOk

`func (o *DevicePosition) GetLongitudeOk() (*float32, bool)`

GetLongitudeOk returns a tuple with the Longitude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLongitude

`func (o *DevicePosition) SetLongitude(v float32)`

SetLongitude sets Longitude field to given value.

### HasLongitude

`func (o *DevicePosition) HasLongitude() bool`

HasLongitude returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


