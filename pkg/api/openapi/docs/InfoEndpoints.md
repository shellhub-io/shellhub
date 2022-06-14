# InfoEndpoints

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Ssh** | Pointer to **string** | SSH endpoint. | [optional] 
**Api** | Pointer to **string** | API endpoint. | [optional] 

## Methods

### NewInfoEndpoints

`func NewInfoEndpoints() *InfoEndpoints`

NewInfoEndpoints instantiates a new InfoEndpoints object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInfoEndpointsWithDefaults

`func NewInfoEndpointsWithDefaults() *InfoEndpoints`

NewInfoEndpointsWithDefaults instantiates a new InfoEndpoints object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSsh

`func (o *InfoEndpoints) GetSsh() string`

GetSsh returns the Ssh field if non-nil, zero value otherwise.

### GetSshOk

`func (o *InfoEndpoints) GetSshOk() (*string, bool)`

GetSshOk returns a tuple with the Ssh field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSsh

`func (o *InfoEndpoints) SetSsh(v string)`

SetSsh sets Ssh field to given value.

### HasSsh

`func (o *InfoEndpoints) HasSsh() bool`

HasSsh returns a boolean if a field has been set.

### GetApi

`func (o *InfoEndpoints) GetApi() string`

GetApi returns the Api field if non-nil, zero value otherwise.

### GetApiOk

`func (o *InfoEndpoints) GetApiOk() (*string, bool)`

GetApiOk returns a tuple with the Api field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApi

`func (o *InfoEndpoints) SetApi(v string)`

SetApi sets Api field to given value.

### HasApi

`func (o *InfoEndpoints) HasApi() bool`

HasApi returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


