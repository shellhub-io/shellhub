<template>
  <div>
    <Publickey data-test="publickey-component" />
  </div>
</template>

<script>

import Publickey from '@/components/public_key/PublicKey';

export default {
  name: 'Publickeys',

  components: {
    Publickey,
  },
};

</script>
