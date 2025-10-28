# FirewallRulesResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** | Firewall rule\&#39;s ID. | [optional] [default to undefined]
**tenant_id** | **string** | Namespace\&#39;s tenant ID | [default to undefined]
**action** | [**FirewallRulesResponseAction**](FirewallRulesResponseAction.md) |  | [default to undefined]
**active** | **boolean** | Firewall rule active\&#39;s status | [default to undefined]
**filter** | [**FirewallRulesResponseFilter**](FirewallRulesResponseFilter.md) |  | [default to undefined]
**priority** | **number** | Firewall rule\&#39;s priority | [default to undefined]
**source_ip** | **string** | Firewall rule\&#39;s source IP regexp | [default to undefined]
**username** | **string** | Firewall rule\&#39;s username regexp | [default to undefined]

## Example

```typescript
import { FirewallRulesResponse } from './api';

const instance: FirewallRulesResponse = {
    id,
    tenant_id,
    action,
    active,
    filter,
    priority,
    source_ip,
    username,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
