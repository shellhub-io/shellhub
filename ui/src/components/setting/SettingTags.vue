<template>
  <fragment>
    <v-container>
      <v-row
        align="center"
        justify="center"
        class="mt-4 mb-4"
      >
        <v-col
          sm="8"
        >
          <v-card class="pb-0 elevation-0">
            <div class="d-flex pa-0 align-center">
              <v-spacer />
              <v-spacer />
            </div>

            <div class="text-justify">
              Tags are a keyword or term assigned to information, as if it were attached to
              someone or something, for device identification purposes.

              In this case, it is possible to perform some actions, including editing,
              changing the tag name on all connected devices.

              Following the same logic, when you delete a tag, it is deleted from all
              related devices.
            </div>
          </v-card>

          <TagList data-test="tagList-component" />
        </v-col>
      </v-row>
    </v-container>
  </fragment>
</template>

<script>

import TagList from '@/components/setting/tag/TagList';

export default {
  name: 'SettingsTagsComponent',

  components: {
    TagList,
  },
};

</script>
