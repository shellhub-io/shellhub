<template>
  <div>
    <DeviceDetails data-test="deviceDetails-component" />
  </div>
</template>

<script>

import DeviceDetails from '@/components/device/DeviceDetails';

export default {
  name: 'DetailsDeviceView',

  components: {
    DeviceDetails,
  },
};

</script>
