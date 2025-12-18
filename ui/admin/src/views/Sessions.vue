<template>
  <PageHeader
    icon="mdi-history"
    title="Sessions"
    overline="Activity Monitoring"
    description="Track live and historical sessions happening across every namespace."
    icon-color="primary"
  />
  <SessionList />
</template>

<script setup lang="ts">
import PageHeader from "@/components/PageHeader.vue";
import SessionList from "../components/Sessions/SessionList.vue";
</script>
