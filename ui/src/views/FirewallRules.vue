<template>
  <div>
    <Firewall />
  </div>
</template>

<script>

import Firewall from '@/components/firewall_rule/FirewallRule';

export default {
  name: 'FirewallRules',

  components: {
    Firewall,
  },
};

</script>
