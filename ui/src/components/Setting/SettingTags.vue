<template>
  <v-container fluid>
    <v-card
      variant="flat"
      class="bg-transparent"
      data-test="tags-settings-card"
    >
      <v-card-item>
        <v-list-item
          class="pa-0 ma-0 mb-2"
          data-test="tags-settings-header"
        >
          <template v-slot:title>
            <h1>Tags</h1>
          </template>
          <template v-slot:subtitle>
            <span data-test="tags-settings-subtitle">Manage your device and connector tags</span>
          </template>
        </v-list-item>
      </v-card-item>
      <TagList />
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import TagList from "../Tags/TagList.vue";
</script>
