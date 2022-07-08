<template>
  <p class="ml-4 pt-4 text-subtitle-2">
    Hello, <strong> {{ name }} </strong> !
  </p>
  <p class="mt-4 ml-4 mr-4 text-subtitle-2">
    ShellHub is a modern SSH server for remotely accessing Linux devices via CLI
    or web-based user interface, it avoids the inconveniences such as firewall
    and NAT when getting your IP address, automating the access process, and
    helping you easily, seamlessly and securely connect your device to the
    internet.
  </p>
</template>

<script lang="ts">
import { defineComponent, computed } from "vue";
import { useStore } from "../../store";

export default defineComponent({
  setup() {
    const store = useStore();

    const name = computed(
      () =>
        store.getters["auth/currentName"] || store.getters["auth/currentUser"]
    );
    return {
      name,
    };
  },
});
</script>
