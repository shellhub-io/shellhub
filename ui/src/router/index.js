import Vue from 'vue';
import Router from 'vue-router';
import Dashboard from '../views/Dashboard';
import store from '../store';

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
      path: '/devices',
      name: 'devices',
      component: () => import(/* webpackChunkName: 'devices' */ './../views/Devices.vue'),
      redirect: {
        name: 'listDevices',
      },
      children: [
        {
          path: '',
          name: 'listDevices',
          component: () => import('./../components/device/DeviceList.vue'),
        },
        {
          path: 'pending',
          name: 'pendingDevices',
          component: () => import('./../components/device/DevicePendingList.vue'),
        },
        {
          path: 'rejected',
          name: 'rejectedDevices',
          component: () => import('./../components/device/DeviceRejectedList.vue'),
        },
      ],
    },
    {
      path: '/device/:id',
      name: 'detailsDevice',
      component: () => import(/* webpackChunkName: 'details-device' */ './../views/DetailsDevice.vue'),
    },
    {
      path: '/sessions',
      name: 'sessions',
      component: () => import('./../views/Sessions.vue'),
    },
    {
      path: '/session/:id',
      name: 'detailsSession',
      component: () => import(/* webpackChunkName: 'details-session' */ './../views/DetailsSession.vue'),
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('./../views/Login.vue'),
    },
    {
      path: '/firewall/rules',
      name: 'firewalls',
      component: () => import('./../views/FirewallRules.vue'),
    },
    {
      path: '/sshkeys/public_keys',
      name: 'publicKeys',
      component: () => import('./../views/PublicKeys.vue'),
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import(/* webpackChunkName: 'settings' */ './../views/Settings.vue'),
      redirect: {
        name: 'profileSettings',
      },
      children: [
        {
          path: 'profile',
          name: 'profileSettings',
          component: () => import('./../components/setting/SettingProfile.vue'),
        },
        {
          path: 'security',
          name: 'securitySettings',
          component: () => import('./../components/setting/SettingSecurity.vue'),
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
  if (to.path !== '/login') {
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
  return next();
});

export default router;
