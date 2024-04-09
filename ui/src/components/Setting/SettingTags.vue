<template>
  <v-container>
    <v-row align="center" justify="center" class="mt-4 mb-4">
      <v-col sm="8">
        <v-card class="pb-0 elevation-0">
          <div class="d-flex pa-0 align-center">
            <v-spacer />
            <v-spacer />
          </div>
        </v-card>

        <TagList data-test="tagList-component" />
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import TagList from "../Tags/TagList.vue";
</script>
