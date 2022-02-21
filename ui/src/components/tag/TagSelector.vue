<template>
  <fragment>
    <div class="mr-4">
      <v-menu
        offset-y
        :close-on-content-click="false"
      >
        <template #activator="{ on, attrs }">
          <v-badge
            bordered
            color="primary"
            :content="selectedTags.length"
            :value="selectedTags.length"
            overlap
          >
            <v-btn
              color="primary"
              v-bind="attrs"
              data-test="tags-btn"
              outlined
              :disabled="getListTags.length==0"
              v-on="on"
              @click="getTags"
            >
              Tags
              <v-icon right>
                mdi-chevron-down
              </v-icon>
            </v-btn>
          </v-badge>
        </template>

        <v-list>
          <v-list-item-group
            v-model="selectedTags"
            multiple
          >
            <template v-for="(item, i) in getListTags">
              <v-list-item
                :key="`item-${i}`"
                :value="item"
                :data-test="item + '-item'"
              >
                <template #default="{ active }">
                  <v-list-item-action>
                    <v-checkbox :input-value="active" />
                  </v-list-item-action>

                  <v-list-item-content>
                    <v-list-item-title
                      :data-test="item + '-title'"
                      v-text="item"
                    />
                  </v-list-item-content>
                </template>
              </v-list-item>
            </template>
          </v-list-item-group>
        </v-list>
      </v-menu>
    </div>
  </fragment>
</template>

<script>

export default {
  name: 'TagSelector',

  computed: {
    getListTags() {
      return this.$store.getters['tags/list'];
    },

    selectedTags: {
      get() {
        return this.$store.getters['tags/selected'];
      },

      set(item) {
        this.$store.dispatch('tags/setSelected', item);
      },
    },
  },

  watch: {
    selectedTags(item) {
      this.getDevices(item);
    },
  },

  created() {
    this.getTags();
  },

  methods: {
    async getTags() {
      await this.$store.dispatch('tags/fetch');
    },

    async getDevices(item) {
      let encodedFilter = null;

      if (item.length > 0) {
        const filter = [{ type: 'property', params: { name: 'tags', operator: 'contains', value: item } }];
        encodedFilter = btoa(JSON.stringify(filter));
      }
      await this.$store.dispatch('devices/setFilter', encodedFilter);

      try {
        this.$store.dispatch('devices/refresh');
      } catch (error) {
        if (error.response.status === 403) {
          this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorDefault');
        }
      }
    },
  },
};
</script>
