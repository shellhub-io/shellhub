# NamespaceMembersInner


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** | User\&#39;s ID. | [optional] [default to undefined]
**added_at** | **string** | The time when the member was invited. | [optional] [default to undefined]
**expires_at** | **string** | **NOTE: ONLY USED IN CLOUD INSTANCE.**  The time when the invite expires. If the member is not in &#x60;pending&#x60; status, this will be set to the zero UTC time.  | [optional] [default to undefined]
**role** | [**NamespaceMemberRole**](NamespaceMemberRole.md) |  | [optional] [default to undefined]
**type** | [**NamespaceMembersInnerType**](NamespaceMembersInnerType.md) |  | [optional] [default to undefined]
**status** | [**NamespaceMembersInnerStatus**](NamespaceMembersInnerStatus.md) |  | [optional] [default to undefined]
**email** | **string** | Member\&#39;s email. | [optional] [default to undefined]

## Example

```typescript
import { NamespaceMembersInner } from './api';

const instance: NamespaceMembersInner = {
    id,
    added_at,
    expires_at,
    role,
    type,
    status,
    email,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
