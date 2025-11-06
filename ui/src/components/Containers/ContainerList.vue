<template>
  <div>
    <DeviceTable
      variant="container"
      header="primary"
      status="accepted"
      :store-methods="getContainerStoreMethods()"
      data-test="container-table"
    />
  </div>
</template>

<script setup lang="ts">
import { getContainerStoreMethods } from "@/utils/storeMethods";
import DeviceTable from "../Tables/DeviceTable.vue";
</script>
