# PublicKeyRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Data** | **string** | Public key&#39;s data.  The &#x60;data&#x60; field receives the public key enconded as &#x60;base64&#x60; string.  | 
**Filter** | [**PublicKeyFilter**](PublicKeyFilter.md) |  | 
**Name** | **string** | Public key&#39;s name. | 
**Username** | **string** | Public key&#39;s regex username.   The &#x60;username&#x60; field define which user, in the device, may be access through this public key.  | 

## Methods

### NewPublicKeyRequest

`func NewPublicKeyRequest(data string, filter PublicKeyFilter, name string, username string, ) *PublicKeyRequest`

NewPublicKeyRequest instantiates a new PublicKeyRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPublicKeyRequestWithDefaults

`func NewPublicKeyRequestWithDefaults() *PublicKeyRequest`

NewPublicKeyRequestWithDefaults instantiates a new PublicKeyRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetData

`func (o *PublicKeyRequest) GetData() string`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *PublicKeyRequest) GetDataOk() (*string, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *PublicKeyRequest) SetData(v string)`

SetData sets Data field to given value.


### GetFilter

`func (o *PublicKeyRequest) GetFilter() PublicKeyFilter`

GetFilter returns the Filter field if non-nil, zero value otherwise.

### GetFilterOk

`func (o *PublicKeyRequest) GetFilterOk() (*PublicKeyFilter, bool)`

GetFilterOk returns a tuple with the Filter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilter

`func (o *PublicKeyRequest) SetFilter(v PublicKeyFilter)`

SetFilter sets Filter field to given value.


### GetName

`func (o *PublicKeyRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PublicKeyRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PublicKeyRequest) SetName(v string)`

SetName sets Name field to given value.


### GetUsername

`func (o *PublicKeyRequest) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *PublicKeyRequest) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *PublicKeyRequest) SetUsername(v string)`

SetUsername sets Username field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


