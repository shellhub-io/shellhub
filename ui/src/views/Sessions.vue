<template>
  <div>
    <ListSession/>
  </div>
</template>

<script>
import ListSession from '@/components/session/SessionList.vue';

export default {
  name: 'Sessions',

  components: {
    ListSession
  }
};
</script>