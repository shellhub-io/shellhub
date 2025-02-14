<template>
  <h1>License Details</h1>
  <v-divider class="mt-2 mb-4" />
  <LicenseComponent />
</template>

<script lang="ts">
import { defineComponent } from "vue";
import LicenseComponent from "../components/License/License.vue";

export default defineComponent({
  components: { LicenseComponent },
});
</script>
