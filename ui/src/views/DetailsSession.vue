<template>
  <div>
    <SessionDetails data-test="sessionDetails-component" />
  </div>
</template>

<script>

import SessionDetails from '@/components/session/SessionDetails';

export default {
  name: 'DetailsSessionView',

  components: {
    SessionDetails,
  },
};

</script>
