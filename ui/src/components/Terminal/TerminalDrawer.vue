<template>
  <v-navigation-drawer
    theme="dark"
    v-model="showTerminalDrawer"
    absolute
    app
    class="bg-v-theme-surface"
    data-test="navigation-drawer"
    location="right"
  >
    <v-container>
      <v-row class="mb-2">
        <v-col>
          <h3>Themes:</h3>
        </v-col>
      </v-row>
      <v-card class="bg-v-theme-surface elevation-3">
        <v-virtual-scroll
          :items="availableThemes"
          height="500"
          class=""
          data-test="virtual-scroller"
        >
          <template #default="{ item }">
            <v-list-item :key="(item as ThemeItem).file" lines="two" @click="changeTheme((item as ThemeItem).file)">
              <v-row cols="12">
                <v-col cols="4" class="d-flex justify-end align-center">
                  <v-icon
                    :style="`background:${(item as ThemeItem).preview.background};border-radius:50%;`"
                    class="pa-4"
                    :color="(item as ThemeItem).preview.foreground"
                    :icon="(item as ThemeItem).dark ? 'mdi-moon-waning-crescent' : 'mdi-white-balance-sunny'"
                  />
                </v-col>
                <v-col>
                  <h4>{{ (item as ThemeItem).name }}</h4>
                </v-col>
              </v-row>
            </v-list-item>
          </template>

        </v-virtual-scroll>
      </v-card>
      <v-row class="mt-2">
        <v-col>
          <h3>Themes:</h3>
        </v-col>
      </v-row>
      <v-card class="bg-v-theme-surface elevation-3 mt-2">
        <v-card-text>
          <div class="d-flex align-center">
            <v-text-field v-model="fontSize">
              <template v-slot:append>
                <v-icon color="red" @click="increaseFontSize">
                  mdi-plus
                </v-icon>
              </template>
              <template v-slot:prepend>
                <v-icon color="green" @click="decreaseFontSize">
                  mdi-minus
                </v-icon>
              </template>
            </v-text-field>
          </div>
        </v-card-text>
      </v-card>
      <div class="pa-4" />
    </v-container>
  </v-navigation-drawer>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useRoute } from "vue-router";
import { useStore } from "@/store";

interface ThemeItem {
  file: string;
  preview: {
    background: string;
    foreground: string;
  };
  name: string;
  dark: boolean;
}
const store = useStore();
const route = useRoute();
const showTerminalDrawer = ref(false);
const token = computed(() => route.params.token as string);

const availableThemes = computed(() => store.getters["terminals/getThemes"]);

const fontSize = computed({
  get() {
    return store.getters["terminals/getTerminal"][token.value].xterm.options.fontSize;
  },
  set(newFontSize) {
    store.commit("terminals/setFontSize", { token: token.value, fontSize: newFontSize > 12 ? newFontSize : 12 });
  },
});

const changeTheme = async (theme: string) => {
  await store.dispatch("terminals/applyTheme", { token: token.value, themeName: theme });
};

const increaseFontSize = () => {
  store.dispatch("terminals/changeFontSize", { token: token.value, adjustment: 2 });
};

const decreaseFontSize = () => {
  store.dispatch("terminals/changeFontSize", { token: token.value, adjustment: -2 });
};

defineExpose({ showTerminalDrawer });
</script>
