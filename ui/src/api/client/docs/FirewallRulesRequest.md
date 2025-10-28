# FirewallRulesRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**action** | [**FirewallRulesResponseAction**](FirewallRulesResponseAction.md) |  | [default to undefined]
**active** | **boolean** | Firewall rule active\&#39;s status | [default to undefined]
**filter** | [**FirewallRulesResponseFilter**](FirewallRulesResponseFilter.md) |  | [default to undefined]
**priority** | **number** | Firewall rule\&#39;s priority | [default to undefined]
**source_ip** | **string** | Firewall rule\&#39;s source IP regexp | [default to undefined]
**username** | **string** | Firewall rule\&#39;s username regexp | [default to undefined]

## Example

```typescript
import { FirewallRulesRequest } from './api';

const instance: FirewallRulesRequest = {
    action,
    active,
    filter,
    priority,
    source_ip,
    username,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
