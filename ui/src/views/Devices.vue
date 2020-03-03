<template>
  <div>
    <DeviceList/>
  </div>
</template>

<script>
import DeviceList from '@/components/device/DeviceList.vue';

export default {
  name: 'Devices',

  components: {
    DeviceList
  }
};
</script>