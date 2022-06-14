# PostAuthDevice200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Uid** | Pointer to **string** | Device&#39;s UID | [optional] 
**Token** | Pointer to **string** | Device&#39;s token | [optional] 
**Name** | Pointer to **string** | Device&#39;s name   By default, the name is the device&#39;s MAC address when it just added.  | [optional] 
**Namespace** | Pointer to **string** | Device&#39;s namespace name | [optional] 

## Methods

### NewPostAuthDevice200Response

`func NewPostAuthDevice200Response() *PostAuthDevice200Response`

NewPostAuthDevice200Response instantiates a new PostAuthDevice200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostAuthDevice200ResponseWithDefaults

`func NewPostAuthDevice200ResponseWithDefaults() *PostAuthDevice200Response`

NewPostAuthDevice200ResponseWithDefaults instantiates a new PostAuthDevice200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUid

`func (o *PostAuthDevice200Response) GetUid() string`

GetUid returns the Uid field if non-nil, zero value otherwise.

### GetUidOk

`func (o *PostAuthDevice200Response) GetUidOk() (*string, bool)`

GetUidOk returns a tuple with the Uid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUid

`func (o *PostAuthDevice200Response) SetUid(v string)`

SetUid sets Uid field to given value.

### HasUid

`func (o *PostAuthDevice200Response) HasUid() bool`

HasUid returns a boolean if a field has been set.

### GetToken

`func (o *PostAuthDevice200Response) GetToken() string`

GetToken returns the Token field if non-nil, zero value otherwise.

### GetTokenOk

`func (o *PostAuthDevice200Response) GetTokenOk() (*string, bool)`

GetTokenOk returns a tuple with the Token field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToken

`func (o *PostAuthDevice200Response) SetToken(v string)`

SetToken sets Token field to given value.

### HasToken

`func (o *PostAuthDevice200Response) HasToken() bool`

HasToken returns a boolean if a field has been set.

### GetName

`func (o *PostAuthDevice200Response) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PostAuthDevice200Response) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PostAuthDevice200Response) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *PostAuthDevice200Response) HasName() bool`

HasName returns a boolean if a field has been set.

### GetNamespace

`func (o *PostAuthDevice200Response) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *PostAuthDevice200Response) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *PostAuthDevice200Response) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *PostAuthDevice200Response) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


