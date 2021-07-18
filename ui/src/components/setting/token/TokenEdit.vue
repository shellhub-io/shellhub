<template>
  <fragment>
    <v-tooltip bottom>
      <template #activator="{ on }">
        <v-icon
          v-on="on"
          @click="dialog = !dialog"
        >
          mdi-pencil
        </v-icon>
      </template>
      <span>Edit</span>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="400"
      @click:outside="close"
    >
      <v-card>
        <v-card-title
          class="headline grey lighten-2 text-center"
        >
          Edit Rule
        </v-card-title>

        <v-card-text>
          <v-text-field
            v-model="tokenLocal.tenant_id"
            disabled
          />

          <v-layout
            justify-space-between
            align-center
          >
            <v-flex>
              <v-card :elevation="0">
                <v-card-text class="v-label theme--light pl-0">
                  Permission read only
                </v-card-text>
              </v-card>
            </v-flex>

            <v-flex
              xs2
            >
              <v-card
                :elevation="0"
              >
                <v-switch
                  v-model="tokenLocal.read_only"
                />
              </v-card>
            </v-flex>
          </v-layout>
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn
            text
            @click="close"
          >
            Cancel
          </v-btn>

          <v-btn
            text
            @click="edit()"
          >
            Edit
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

export default {
  name: 'DeviceAdd',

  props: {
    token: {
      type: Boolean,
      required: true,
    },
  },

  data() {
    return {
      dialog: false,
    };
  },

  created() {
    this.tokenLocal = { ...this.token };
  },

  methods: {
    async edit() {
      try {
        await this.$store.dispatch('tokens/put', this.tokenLocal);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.tokenEditing);
        this.update();
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.tokenEditing);
      }
    },

    update() {
      this.$emit('update');
      this.close();
    },

    close() {
      this.dialog = !this.dialog;
    },
  },
};

</script>
