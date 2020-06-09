<template>
  <div>
    <ListFirewall />
  </div>
</template>

<script>

import ListFirewall from '@/components/firewall_rules/FirewallRulesList';

export default {
  name: 'Firewalls',

  components: {
    ListFirewall,
  },
};

</script>
