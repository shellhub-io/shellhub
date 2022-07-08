<template>
  <v-hover v-slot="{ isHovering, props }">
    <v-card class="pa-1 mb-2 border bg-v-theme-surface" v-bind="props" :elevation="isHovering ? 24 : 8" density="compact" variant="flat">
      <v-list-item :data-test="nameUseTest" lines="three">
        <div>
          <p class="text-overline mb-1" :title-test="title">
            {{ title }}
          </p>
          <v-list-item-title class="text-h5 mb-1" :stats-test="stats">
            {{ stats }}
          </v-list-item-title>
          <v-list-item-subtitle :content-test="content">
            {{ content }}
          </v-list-item-subtitle>
        </div>

        <v-spacer />

        <template v-slot:append>
          <div class="mt-4">
            <v-icon :icon="icon" size="x-large" color="white" :iconName-test="icon" />
          </div>
        </template>
      </v-list-item>

      <v-card-actions>
        <v-btn
          v-if="buttonName != 'Add Device'"
          :to="pathName"
          color="primary"
          variant="elevated"
          tabindex="0"
          :buttonName-test="buttonName"
          class="text-subtitle"
          size="small"
        >
          {{ buttonName }}
        </v-btn>

        <DeviceAdd v-else size="small"/>
      </v-card-actions>
    </v-card>
  </v-hover>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import DeviceAdd from "../Devices/DeviceAdd.vue";

export default defineComponent({
    props: {
        id: {
            type: Number,
            default: 0,
        },
        title: {
            type: String,
            default: "",
        },
        fieldObject: {
            type: String,
            default: "",
        },
        content: {
            type: String,
            default: "",
        },
        icon: {
            type: String,
            default: "",
        },
        buttonName: {
            type: String,
            default: "",
        },
        pathName: {
            type: String,
            default: "",
        },
        nameUseTest: {
            type: String,
            default: "",
        },
        stats: {
            type: Number,
            default: 0,
        },
    },
    components: { DeviceAdd }
});
</script>

<style lang="scss">
  .border {
    border: thin solid rgba(255, 255, 255, 0.12);
  }
</style>
