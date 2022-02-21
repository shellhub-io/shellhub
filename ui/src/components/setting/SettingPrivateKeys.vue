<template>
  <v-form>
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

              <span @click="privateKeyFormDialogAddShow = !privateKeyFormDialogAddShow">
                <PrivateKeyFormDialog
                  :create-key="true"
                  action="private"
                  :show.sync="privateKeyFormDialogAddShow"
                  data-test="privateKeyFormDialogFirst-component"
                />
              </span>
            </div>

            <v-data-table
              :headers="headers"
              :items="getListPrivateKeys"
              data-test="dataTable-field"
              :server-items-length="getNumberPrivateKeys"
              hide-default-footer
            >
              <template #[`item.name`]="{ item }">
                {{ item.name }}
              </template>

              <template #[`item.data`]="{ item }">
                {{ convertToFingerprint(item.data) }}
              </template>

              <template #[`item.actions`]="{ item }">
                <v-menu
                  :ref="'menu'+getListPrivateKeys.indexOf(item)"
                  offset-y
                >
                  <template #activator="{ on, attrs }">
                    <v-chip
                      color="transparent"
                      v-on="on"
                    >
                      <v-icon
                        small
                        class="icons"
                        v-bind="attrs"
                        v-on="on"
                      >
                        mdi-dots-horizontal
                      </v-icon>
                    </v-chip>
                  </template>

                  <v-card>
                    <v-list-item
                      @click="showPrivateKeyFormDialog(getListPrivateKeys.indexOf(item))"
                    >
                      <PrivateKeyFormDialog
                        :key-object="item"
                        :create-key="false"
                        action="private"
                        :show.sync="privateKeyFormDialogShow[getListPrivateKeys.indexOf(item)]"
                        data-test="privateKeyFormDialogSecond-component"
                      />
                    </v-list-item>

                    <v-list-item @click="showPrivateKeyDelete(getListPrivateKeys.indexOf(item))">
                      <PrivateKeyDelete
                        :fingerprint="item.data"
                        action="private"
                        :show.sync="privateKeyDeleteShow[getListPrivateKeys.indexOf(item)]"
                        data-test="privateKeyDelete-component"
                      />
                    </v-list-item>
                  </v-card>
                </v-menu>
              </template>
            </v-data-table>
          </v-card>
        </v-col>
      </v-row>
    </v-container>

    <v-dialog
      v-model="dialog"
      persistent
      width="500"
    >
      <v-card>
        <v-card-title class="headline primary">
          Privacy Policy
        </v-card-title>

        <v-card-text
          class="mt-4"
        >
          The private key is never submitted to ShellHub, it gets stored in your browser’s
          local storage, only the public key gets uploaded and stored by ShellHub.
        </v-card-text>

        <v-divider />
        <v-card-actions
          class="px-6"
        >
          <v-checkbox
            v-model="privatekeyPrivacyPolicy"
            label="Never show this again"
          />
          <v-spacer />
          <v-btn
            color="primary"
            text
            data-test="gotIt-btn"
            @click="accept"
          >
            Got it
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-form>
</template>

<script>

import PrivateKeyFormDialog from '@/components/public_key/KeyFormDialog';
import PrivateKeyDelete from '@/components/private_key/PrivateKeyDelete';

import { parsePrivateKey } from '@/sshpk';

export default {
  name: 'SettingPrivateKeysComponent',

  components: {
    PrivateKeyFormDialog,
    PrivateKeyDelete,
  },

  data() {
    return {
      pagination: {},
      dialog: true,
      privatekeyPrivacyPolicy: false,
      privateKeyFormDialogAddShow: false,
      privateKeyFormDialogShow: [],
      privateKeyDeleteShow: [],

      headers: [
        {
          text: 'Name',
          value: 'name',
          align: 'center',
        },
        {
          text: 'Fingerprint',
          value: 'data',
          align: 'center',
        },
        {
          text: 'Actions',
          value: 'actions',
          align: 'center',
        },
      ],
    };
  },

  computed: {
    getListPrivateKeys() {
      return this.$store.getters['privatekeys/list'];
    },

    getNumberPrivateKeys() {
      return this.$store.getters['privatekeys/getNumberPrivateKeys'];
    },
  },

  created() {
    this.dialog = !(localStorage.getItem('privatekeyPrivacyPolicy') === 'true');
    this.setArrays();
  },

  methods: {
    convertToFingerprint(privateKey) {
      return parsePrivateKey(privateKey).fingerprint('md5');
    },

    accept() {
      localStorage.setItem('privatekeyPrivacyPolicy', this.privatekeyPrivacyPolicy);
      this.dialog = false;
    },

    showPrivateKeyFormDialog(index) {
      this.privateKeyFormDialogShow[index] = this.privateKeyFormDialogShow[index] === undefined
        ? true : !this.privateKeyFormDialogShow[index];
      this.$set(this.privateKeyFormDialogShow, index, this.privateKeyFormDialogShow[index]);

      this.closeMenu(index);
    },

    showPrivateKeyDelete(index) {
      this.privateKeyDeleteShow[index] = this.privateKeyDeleteShow[index] === undefined
        ? true : !this.privateKeyDeleteShow[index];
      this.$set(this.privateKeyDeleteShow, index, this.privateKeyDeleteShow[index]);

      this.closeMenu(index);
    },

    setArrays() {
      const numberPrivateKey = this.getListPrivateKeys.length;

      if (numberPrivateKey > 0) {
        this.privateKeyFormDialogShow = new Array(numberPrivateKey).fill(false);
        this.privateKeyDeleteShow = new Array(numberPrivateKey).fill(false);
      }
    },

    closeMenu(index) {
      this.$refs[`menu${index}`].isActive = false;
    },
  },
};
</script>
