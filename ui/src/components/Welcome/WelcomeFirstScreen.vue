<template>
  <p class="ml-4 pt-4 text-subtitle-2" data-test="welcome-first-screen-name">
    Hello, <strong> {{ name }} </strong> !
  </p>
  <p class="mt-4 ml-4 mr-4 text-subtitle-2" data-test="welcome-first-screen-text">
    ShellHub is a modern SSH server for remotely accessing Linux devices via CLI
    or web-based user interface, it avoids the inconveniences such as firewall
    and NAT when getting your IP address, automating the access process, and
    helping you easily, seamlessly and securely connect your device to the
    internet.
  </p>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useStore } from "../../store";

const store = useStore();

const name = computed(
  () => store.getters["auth/currentName"] || store.getters["auth/currentUser"],
);

defineExpose({ name });
</script>
