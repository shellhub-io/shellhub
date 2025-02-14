<template>
  <v-hover v-slot="{ isHovering, props }">
    <v-card class="pa-2 mb-2" v-bind="props" :elevation="isHovering ? 24 : 8">
      <v-list-item :data-test="nameUseTest" lines="three">
        <div>
          <p class="text-overline mb-4" :title-test="title">
            {{ title }}
          </p>
          <v-list-item-title class="text-h4 mb-1" :stats-test="stats">
            {{ stats }}
          </v-list-item-title>
          <v-list-item-subtitle :content-test="content">
            {{ content }}
          </v-list-item-subtitle>
        </div>

        <v-spacer />

        <template v-slot:append>
          <div class="mt-4">
            <v-icon :icon="icon" size="x-large" :iconName-test="icon" />
          </div>
        </template>
      </v-list-item>

      <v-card-actions>
        <v-btn :to="pathName" color="dark" variant="text" tabindex="0" :buttonName-test="buttonName">
          {{ buttonName }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-hover>
</template>

<script lang="ts">
import { defineComponent } from "vue";

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
});
</script>
