<template>
  <div>
    <h1>Sessions</h1>
  </div>
  <v-card class="mt-2">
    <SessionList />
  </v-card>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import SessionList from "../components/Sessions/SessionList.vue";

export default defineComponent({ components: { SessionList } });
</script>
