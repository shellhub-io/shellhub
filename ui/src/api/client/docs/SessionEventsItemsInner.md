# SessionEventsItemsInner


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**type** | **string** | The type of the event | [optional] [default to undefined]
**timestamp** | **string** | The time the event occurred in ISO 8601 format | [optional] [default to undefined]
**data** | **object** | Additional data related to the event | [optional] [default to undefined]
**seat** | **number** | Seat where the event happened | [optional] [default to undefined]

## Example

```typescript
import { SessionEventsItemsInner } from './api';

const instance: SessionEventsItemsInner = {
    type,
    timestamp,
    data,
    seat,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
