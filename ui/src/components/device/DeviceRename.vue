<template>
  <fragment>
    <v-tooltip bottom>
      <template v-slot:activator="{ on }">
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
      max-width="450"
      @click:outside="cancel"
    >
      <v-card>
        <v-card-title class="headline grey lighten-2 text-center">
          Rename Device
        </v-card-title>
        <ValidationObserver
          ref="obs"
          v-slot="{ validated, passes }"
        >
          <v-card-text class="caption mb-0">
            <ValidationProvider
              v-slot="{ errors }"
              name="Hostname"
              rules="required|rfc1123"
            >
              <v-text-field
                v-model="editName"
                label="Hostname"
                :error-messages="errors"
                require
                messages="Examples: (foobar, foo-bar.ba-z.qux, foo.example.com, 127.0.0.1)"
              />
            </ValidationProvider>
          </v-card-text>
          <v-card-actions>
            <v-spacer />
            <v-btn
              text
              @click="cancel"
            >
              Close
            </v-btn>
            <v-btn
              color="primary"
              text
              @click="passes(edit)"
            >
              Rename
            </v-btn>
          </v-card-actions>
        </ValidationObserver>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

import {
  ValidationObserver,
  ValidationProvider,
} from 'vee-validate';

export default {
  name: 'DeviceRename',

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  props: {
    name: {
      type: String,
      required: true,
    },
    uid: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      dialog: false,
      invalid: false,
      editName: '',
    };
  },

  computed: {
    device: {
      get() {
        return {
          name: this.name,
          uid: this.uid,
        };
      },
    },
  },

  created() {
    this.editName = this.device.name;
  },

  updated() {
    this.editName = this.device.name;
  },

  methods: {
    cancel() {
      this.dialog = false;
      this.invalid = false;
      this.editName = '';
    },

    edit() {
      this.$store.dispatch('devices/rename', {
        uid: this.device.uid,
        name: this.editName,
      });
      this.dialog = false;
      this.$emit('newHostname', this.editName);
      this.editName = '';
    },
  },
};

</script>
