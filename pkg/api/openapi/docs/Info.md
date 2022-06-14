# Info

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Version** | Pointer to **string** | ShellHub instance version. | [optional] 
**Endpoints** | Pointer to [**InfoEndpoints**](InfoEndpoints.md) |  | [optional] 

## Methods

### NewInfo

`func NewInfo() *Info`

NewInfo instantiates a new Info object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInfoWithDefaults

`func NewInfoWithDefaults() *Info`

NewInfoWithDefaults instantiates a new Info object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetVersion

`func (o *Info) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *Info) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *Info) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *Info) HasVersion() bool`

HasVersion returns a boolean if a field has been set.

### GetEndpoints

`func (o *Info) GetEndpoints() InfoEndpoints`

GetEndpoints returns the Endpoints field if non-nil, zero value otherwise.

### GetEndpointsOk

`func (o *Info) GetEndpointsOk() (*InfoEndpoints, bool)`

GetEndpointsOk returns a tuple with the Endpoints field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEndpoints

`func (o *Info) SetEndpoints(v InfoEndpoints)`

SetEndpoints sets Endpoints field to given value.

### HasEndpoints

`func (o *Info) HasEndpoints() bool`

HasEndpoints returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


