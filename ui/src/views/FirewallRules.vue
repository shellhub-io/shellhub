<template>
  <div>
    <Firewall />
  </div>
</template>

<script>

import Firewall from '@/components/firewall_rules/FirewallRules';

export default {
  name: 'Firewalls',

  components: {
    Firewall,
  },
};

</script>
