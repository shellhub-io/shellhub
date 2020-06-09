<template>
  <div />
</template>

<script>
import { Terminal } from 'xterm';

export default {
  name: 'Terminal',

  mounted() {
    this.xterm = new Terminal({
      cursorBlink: true,
      fontFamily: 'monospace'
    });
  },

};
</script>
