<template>
  <div>
    <DeviceTable
      variant="device"
      header="secondary"
      status="pending"
      :storeMethods="getDeviceStoreMethods()"
      data-test="device-table"
    />
  </div>
</template>

<script setup lang="ts">
import { getDeviceStoreMethods } from "@/utils/storeMethods";
import DeviceTable from "../Tables/DeviceTable.vue";
</script>
