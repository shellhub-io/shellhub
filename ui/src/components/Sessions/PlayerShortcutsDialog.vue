<template>
  <v-dialog
    :fullscreen="false"
    v-model="showDialog"
  >
    <v-card title="Keyboard Shortcuts">
      <v-card-text>
        <div class="shortcut"><v-kbd>space</v-kbd>: pause / resume</div>
        <div class="shortcut"><v-kbd>f</v-kbd>: toggle fullscreen mode</div>
        <div class="shortcut"><v-kbd>←</v-kbd> / <v-kbd>→</v-kbd>: rewind / fast-forward by 5 seconds</div>
        <div class="shortcut"><v-kbd>Shift</v-kbd> + <v-kbd>←</v-kbd> / <v-kbd>→</v-kbd>: rewind / fast-forward by 10%</div>
        <div class="shortcut"><v-kbd>0</v-kbd>, <v-kbd>1</v-kbd>, <v-kbd>2</v-kbd> ... <v-kbd>9</v-kbd>: jump to 0%, 10%, 20% ... 90%</div>
        <div class="shortcut"><v-kbd>,</v-kbd>/<v-kbd>.</v-kbd>: step back / forward, a frame at a time (only when paused)</div>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
const showDialog = defineModel<boolean>("showDialog");
</script>

<style lang="scss" scoped>
.shortcut {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  margin-bottom: 1rem;

  .v-kbd {
    padding: .2rem .5rem;
    font-weight: 700;
  }
}
</style>
