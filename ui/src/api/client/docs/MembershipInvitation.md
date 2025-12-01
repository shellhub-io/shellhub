# MembershipInvitation

A membership invitation to a namespace

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**namespace** | [**MembershipInvitationNamespace**](MembershipInvitationNamespace.md) |  | [optional] [default to undefined]
**user** | [**MembershipInvitationUser**](MembershipInvitationUser.md) |  | [optional] [default to undefined]
**invited_by** | **string** | The ID of the user who sent the invitation | [optional] [default to undefined]
**created_at** | **string** | When the invitation was created | [optional] [default to undefined]
**updated_at** | **string** | When the invitation was last updated | [optional] [default to undefined]
**expires_at** | **string** | When the invitation expires | [optional] [default to undefined]
**status** | [**MembershipInvitationStatus**](MembershipInvitationStatus.md) |  | [optional] [default to undefined]
**status_updated_at** | **string** | When the status was last updated | [optional] [default to undefined]
**role** | [**NamespaceMemberRole**](NamespaceMemberRole.md) |  | [optional] [default to undefined]

## Example

```typescript
import { MembershipInvitation } from './api';

const instance: MembershipInvitation = {
    namespace,
    user,
    invited_by,
    created_at,
    updated_at,
    expires_at,
    status,
    status_updated_at,
    role,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
