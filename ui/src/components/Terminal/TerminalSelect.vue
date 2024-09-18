<template>
  <v-select
    ref="terminalSelect"
    variant="outlined"
    :label="selectedToken === null ? 'Select Terminal' : 'Active Terminal'"
    :items="terminalTokens"
    :hide-details="true"
    item-title="uid"
    item-value="token"
    v-model="selectedToken"
  >
    <template #prepend-inner v-if="selectedToken !== null">
      <v-chip label icon color="primary" class="text-uppercase">
        <v-icon>mdi-console</v-icon>
      </v-chip>
    </template>
    <template #prepend-item>
      <v-list-subheader>
        All Open Sessions
      </v-list-subheader>
    </template>
    <template #item="{ item }">
      <v-list-item @click="handleTerminalClick(item.value)">
        <v-row cols="12">
          <v-col cols="10" class="d-flex justify-start align-center">
            <v-chip label icon color="primary" class="text-uppercase mr-2">
              <v-icon>mdi-console</v-icon>
            </v-chip>
            <span>{{ item.raw.uid.slice(0, 10) }}</span>
          </v-col>
          <v-col cols="2" class="d-flex justify-center align-center ma-0 pa-0 ">
            <v-btn icon variant="plain" @click.stop="closeTerminal(item.value)">
              <v-icon>mdi-close</v-icon>
            </v-btn>
          </v-col>
        </v-row>
      </v-list-item>
    </template>

    <template #append-item>
      <v-divider />
      <v-list-item class="mt-2 mb-0">
        <v-btn
          variant="flat"
          prepend-icon="mdi-plus-box"
          color="primary"
          class="ma-0"
          block
          @click="openQuickConnection()"
        >
          Quick Connection
        </v-btn>
      </v-list-item>
    </template>
  </v-select>
</template>

<script setup lang="ts">
import { computed, ref, watch, defineEmits } from "vue";
import { useRouter, useRoute } from "vue-router";
import type { VSelect } from "vuetify";
import { useStore } from "@/store";

const router = useRouter();
const route = useRoute();
const store = useStore();
const emit = defineEmits(["openQuickDialog"]);
const selectedToken = ref<string | null>(null);
const terminalSelect = ref<InstanceType<typeof VSelect> | null>(null);

const terminalTokens = computed(() => {
  const terminals = store.getters["terminals/getTerminal"];
  return Object.keys(terminals).map((token) => ({
    token,
    uid: terminals[token].uid,
  }));
});

const goToTerminal = (token) => {
  router.push({ name: "Connection", params: { token } });
};

const handleTerminalClick = (token: string) => {
  goToTerminal(token);
  if (terminalSelect.value) {
    terminalSelect.value.blur();
  }
};

const closeTerminal = (token: string) => {
  store.dispatch("terminals/removeTerminal", token);
  if (route.path === `/connection/${token}`) {
    router.push({ path: "/" });
  }
};

const openQuickConnection = () => emit("openQuickDialog");
const currentRoute = computed(() => route.params.token as string);
const currentToken = computed(() => store.getters["terminals/getTerminal"][currentRoute.value]);

watch(route, (newRoute) => {
  if (newRoute.name === "Connection") {
    selectedToken.value = currentToken.value.uid.slice(0, 10);
    return;
  }
  selectedToken.value = null;
});
</script>
