<template>
  <CopyWarning copied-item="Command">
    <template #default="{ copyText }">
      <v-text-field
        :model-value="command"
        prefix="$"
        :label
        :hint
        data-test="copy-command-field"
        :persistent-hint
        :hide-details
        class="code text-white"
        readonly
        density="compact"
      >
        <template #append>
          <v-btn
            icon="mdi-content-copy"
            data-test="copy-button"
            color="primary"
            variant="flat"
            rounded
            size="small"
            @click="copyText(command)"
          />
        </template>
      </v-text-field>
    </template>
  </CopyWarning>
</template>

<script setup lang="ts">
import CopyWarning from "@/components/User/CopyWarning.vue";

interface Props {
  command: string;
  label?: string;
  hint?: string;
  persistentHint?: boolean;
  hideDetails?: boolean;
}

withDefaults(defineProps<Props>(), {
  label: undefined,
  hint: undefined,
  persistentHint: false,
  hideDetails: true,
});
</script>

<style scoped lang="scss">
.code ::v-deep(.v-field__field) {
  font-family: monospace;
  font-size: 85%;
  font-weight: normal;
}
</style>
