<template>
  <div>
    <ListPublickey />
  </div>
</template>

<script>

import ListPublickey from '@/components/public_key/PublicKeyList';

export default {
  name: 'Publickeys',

  components: {
    ListPublickey,
  },
};

</script>
