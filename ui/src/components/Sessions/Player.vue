<template>
  <div class="wrapper" v-if="logs" ref="wrapper" />
</template>

<script setup lang="ts">
import * as AsciinemaPlayer from "asciinema-player";
import { onMounted, ref } from "vue";

const { logs } = defineProps<{
  logs: string | null;
}>();

const wrapper = ref<HTMLDivElement | null>(null);
const player = ref<AsciinemaPlayer.AsciinemaPlayer | null>(null);

onMounted(() => {
  console.log(logs);
  player.value = AsciinemaPlayer.create({ data: logs }, wrapper.value, { fit: "height" });
});
</script>

<style lang="scss" scoped>
.wrapper {
  height: 100%;
  overflow: scroll-y;
}
</style>
