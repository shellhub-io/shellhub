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

<script lang="ts">
import { defineComponent } from "vue";
import TagList from "../Tags/TagList.vue";

export default defineComponent({
  setup() {
    return {
      //
    };
  },
  components: { TagList },
});
</script>
