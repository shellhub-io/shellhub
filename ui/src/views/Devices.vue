<template>
  <div>
    <DeviceList data-test="deviceList-component" />
  </div>
</template>

<script>

import DeviceList from '@/components/device/Device';

export default {
  name: 'DevicesView',

  components: {
    DeviceList,
  },
};

</script>
