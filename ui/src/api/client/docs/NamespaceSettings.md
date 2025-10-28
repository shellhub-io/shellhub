# NamespaceSettings

Namespace\'s settings.

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**session_record** | **boolean** | The session records define when the namespace should save or not record a session. This can be used to check logged activity when connecting to a device. | [optional] [default to undefined]
**connection_announcement** | **string** | A connection announcement is a custom string written during a session when a connection is established on a device within the namespace. | [optional] [default to undefined]

## Example

```typescript
import { NamespaceSettings } from './api';

const instance: NamespaceSettings = {
    session_record,
    connection_announcement,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
