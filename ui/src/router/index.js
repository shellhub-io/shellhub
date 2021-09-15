import Vue from 'vue';
import Router from 'vue-router';
import Dashboard from '@/views/Dashboard';
import store from '@/store';

Vue.use(Router);

const router = new Router({
  mode: 'history',
  base: process.env.BASE_URL,
  routes: [
    {
      path: '/',
      name: 'dashboard',
      component: Dashboard,
    },
    {
      path: '/sign-up',
      name: 'signUp',
      component: () => import('@/views/SignUp'),
    },
    {
      path: '/validation-account',
      name: 'validationAccount',
      component: () => import('@/views/ValidationAccount'),
    },
    {
      path: '/forgot-password',
      name: 'forgotPassword',
      component: () => import('@/views/ForgotPassword'),
    },
    {
      path: '/update-password',
      name: 'updatePassword',
      component: () => import('@/views/UpdatePassword'),
    },
    {
      path: '/devices',
      name: 'devices',
      component: () => import(/* webpackChunkName: 'devices' */ '@/views/Devices'),
      redirect: {
        name: 'listDevices',
      },
      children: [
        {
          path: '',
          name: 'listDevices',
          component: () => import('@/components/device/DeviceList'),
        },
        {
          path: 'pending',
          name: 'pendingDevices',
          component: () => import('@/components/device/DevicePendingList'),
        },
        {
          path: 'rejected',
          name: 'rejectedDevices',
          component: () => import('@/components/device/DeviceRejectedList'),
        },
      ],
    },
    {
      path: '/device/:id',
      name: 'detailsDevice',
      component: () => import(/* webpackChunkName: 'details-device' */ '@/views/DetailsDevice'),
    },
    {
      path: '/sessions',
      name: 'sessions',
      component: () => import(/* webpackChunkName: 'sessions' */ '@/views/Sessions'),
      redirect: {
        name: 'listSessions',
      },
      children: [
        {
          path: '',
          name: 'listSessions',
          component: () => import('@/components/session/SessionList'),
        },
      ],
    },
    {
      path: '/session/:id',
      name: 'detailsSession',
      component: () => import(/* webpackChunkName: 'details-session' */ '@/views/DetailsSession'),
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/Login'),
    },
    {
      path: '/firewall/rules',
      name: 'firewalls',
      component: () => import(/* webpackChunkName: 'firewalls' */ '@/views/FirewallRules'),
      redirect: {
        name: 'listFirewalls',
      },
      children: [
        {
          path: '',
          name: 'listFirewalls',
          component: () => import('@/components/firewall_rule/FirewallRuleList'),
        },
      ],
    },
    {
      path: '/sshkeys/public-keys',
      name: 'publicKeys',
      component: () => import(/* webpackChunkName: 'publickeys' */'@/views/PublicKeys'),
      redirect: {
        name: 'listPublickeys',
      },
      children: [
        {
          path: '',
          name: 'listPublickeys',
          component: () => import('@/components/public_key/PublicKeyList'),
        },
      ],
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import(/* webpackChunkName: 'settings' */ '@/views/Settings'),
      redirect: {
        name: 'profileSettings',
      },
      children: [
        {
          path: 'profile',
          name: 'profileSettings',
          component: () => import('@/components/setting/SettingProfile'),
        },
        {
          path: 'namespace-manager',
          name: 'namespaceSettings',
          component: () => import('@/components/setting/SettingNamespace'),
        },
        {
          path: 'private-keys',
          name: 'privateKeysSettings',
          component: () => import('@/components/setting/SettingPrivateKeys'),
        },
        {
          path: 'tags',
          name: 'tagsSettings',
          component: () => import('@/components/setting/SettingTags'),
        },
      ],
    },
    {
      path: '*',
      name: 'NotFound',
      component: Dashboard,
      redirect: () => {
        localStorage.setItem('flag', true);
        return '/';
      },
    },
  ],
});

router.beforeEach((to, from, next) => {
  if ((to.path !== '/login' && to.path !== '/sign-up') && to.path !== '/forgot-password' && to.path !== '/validation-account' && to.path !== '/update-password') {
    if (store.getters['auth/isLoggedIn']) {
      return next();
    }
    return next(`/login?redirect=${to.path}`);
  }
  if (store.getters['auth/isLoggedIn']) {
    if (to.path === '/login' && to.query.token) {
      return next();
    }
    return next('/');
  }
  if (to.path === '/sign-up' && process.env.VUE_APP_SHELLHUB_CLOUD === 'false') {
    return next('/');
  }
  return next();
});

export default router;
