<template>
  <div>
    <Session data-test="session-component" />
  </div>
</template>

<script>

import Session from '@/components/session/Session';

export default {
  name: 'SessionsView',

  components: {
    Session,
  },
};

</script>
